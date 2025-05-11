package core

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

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

type CustomDelegate struct {
	list.DefaultDelegate
}

func (d CustomDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	entry, ok := item.(helpers.FileEntry)
	if !ok {
		return
	}

	str := lipgloss.NewStyle().Width(50).Render(entry.Title())

	if entry.IsDir {
		str = "📁 " + str
	} else {
		str = "📄 " + str
	}

	if entry.IsManaged {
		str = str + " ✅"
	} else {
		str = str + " ❌"
	}

	if index == m.Index() {
		fmt.Fprint(w, selectedStyle.Render(" "+str+" "))
	} else {
		fmt.Fprint(w, normalStyle.Render(" "+str+" "))
	}
}

type Model struct {
	Files             list.Model
	IsAltranateScreen bool
	CurrentDir        string
	Toast             *toast
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
		m.Toast = &toast{
			message: msg.message,
			expires: msg.expires,
		}
		return m, tea.Tick(time.Until(msg.expires), func(t time.Time) tea.Msg {
			return clearToastMsg{}
		})

	case clearToastMsg:
		m.Toast = nil
		return m, nil

	case tea.KeyMsg:
		selectedFile := m.Files.SelectedItem()
		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		case "m":

			var managedFiles []list.Item
			userHomeDir, _ := os.UserHomeDir()
			if userHomeDir == m.CurrentDir {
				managedFiles, _ = chezmoi.GetAllFiles()
			} else {
				managedFiles, _ = chezmoi.GetChezmoiManagedFiles("-i", "files", m.CurrentDir)
			}

			return m, m.Files.SetItems(managedFiles)
		case "u":

			var unmanagedFiles []list.Item
			userHomeDir, _ := os.UserHomeDir()

			if userHomeDir == m.CurrentDir {
				unmanagedFiles, _ = chezmoi.GetUnmanagedFiles()
			} else {
				unmanagedFiles, _ = chezmoi.GetUnmanagedFiles(m.CurrentDir)
			}

			return m, m.Files.SetItems(unmanagedFiles)

		case "a":
			if selectedFile != nil {
				path := filepath.Join(m.CurrentDir, selectedFile.FilterValue())
				err := chezmoi.AddFile(path)
				if err != nil {
					return m, showToast(err.Error(), 2*time.Second)
				} else {

					var updatedFiles []list.Item
					userHomeDir, _ := os.UserHomeDir()

					if userHomeDir == m.CurrentDir {
						updatedFiles, err = chezmoi.GetAllFiles()

					} else {
						updatedFiles, err = chezmoi.GetAllFiles("-i", "files", m.CurrentDir)
					}

					if err != nil {
						return m, showToast(err.Error(), 2*time.Second)
					}
					m.Files.SetItems(updatedFiles)
					return m, showToast("File added successfully", 2*time.Second)
				}
			}

		case "r":
			if selectedFile != nil {
				path := filepath.Join(m.CurrentDir, selectedFile.FilterValue())
				err := chezmoi.ForgetFile(path)
				if err != nil {
					return m, showToast(err.Error(), 2*time.Second)
				} else {
					updatedFiles, err := chezmoi.GetChezmoiManagedFiles()
					if err != nil {
						return m, showToast(err.Error(), 2*time.Second)
					}
					m.Files.SetItems(updatedFiles)
				}
			}

		case "d":
			if selectedFile != nil {
				path := filepath.Join(m.CurrentDir, selectedFile.FilterValue())
				m.IsAltranateScreen = true
				return m, chezmoi.DiffFile(path)
			}

		case "e":
			if selectedFile != nil {
				fileProperty := utils.FindFileProperty(selectedFile.FilterValue(), m.Files.Items())
				if fileProperty.IsDir {
					return m, showToast("Cannot edit directory", 2*time.Second)
				}
				path := filepath.Join(m.CurrentDir, selectedFile.FilterValue())
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
					m.Files.SetItems(updatedFiles)
					return m, showToast("Changes applied successfully", 2*time.Second)
				}
			}

		case "p":
			if selectedFile != nil {
				// TODO: push to GitHub
			}

		case "enter":
			if selectedFile != nil {
				fileProperty := utils.FindFileProperty(selectedFile.FilterValue(), m.Files.Items())

				if !fileProperty.IsDir {
					return m, showToast("Cannot navigate to file", 2*time.Second)
				}
				fullPath := filepath.Join(m.CurrentDir, selectedFile.FilterValue())
				homeDir, _ := os.UserHomeDir()

				if fullPath == homeDir {
					allFiles, _ := chezmoi.GetAllFiles()
					m.Files.SetItems(allFiles)
					m.CurrentDir = homeDir
					return m, nil
				}

				filesNewDir, err := utils.GetFilesFromSpecificPath(fullPath)
				if err != nil {
					fmt.Println("There was an error while navigating to new directory", err.Error())
				}
				filesNewDir = append([]list.Item{helpers.FileEntry{Name: "..", Path: "..", IsManaged: false, IsDir: true, BackButton: true}}, filesNewDir...)
				m.Files.SetItems(filesNewDir)
				m.CurrentDir = fullPath
			}

			return m, nil
		}

	case chezmoi.AltarnateScreeenExec:
		m.IsAltranateScreen = false
		return m, nil

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.Files.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.Files, cmd = m.Files.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	listView := m.Files.View()
	m.Files.Title = "Chezmoi Files"

	if m.Toast != nil && time.Now().Before(m.Toast.expires) {
		toastView := renderToast(m.Toast.message)
		return listView + "\n" + toastView
	}

	return docStyle.Render(listView)
}
