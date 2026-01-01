package utils

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"os/exec"
	"path/filepath"
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
		fmt.Print("\n", err)
		fmt.Print("\n", errors.String(), "\n")
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
