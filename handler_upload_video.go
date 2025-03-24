package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {

	r.Body = http.MaxBytesReader(w, r.Body, int64(1<<30))

	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	videoData, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "could not get video data from db", err)
		return
	}

	if videoData.UserID != userID {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	fileData, fileHeader, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "could not get video file data/header", err)
		return
	}
	mediaType := fileHeader.Header.Get("Content-type")
	mimeType, _, err := mime.ParseMediaType(mediaType)
	fmt.Println(mimeType)
	if mimeType != "video/mp4" {
		respondWithError(w, http.StatusUnsupportedMediaType, "video files must be video/mp4", err)
		return
	}

	tmpFile, err := os.CreateTemp("", "tubely-upload.mp4")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not create temporary file", err)
		return
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()
	io.Copy(tmpFile, fileData)

	// Reset the tempFile's file pointer to the beginning to read the file again
	tmpFile.Seek(0, io.SeekStart)

	aspectRatio, err := getVideoAspectRatio(tmpFile.Name())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not get aspect ratio of file", err)
	}

	randByteSlice := make([]byte, 32)
	rand.Read(randByteSlice)
	randString := base64.RawURLEncoding.EncodeToString(randByteSlice)

	var prefix string
	switch aspectRatio {
	case "16:9":
		prefix = "landscape"
	case "9:16":
		prefix = "portrait"
	default:
		prefix = "other"
	}

	filename := prefix + "/" + randString + ".mp4"

	_, err = cfg.s3Client.PutObject(r.Context(), &s3.PutObjectInput{
		Bucket:      &cfg.s3Bucket,
		Key:         &filename,
		Body:        tmpFile,
		ContentType: &mimeType,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to upload to S3", err)
		return
	}

	videoURL := fmt.Sprintf("https://%v.s3.%v.amazonaws.com/%v", cfg.s3Bucket, cfg.s3Region, filename)
	videoData.VideoURL = &videoURL

	err = cfg.db.UpdateVideo(videoData)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "could not update video db", err)
		return
	}

	respondWithJSON(w, http.StatusOK, videoData)
}
