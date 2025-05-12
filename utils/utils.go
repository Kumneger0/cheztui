package utils

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"slices"

	"github.com/charmbracelet/bubbles/list"
	"github.com/kumneger0/chez-tui/helpers.go"
)

func GetAbsolutePath(path string, baseDir string) (string, error) {
	return filepath.Join(baseDir, path), nil
}

func GetFilesFromSpecificPath(path string) ([]list.Item, error) {
	dirEntery, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var allFilesInCurrentDir []list.Item
	output, _ := exec.Command("chezmoi", "managed", path).Output()
	mangedFilesAndDirsInCurrentDir := strings.Split(string(output), "/n")

	for _, v := range dirEntery {
		fileEntery := helpers.FileEntry{Name: v.Name(), Path: v.Name(), IsManaged: isCurrentFileOrDirManaged(mangedFilesAndDirsInCurrentDir, v.Name()), IsDir: v.IsDir()}
		allFilesInCurrentDir = append(allFilesInCurrentDir, fileEntery)
	}

	userHomeDir, _ := os.UserHomeDir()

	if path == userHomeDir {
		return allFilesInCurrentDir, nil
	}

	allFilesInCurrentDir = append([]list.Item{helpers.FileEntry{
		Name:       "..",
		Path:       "..",
		IsManaged:  false,
		IsDir:      true, 
	}}, allFilesInCurrentDir...) 

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

func isCurrentFileOrDirManaged(managedFilesAndDirsInCurrentDir []string, fileNameOrDirName string) bool {
	var isManaged bool
	for _, v := range managedFilesAndDirsInCurrentDir {
		dirs := strings.Split(v, "/")
		if slices.Contains(dirs, fileNameOrDirName) {
			isManaged = true
			break
		}

	}
	return isManaged
}

func GoTernary(condition bool, trueVal, falseVal string) string {
	if condition {
		return trueVal
	}
	return falseVal
}
