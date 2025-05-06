package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kumneger0/chez-tui/chezmoi"
	"github.com/kumneger0/chez-tui/helpers.go"
	"github.com/kumneger0/chez-tui/utils"
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

type toast struct {
	message string
	expires time.Time
}

type customDelegate struct {
	list.DefaultDelegate
}

func (d customDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	entry, ok := item.(helpers.FileEntry)
	if !ok {
		return
	}

	str := entry.Title()

	if index == m.Index() {
		fmt.Fprint(w, selectedStyle.Render(" "+str+" "))
	} else {
		if entry.IsManaged {
			fmt.Fprint(w, " [managed]")
		}
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
	{key: "m", description: "Show only managed files"},
	{key: "u", description: "Show only unmanaged files"},
}

type Model struct {
	files             list.Model
	isAltranateScreen bool
	currentDir        string
	toast             *toast
}

type toastMsg struct {
	message string
	expires time.Time
}

type clearToastMsg struct{}

func showToast(message string, duration time.Duration) tea.Cmd {
	return func() tea.Msg {
		return toastMsg{
			message: message,
			expires: time.Now().Add(duration),
		}
	}
}

func renderToast(msg string) string {
	style := lipgloss.NewStyle().
		Background(lipgloss.Color("#FF5F5F")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 1).
		Bold(true)

	return style.Render(msg)
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case toastMsg:
		m.toast = &toast{
			message: msg.message,
			expires: msg.expires,
		}
		return m, tea.Tick(time.Until(msg.expires), func(t time.Time) tea.Msg {
			return clearToastMsg{}
		})

	case clearToastMsg:
		m.toast = nil
		return m, nil

	case tea.KeyMsg:
		selectedFile := m.files.SelectedItem()
		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		case "m":
			managedFiles, err := chezmoi.GetChezmoiManagedFiles()
			if err != nil {
				return m, showToast(err.Error(), 2*time.Second)
			}
			m.files.SetItems(managedFiles)

		case "u":
			unmanagedFiles, err := chezmoi.GetUnmanagedFiles()
			if err != nil {
				return m, showToast(err.Error(), 2*time.Second)
			}
			m.files.SetItems(unmanagedFiles)

		case "a":
			if selectedFile != nil {
				path := filepath.Join(m.currentDir, selectedFile.FilterValue())
				err := chezmoi.AddFile(path)
				if err != nil {
					return m, showToast(err.Error(), 2*time.Second)
				} else {
					updatedFiles, err := chezmoi.GetAllFiles()
					if err != nil {
						return m, showToast(err.Error(), 2*time.Second)
					}
					m.files.SetItems(updatedFiles)
					return m, showToast("File added successfully", 2*time.Second)
				}
			}

		case "r":
			if selectedFile != nil {
				path := filepath.Join(m.currentDir, selectedFile.FilterValue())
				err := chezmoi.ForgetFile(path)
				if err != nil {
					return m, showToast(err.Error(), 2*time.Second)
				} else {
					updatedFiles, err := chezmoi.GetChezmoiManagedFiles()
					if err != nil {
						return m, showToast(err.Error(), 2*time.Second)
					}
					m.files.SetItems(updatedFiles)
				}
			}

		case "d":
			if selectedFile != nil {
				path := filepath.Join(m.currentDir, selectedFile.FilterValue())
				m.isAltranateScreen = true
				return m, chezmoi.DiffFile(path)
			}

		case "e":
			if selectedFile != nil {
				fileProperty := utils.FindFileProperty(selectedFile.FilterValue(), m.files.Items())
				if fileProperty.IsDir {
					return m, showToast("Cannot edit directory", 2*time.Second)
				}
				path := filepath.Join(m.currentDir, selectedFile.FilterValue())
				return m, chezmoi.EditFile(path)
			}

		case "A":
			if selectedFile != nil {
				err := chezmoi.RunChezmoiCommandInteractive("apply")
				if err != nil {
					return m, showToast(err.Error(), 2*time.Second)
				} else {
					updatedFiles, err := chezmoi.GetChezmoiManagedFiles()
					if err != nil {
						return m, showToast(err.Error(), 2*time.Second)
					}
					m.files.SetItems(updatedFiles)
					return m, showToast("Changes applied successfully", 2*time.Second)
				}
			}

		case "p":
			if selectedFile != nil {
				// TODO: push to GitHub
			}

		case "enter":
			if selectedFile != nil {
				fileProperty := utils.FindFileProperty(selectedFile.FilterValue(), m.files.Items())

				if !fileProperty.IsDir {
					return m, showToast("Cannot navigate to file", 2*time.Second)
				}

				fullPath := filepath.Join(m.currentDir, selectedFile.FilterValue())
				filesNewDir, err := utils.GetFilesFromSpecificPath(fullPath)
				if err != nil {
					fmt.Println("There was an error while navigating to new directory", err.Error())
				}
				filesNewDir = append([]list.Item{helpers.FileEntry{Name: "..", Path: "..", IsManaged: false, IsDir: true, BackButton: true}}, filesNewDir...)
				m.files.SetItems(filesNewDir)
				m.currentDir = fullPath
			}

			return m, nil
		}

	case chezmoi.AltarnateScreeenExec:
		m.isAltranateScreen = false
		return m, nil

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.files.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.files, cmd = m.files.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	listView := m.files.View()
	m.files.Title = "Chezmoi Files"

	if m.toast != nil && time.Now().Before(m.toast.expires) {
		toastView := renderToast(m.toast.message)
		return listView + "\n" + toastView
	}

	return docStyle.Render(listView)
}

func main() {
	err := chezmoi.RunChezmoiCommand("status")
	if err != nil {
		tea.Printf("Error: %v", err)
	}

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

	managedFiles, err := chezmoi.GetAllFiles()
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
			keys = append(keys, newKey)
		}
		return keys
	}
	currentDir, _ := os.UserHomeDir()

	m := Model{files: bubleList, currentDir: currentDir}

	p := tea.NewProgram(m, tea.WithAltScreen())

	_, err = p.Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
		return
	}
}
