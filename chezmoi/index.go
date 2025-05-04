package chezmoi

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type FileEntry struct {
	Name      string
	Path      string
	IsManaged bool
}

type AltarnateScreeenExec struct{ error }

func (f FileEntry) Title() string       { return f.Name }
func (f FileEntry) FilterValue() string { return f.Name }

const Command = "chezmoi"

func IsChezmoiInstalled() bool {
	_, err := exec.LookPath(Command)
	return err == nil
}

func getChezmoiSourceDir() (string, error) {
	userHomeDir, err := os.UserHomeDir()
	//TODO:migrate to better-go
	if err != nil {
		return "", err
	}
	return filepath.Join(userHomeDir, ".local", "share", "chezmoi"), nil
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

func GetChezmoiManagedFiles() ([]list.Item, error) {
	output, err := exec.Command(Command, "managed").Output()
	//TODO:migrate to use better-go
	if err != nil {
		return nil, err
	}
	files := strings.Split(string(output), "\n")

	var allManagedFilesEntery []list.Item

	for i := range files {
		path := strings.Trim(files[i], " ")
		if path == "" {
			continue
		}
		//TODO: migrate to better-go
		allManagedFilesEntery = append(allManagedFilesEntery, FileEntry{Name: path, Path: path, IsManaged: true})
	}
	return allManagedFilesEntery, nil
}

func GetUnmanagedFiles() ([]list.Item, error) {
	unmanaged, err := exec.Command(Command, "unmanaged").Output()
	if err != nil {
		return nil, err
	}

	var allUnManagedFilesEntery []list.Item

	unmangedfiles := strings.Split(string(unmanaged), "\n")

	for i := range unmangedfiles {
		path := strings.Trim(unmangedfiles[i], " ")
		if path == "" {
			continue
		}
		//TODO: migrate to better-go
		allUnManagedFilesEntery = append(allUnManagedFilesEntery, FileEntry{Name: path, Path: path, IsManaged: false})
	}
	return allUnManagedFilesEntery, nil
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

func AddFile(path string) error {
	return RunChezmoiCommandInteractive("add", path)
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
