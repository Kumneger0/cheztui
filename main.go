package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kumneger0/chez-tui/chezmoi"
)

var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#7D56F4")).
			Bold(true)
)

type customDelegate struct {
	list.DefaultDelegate
}

func (d customDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	entry, ok := item.(chezmoi.FileEntry)
	if !ok {
		return
	}

	str := entry.Title()

	if index == m.Index() {
		fmt.Fprint(w, selectedStyle.Render(" "+str+" "))
		if entry.IsManaged {
			fmt.Fprint(w, " [managed]")
		}
	} else {
		fmt.Fprint(w, normalStyle.Render(" "+str+" "))
	}
}

var keyBindings = []struct {
	key         string
	description string
}{
	{key: "a", description: "Add file"},
	{key: "r", description: "Remove file"},
	{key: "e", description: "Edit file"},
	{key: "p", description: "Push to GitHub"},
	{key: "d", description: "Show diff"},
	{key: "A", description: "Apply changes"},
}

type model struct {
	files list.Model
}

func getAbsolutePath(path string) (string, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userHomeDir, path), nil
}

func (m model) Init() tea.Cmd {
	return func() tea.Msg {
		err := chezmoi.RunChezmoiCommand("status")
		if err != nil {
			return tea.Printf("Error: %v", err)
		}
		return nil
	}

}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		selectedFile := m.files.SelectedItem()
		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit
		case "a":
			if selectedFile != nil {
				path, err := getAbsolutePath(selectedFile.FilterValue())
				if err != nil {
					fmt.Println("Error getting absolute path:", err)
				}
				fmt.Println("Adding file:", path)
				err = chezmoi.AddFile(path)
				if err != nil {
					fmt.Println(err.Error())
				} else {
					// figure out a way to show nice toast message
				}
			}
		case "r":
			if selectedFile != nil {
				path, err := getAbsolutePath(selectedFile.FilterValue())
				if err != nil {
					fmt.Println("Error getting absolute path:", err)
				}
				err = chezmoi.ForgetFile(path)
				if err != nil {
					//TODO: figure out a way to show nice toast message
				} else {
					//TODO: figure out a way to show nice toast message
				}
			}
		case "e":
			if selectedFile != nil {
				path, err := getAbsolutePath(selectedFile.FilterValue())
				if err != nil {
					fmt.Println("Error getting absolute path:", err)
				}
				err = chezmoi.EditFile(path)
				if err != nil {
					//TODO: figure out a way to show nice toast message
				} else {
					//TODO: figure out a way to show nice toast message
				}
			}
		case "A":
			if selectedFile != nil {
				err := chezmoi.RunChezmoiCommand("apply")
            if err != nil {
					//TODO: figure out a way to show nice toast message
				} else {
					//TODO: figure out a way to show nice toast message
				}
			}
		case "p":
			if selectedFile != nil {
				//TODO: push to github
			}
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.files.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.files, cmd = m.files.Update(msg)
	return m, cmd
}

func (m model) View() string {
	listView := m.files.View()
	return docStyle.Render(listView)
}

func main() {
	if !chezmoi.IsChezmoiInstalled() {
		fmt.Println("Chezmoi is not installed please install chezmoi first")
		os.Exit(1)
		return
	}

	if !chezmoi.IsChezmoiInitialized() {
		fmt.Println("Chezmoi is not initialized please initialize chezmoi first")
		var userPromt string = "y"
		fmt.Print("Do you want to us to initialize it for you [Y/n]?")
		_, err := fmt.Scan(&userPromt)

		//TODO:migrate to update this to use better-go
		if err != nil {
			fmt.Println("Error reading input:", err)
			os.Exit(1)
		}

		if userPromt == "y" || userPromt == "Y" {
			err := chezmoi.RunChezmoiCommand("init")
			if err != nil {
				fmt.Println("Error:", err)
			}
			fmt.Println("Chezmoi initialized successfully")
		} else {
			os.Exit(1)
		}
	}

	managedFiles, err := chezmoi.GetChezmoiManagedFiles()

	//TODO:migrate to use better-go
	if err != nil {
		fmt.Println("Error getting managed files:", err)
		os.Exit(1)
		return

	}

	bubleList := list.New(managedFiles, customDelegate{}, 0, 0)
	bubleList.AdditionalFullHelpKeys = func() []key.Binding {
		var keys []key.Binding
		for _, v := range keyBindings {
			newKey := key.NewBinding(key.WithKeys(v.key), key.WithHelp(v.key, v.description))
			//TODO: use better-go
			keys = append(keys, newKey)
		}
		return keys
	}

	m := model{files: bubleList}

	p := tea.NewProgram(m, tea.WithAltScreen())

	_, err = p.Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
		return
	}

}
