package chezmoi

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kumneger0/chez-tui/helpers.go"
	"github.com/kumneger0/chez-tui/utils"
)

type tempDir struct {
	path string
}

type AltarnateScreeenExec struct{ error }

const Command = "chezmoi"

func IsChezmoiInstalled() bool {
	_, err := exec.LookPath(Command)
	return err == nil
}

func getChezmoiSourceDir() (string, error) {
	userHomeDir, _ := os.UserHomeDir()
	return utils.GetAbsolutePath(".local/share/chezmoi", userHomeDir)
}

func IsChezmoiInitialized() bool {
	chezmoiSourceDir, err := getChezmoiSourceDir()
	//TODO:migrate to use better-go
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	_, err = os.Stat(chezmoiSourceDir)
	return err == nil
}

func IsDir(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		return false
	}
	return fileInfo.IsDir()
}

func GetChezmoiManagedFiles(path ...string) ([]list.Item, error) {
	output, err := exec.Command(Command, append([]string{"managed"}, path...)...).Output()
	//TODO:migrate to use better-go
	if err != nil {
		return nil, err
	}
	files := strings.Split(string(output), "\n")
    var mangedFilesWithoutParentDirName []string

	if len(path) > 0 {
		lastValue := path[len(path)-1]
		for _, v := range files {
			pathWithouBasePath := strings.Join(strings.Split(v, "/")[1:], "/")
			mangedFilesWithoutParentDirName = append(mangedFilesWithoutParentDirName , pathWithouBasePath)
		}
		return getFileEntery(mangedFilesWithoutParentDirName, true, lastValue)
	}

	userHomeDir, _ := os.UserHomeDir()

	return getFileEntery(files, true, userHomeDir)

}

func GetUnmanagedFiles(path ...string) ([]list.Item, error) {
	unmanaged, err := exec.Command(Command, append([]string{"unmanaged"}, path...)...).Output()
	if err != nil {
		return nil, err
	}

	var unmangedfiles []string
	files := strings.Split(string(unmanaged), "\n")

	if len(path) > 0 {
		for _, v := range files {
			pathWitoutTheBasePath := strings.Join(strings.Split(v, "/")[1:], "/")
			unmangedfiles = append(unmangedfiles, pathWitoutTheBasePath)
		}
		return getFileEntery(unmangedfiles, false, path[len(path)-1])
	}
	unmangedfiles = files
	userHomeDir, _ := os.UserHomeDir()

	return getFileEntery(unmangedfiles, false, userHomeDir)
}

func getFileEntery(files []string, isManaged bool, baseDir string) ([]list.Item, error) {
	var fileEntery []list.Item

	var currentTempDir tempDir
	for i := range files {
		path := strings.Trim(files[i], " ")
		if path == "" {
			continue
		}

		absolutePah, err := utils.GetAbsolutePath(path, baseDir)

		if err != nil {
			return nil, err
		}

		isCurrentPathDir := IsDir(absolutePah)

		if currentTempDir.path == strings.Split(path, "/")[0] {
			continue
		}

		if isCurrentPathDir {
			currentTempDir = tempDir{path: strings.Split(path, "/")[0]}
		} else {
			if strings.Split(path, "/")[0] == currentTempDir.path {
				continue
			}
		}

		//TODO: migrate to better-go
		fileEntery = append(fileEntery, helpers.FileEntry{Name: path, Path: path, IsManaged: isManaged, IsDir: isCurrentPathDir})
	}
	return fileEntery, nil
}

func GetAllFiles(arg ...string) ([]list.Item, error) {
	managedFiles, err := GetChezmoiManagedFiles(arg...)
	//TOOD:migrate to use better-go
	if err != nil {
		return nil, err
	}
	var unmanagedFiles []list.Item

	if len(arg) > 0 {
		unmanagedFiles, err = GetUnmanagedFiles(arg[len(arg)-1])
	} else {
		unmanagedFiles, err = GetUnmanagedFiles()
	}
	//TOOD:migrate to use better-go
	if err != nil {
		return nil, err
	}

	if len(arg) > 0 {
		allFiles := append([]list.Item{helpers.FileEntry{
			Name:      "..",
			Path:      "..",
			IsManaged: false,
			IsDir:     true,
		}}, append(managedFiles, unmanagedFiles...)...)

		return allFiles, nil
	}
	return append(managedFiles, unmanagedFiles...), nil
}

func RunChezmoiCommand(command ...string) error {
	cmd := exec.Command(Command, command...)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to run chezmoi %s: %w", command, err)
	}
	return nil
}

func RunChezmoiCommandInteractive(command ...string) error {
	cmd := exec.Command(Command, command...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to run chezmoi %s: %w", command, err)
	}
	return nil
}
func ExecuteInGoRoutine(fn func() error) {
	go func() {
		if err := fn(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	}()
}
func AddFile(path string) error {
	ExecuteInGoRoutine(func() error {
		return RunChezmoiCommandInteractive("add", path)
	})
	return nil
}

func ForgetFile(path string) error {
	return RunChezmoiCommandInteractive("forget", path)
}

func EditFile(path string) tea.Cmd {
	c := exec.Command(Command, "edit", path)

	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return tea.ExecProcess(c, func(err error) tea.Msg {
		return AltarnateScreeenExec{err}
	})
}

func Diff(path ...string) tea.Cmd {
	c := exec.Command(Command, append([]string{"diff"}, path...)...)

	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return tea.ExecProcess(c, func(err error) tea.Msg {
		return AltarnateScreeenExec{err}
	})
}
