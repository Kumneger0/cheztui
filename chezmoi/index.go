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
	return utils.GetAbsolutePath(".local/share/chezmoi")
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
	return getFileEntery(files, true)
}

func GetUnmanagedFiles(path ...string) ([]list.Item, error) {
	unmanaged, err := exec.Command(Command, append([]string{"unmanaged"}, path...)...).Output()
	if err != nil {
		return nil, err
	}

	unmangedfiles := strings.Split(string(unmanaged), "\n")
	return getFileEntery(unmangedfiles, false)

}

func getFileEntery(files []string, isManaged bool) ([]list.Item, error) {
	var fileEntery []list.Item

	var currentTempDir tempDir
	for i := range files {
		path := strings.Trim(files[i], " ")
		if path == "" {
			continue
		}

		absolutePah, err := utils.GetAbsolutePath(path)

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
		fileEntery = append(fileEntery, helpers.FileEntry{Name: path, Path: path, IsManaged:isManaged , IsDir: isCurrentPathDir})
	}
	return fileEntery, nil
}

func GetAllFiles() ([]list.Item, error) {
	managedFiles, err := GetChezmoiManagedFiles()
	//TOOD:migrate to use better-go
	if err != nil {
		return nil, err
	}
	unmanagedFiles, err := GetUnmanagedFiles()
	//TOOD:migrate to use better-go
	if err != nil {
		return nil, err
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

func DiffFile(path string) tea.Cmd {
	c := exec.Command(Command, "diff", path)

	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return tea.ExecProcess(c, func(err error) tea.Msg {
		return AltarnateScreeenExec{err}
	})
}
