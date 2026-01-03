package main

import (
	"errors"
	"strings"
	"time"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/utils"
)

func (cfg *apiConfig) dbVideoToSigneVideo(video database.Video) (database.Video, error) {
	if video.VideoURL == nil {
		return database.Video{}, errors.New("invalid video url. must contain s3bucket url at the beginning")
	}

	urldata := strings.Split(*video.VideoURL, ",")
	if len(urldata) != 2 {
		return database.Video{}, errors.New("invalid video url. must contain s3bucket url at the beginning")
	}
	bucket, key := urldata[0], urldata[1]

	presignedURL, err := utils.GenratePresignedUrl(cfg.s3Client, bucket, key, time.Minute*5)

	video.VideoURL = &presignedURL
	return video, err
}
