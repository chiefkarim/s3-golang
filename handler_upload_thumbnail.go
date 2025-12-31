package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
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

	var maxMemory int64 = 10 << 20
	r.ParseMultipartForm(maxMemory)
	media, headers, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Please provide thumbnail image", err)
		return
	}

	mediaType := headers.Header.Get("Content-type")
	if mediaType == "" || 2 != len(strings.Split(mediaType, "/")) {
		respondWithError(w, http.StatusBadRequest, "Please provide correct content type header", err)
		return
	}

	mediaType = strings.Split(mediaType, "/")[1]
	if mediaType != "png" && mediaType != "jpeg" {
		respondWithError(w, http.StatusBadRequest, "Please provide images of jpeg or png format", err)
		return
	}

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Video not found", err)
		return
	}

	if video.CreateVideoParams.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	filename := make([]byte, 32)
	rand.Read(filename)
	imagePath := filepath.Join(cfg.assetsRoot, base64.RawURLEncoding.EncodeToString(filename)+"."+mediaType)
	imageFile, err := os.Create(imagePath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal server error", err)
		return
	}
	io.Copy(imageFile, media)

	thumbnailUrl := fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, imagePath)
	video.ThumbnailURL = &thumbnailUrl

	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Video not found", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}
