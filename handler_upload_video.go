package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/utils"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	var maxUploadSize int64 = 1 << 30 // 1GB
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	err := r.ParseMultipartForm(maxUploadSize)
	stringVideoID := r.PathValue("videoID")
	videoID, err := uuid.Parse(stringVideoID)
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

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Not Found!", err)
		return
	}
	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized!", err)
		return
	}

	media, mediaHeader, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Please upload a video!", err)
		return
	}
	defer media.Close()

	mediaType, _, err := mime.ParseMediaType(mediaHeader.Header.Get("Content-type"))

	if err != nil || mediaType != "video/mp4" {
		respondWithError(w, http.StatusBadRequest, "Please upload a valid video!", err)
		return
	}

	tempVideo, err := os.CreateTemp("", "tubely-upload")
	defer os.Remove(tempVideo.Name())
	defer tempVideo.Close()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal server error", err)
		return
	}

	io.Copy(tempVideo, media)
	tempVideo.Seek(0, io.SeekStart)
	extention := strings.Split(mediaType, "/")[1]

	proccessedVideoPath, err := utils.ProcessVideoForFastStart(tempVideo.Name())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal server error", err)
		return
	}
	proccessedVideo, err := os.Open(proccessedVideoPath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal server error", err)
		return
	}
	defer proccessedVideo.Close()
	defer os.Remove(proccessedVideoPath)

	width, height, err := utils.GetVideoWidthAndHeight(proccessedVideo.Name())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal server error", err)
		return
	}
	aspectRatio := utils.GetVideoAspectRatio(width, height)

	videoKey, err := utils.MakeFilePath(aspectRatioToFolderName(aspectRatio), extention)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal server error", err)
		return
	}

	videoMetaData := s3.PutObjectInput{
		Key:         &videoKey,
		Bucket:      &cfg.s3Bucket,
		Body:        proccessedVideo,
		ContentType: &mediaType,
	}
	_, err = cfg.s3Client.PutObject(r.Context(), &videoMetaData)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal server error", err)
		return
	}
	videoUrl := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.s3Bucket, cfg.s3Region, videoKey)
	video.VideoURL = &videoUrl
	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal server error", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, "Video successfully uploaed")
}

func aspectRatioToFolderName(aspectRatio string) string {
	switch aspectRatio {
	case "16:9":
		return "landscape"
	case "9:16":
		return "portrait"
	default:
		return "other"
	}
}
