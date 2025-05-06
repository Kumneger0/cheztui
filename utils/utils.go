package utils

import (
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/list"
	"github.com/kumneger0/chez-tui/helpers.go"
)

func GetAbsolutePath(path string) (string, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userHomeDir, path), nil
}

func GetFilesFromSpecificPath(path string) ([]list.Item, error) {
	dirEntery, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var allFilesInCurrentDir []list.Item

	for _, v := range dirEntery {
		fileEntery := helpers.FileEntry{Name: v.Name(), Path: v.Name(), IsManaged: false, IsDir: v.IsDir()}
		allFilesInCurrentDir = append(allFilesInCurrentDir, fileEntery)
	}
	return allFilesInCurrentDir, nil
}

func FindFileProperty(path string, files []list.Item) helpers.FileEntry {
	var fileProperty helpers.FileEntry
	for _, v := range files {
		if v.(helpers.FileEntry).Path == path {
			fileProperty = v.(helpers.FileEntry)
			break
		}
	}
	return fileProperty 
}