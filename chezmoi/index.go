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
	unmanaged, err := exec.Command(Command, "unmanaged").Output()
	if err != nil {
		return nil, err
	}

	unmangedfiles := strings.Split(string(unmanaged), "\n")

	for i := range unmangedfiles {
		path := strings.Trim(unmangedfiles[i], " ")
		if path == "" {
			continue
		}
		//TODO: migrate to better-go
		allManagedFilesEntery = append(allManagedFilesEntery, FileEntry{Name: path, Path: path, IsManaged: false})
	}
	return allManagedFilesEntery, nil
}

func RunChezmoiCommand(command string) error {
	fmt.Println("The commad to excute", command)
	cmd := exec.Command(Command, command)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to run chezmoi %s: %w", command, err)
	}
	return nil
}
func RunChezmoiCommand2(command string, args ...string) error {
	fmt.Println("The commad to excute", command)
	cmd := exec.Command(Command, append([]string{command}, args...)...)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to run chezmoi %s: %w", command, err)
	}
	return nil
}

func AddFile(path string) error {
	return RunChezmoiCommand2("add", path)
}

func ForgetFile(path string) error {
	return RunChezmoiCommand(fmt.Sprintf("forget %s", path))
}

type EditError struct{ err error }

func EditFile(path string) error {
	c := exec.Command(Command, fmt.Sprintf("edit %s", path))
	_ = tea.ExecProcess(c, func(err error) tea.Msg {
		return EditError{err: err}
	})
	return nil
}
