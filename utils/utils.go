package utils

import (
	"os"
	"path/filepath"
)

func GetAbsolutePath(path string) (string, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userHomeDir, path), nil
}
