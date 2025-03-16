package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
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

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here
	// Set maxMemory to 10MB (10 bitshifted 20 times to left)
	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)

	fileData, fileHeader, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "could not get file data/header", err)
	}
	mediaType := fileHeader.Header.Get("Content-type")
	imageData, err := io.ReadAll(fileData)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "could not read file data", err)
	}

	videoData, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "could not get video data from db", err)
	}

	if videoData.UserID != userID {
		w.WriteHeader(http.StatusUnauthorized)
	}

	newThumbnail := thumbnail{
		imageData,
		mediaType,
	}

	videoThumbnails[videoData.ID] = newThumbnail

	thumbnailURL := fmt.Sprintf("http://localhost:%v/api/thumbnails/%v", cfg.port, videoIDString)
	videoData.ThumbnailURL = &thumbnailURL
	err = cfg.db.UpdateVideo(videoData)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "could not update video db", err)
	}

	respondWithJSON(w, http.StatusOK, videoData)
}
