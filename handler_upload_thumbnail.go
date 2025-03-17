package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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

	// Set maxMemory to 10MB (10 bitshifted 20 times to left)
	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)

	fileData, fileHeader, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "could not get file data/header", err)
		return
	}
	mediaType := fileHeader.Header.Get("Content-type")
	mimeType, _, err := mime.ParseMediaType(mediaType)

	if mimeType != "image/jpeg" || mimeType != "image/png" {
		respondWithError(w, http.StatusUnsupportedMediaType, "thumbnail files must be jpeg/png", err)
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

	imageExt, err := getImageFileType(mediaType)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "image filetype not found", err)
		return
	}
	thumbnailFilepath := filepath.Join(cfg.assetsRoot, videoIDString+imageExt)
	newFile, err := os.Create(thumbnailFilepath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to create save image file", err)
		return
	}
	defer newFile.Close()

	_, err = io.Copy(newFile, fileData)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to save image file", err)
		return
	}
	thumbnailURL := fmt.Sprintf("http://localhost:%v/assets/%v%v", cfg.port, videoIDString, imageExt)

	videoData.ThumbnailURL = &thumbnailURL

	err = cfg.db.UpdateVideo(videoData)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "could not update video db", err)
		return
	}

	respondWithJSON(w, http.StatusOK, videoData)
}

func getImageFileType(s string) (string, error) {
	if strings.HasPrefix(s, "image/") {
		parts := strings.Split(s, "/")
		if len(parts) == 2 {
			return "." + parts[1], nil
		}
	}
	return "", fmt.Errorf("no image filetype found")
}
