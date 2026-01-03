package utils

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func MakeFilePath(root, extention string) (string, error) {
	randomName := make([]byte, 32)
	_, err := rand.Read(randomName)
	filePath := filepath.Join(root, base64.RawURLEncoding.EncodeToString(randomName)+"."+extention)
	return filePath, err
}

type VideoMetaData struct {
	Streams []struct {
		CodecType string `json:"codec_type"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
	} `json:"streams"`
}

func GetVideoAspectRatio(width, height int) string {
	tolerance := 0.02

	ratio := float64(width) / float64(height)
	switch {
	case math.Abs(ratio-(16.0/9.0)) < tolerance:
		return "16:9"
	case math.Abs(ratio-(9.0/16.0)) < tolerance:
		return "9:16"
	default:
		return "other"
	}
}

func GetVideoWidthAndHeight(filePath string) (int, int, error) {
	command := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	var output bytes.Buffer
	command.Stdout = &output
	var errors bytes.Buffer
	command.Stderr = &errors
	err := command.Run()
	if err != nil {
		log.Print("\n", err)
		log.Print("\n", errors.String(), "\n")
		return 0, 0, err
	}

	var videoMetaData VideoMetaData
	json.Unmarshal(output.Bytes(), &videoMetaData)
	var width int
	var height int
	for _, s := range videoMetaData.Streams {
		if s.CodecType == "video" {
			width = s.Width
			height = s.Height
		}
	}

	return width, height, err
}

func ProcessVideoForFastStart(filePath string) (string, error) {
	output := strings.ReplaceAll(filePath, "mp4", "") + ".processing.mp4"
	command := exec.Command("ffmpeg",
		"-i",
		filePath,
		"-c",
		"copy",
		"-movflags",
		"faststart",
		"-f",
		"mp4",
		output,
	)
	var errors bytes.Buffer
	command.Stderr = &errors

	err := command.Run()
	if err != nil {
		fmt.Print("\n", err)
		fmt.Print("\n", errors.String(), "\n")
		return "", err
	}

	return output, nil
}

func GenratePresignedUrl(s3Client *s3.Client, bucket, key string, expireTime time.Duration) (string, error) {
	presignedClient := s3.NewPresignClient(s3Client)
	presignedUrl, err := presignedClient.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Key:    &key,
		Bucket: &bucket,
	}, s3.WithPresignExpires(expireTime))
	return presignedUrl.URL, err
}
