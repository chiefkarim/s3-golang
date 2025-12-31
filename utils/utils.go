package utils

import (
	"crypto/rand"
	"encoding/base64"
	"path/filepath"
)

func MakeFilePath(root, extention string) (string, error) {
	randomName := make([]byte, 32)
	_, err := rand.Read(randomName)
	filePath := filepath.Join(root, base64.RawURLEncoding.EncodeToString(randomName)+"."+extention)
	return filePath, err
}
