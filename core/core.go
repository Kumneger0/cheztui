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
		return handleToastMsg(m, msg)
	case clearToastMsg:
		return handleClearToastMsg(m)
	case tea.KeyMsg:
		return handleKeyMsg(m, msg)
	case chezmoi.AltarnateScreeenExec:
		return handleAlternateScreenExec(m)
	case tea.WindowSizeMsg:
		return handleWindowSizeMsg(m, msg)
	default:
		return updateFilesList(m, msg)
	}
}

func handleToastMsg(m Model, msg toastMsg) (tea.Model, tea.Cmd) {
	m.Toast = &toast{
		message: msg.message,
		expires: msg.expires,
	}
	return m, tea.Tick(time.Until(msg.expires), func(t time.Time) tea.Msg {
		return clearToastMsg{}
	})
}

func handleClearToastMsg(m Model) (tea.Model, tea.Cmd) {
	m.Toast = nil
	return m, nil
}

func handleAlternateScreenExec(m Model) (tea.Model, tea.Cmd) {
	m.IsAltranateScreen = false
	return m, nil
}

func handleWindowSizeMsg(m Model, msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	h, v := docStyle.GetFrameSize()
	m.Files.SetSize(msg.Width-h, msg.Height-v)
	return updateFilesList(m, msg)
}

func updateFilesList(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.Files, cmd = m.Files.Update(msg)
	return m, cmd
}

func handleKeyMsg(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	selectedFile := m.Files.SelectedItem()

	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "m":
		return handleManagedFilesView(m)
	case "u":
		return handleUnmanagedFilesView(m)
	case "a":
		return handleAddFile(m, selectedFile)
	case "L":
		return handleListAllFiles(m)
	case "r":
		return handleRemoveFile(m, selectedFile)
	case "d":
		return handleDiffFile(m, selectedFile)
	case "D":
		return handleDiffAllFiles(m)
	case "e":
		return handleEditFile(m, selectedFile)
	case "A":
		return handleApplyChanges(m, selectedFile)
	case "p":
		if selectedFile != nil {
			// TODO: push to GitHub
		}
		return m, nil
	case "enter":
		return handleNavigateDirectory(m, selectedFile)
	default:
		return updateFilesList(m, msg)
	}
}

func handleManagedFilesView(m Model) (tea.Model, tea.Cmd) {
	var managedFiles []list.Item
	userHomeDir, _ := os.UserHomeDir()

	if userHomeDir == m.CurrentDir {
		managedFiles, _ = chezmoi.GetChezmoiManagedFiles()
	} else {
		managedFiles, _ = chezmoi.GetChezmoiManagedFiles("-i", "files", m.CurrentDir)
	}

	return m, m.Files.SetItems(managedFiles)
}

func handleUnmanagedFilesView(m Model) (tea.Model, tea.Cmd) {
	var unmanagedFiles []list.Item
	userHomeDir, _ := os.UserHomeDir()

	if userHomeDir == m.CurrentDir {
		unmanagedFiles, _ = chezmoi.GetUnmanagedFiles()
	} else {
		unmanagedFiles, _ = chezmoi.GetUnmanagedFiles(m.CurrentDir)
	}

	return m, m.Files.SetItems(unmanagedFiles)
}

func handleAddFile(m Model, selectedFile list.Item) (tea.Model, tea.Cmd) {
	if selectedFile == nil {
		return m, nil
	}

	path := filepath.Join(m.CurrentDir, selectedFile.FilterValue())
	err := chezmoi.AddFile(path)

	if err != nil {
		return m, showToast(err.Error(), 2*time.Second)
	}

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

func handleListAllFiles(m Model) (tea.Model, tea.Cmd) {
	var allFiles []list.Item
	userHomeDir, _ := os.UserHomeDir()

	if userHomeDir == m.CurrentDir {
		allFiles, _ = chezmoi.GetAllFiles()
	} else {
		allFiles, _ = chezmoi.GetAllFiles("-i", "files", m.CurrentDir)
	}

	return m, m.Files.SetItems(allFiles)
}

func handleRemoveFile(m Model, selectedFile list.Item) (tea.Model, tea.Cmd) {
	if selectedFile == nil {
		return m, nil
	}

	path := filepath.Join(m.CurrentDir, selectedFile.FilterValue())
	err := chezmoi.ForgetFile(path)

	if err != nil {
		return m, showToast(err.Error(), 2*time.Second)
	}

	userHomeDir, _ := os.UserHomeDir()
	var updatedFiles []list.Item

	if userHomeDir == m.CurrentDir {
		updatedFiles, err = chezmoi.GetAllFiles()
	} else {
		updatedFiles, err = chezmoi.GetAllFiles("-i", "files", m.CurrentDir)
	}

	if err != nil {
		return m, showToast(err.Error(), 2*time.Second)
	}

	m.Files.SetItems(updatedFiles)
	return m, nil
}

func handleDiffFile(m Model, selectedFile list.Item) (tea.Model, tea.Cmd) {
	if selectedFile == nil {
		return m, nil
	}

	path := filepath.Join(m.CurrentDir, selectedFile.FilterValue())
	m.IsAltranateScreen = true
	return m, chezmoi.Diff(path)
}

func handleDiffAllFiles(m Model) (tea.Model, tea.Cmd) {
	m.IsAltranateScreen = true
	return m, chezmoi.Diff()
}

func handleEditFile(m Model, selectedFile list.Item) (tea.Model, tea.Cmd) {
	if selectedFile == nil {
		return m, nil
	}

	fileProperty := utils.FindFileProperty(selectedFile.FilterValue(), m.Files.Items())
	if fileProperty.IsDir {
		return m, showToast("Cannot edit directory", 2*time.Second)
	}

	path := filepath.Join(m.CurrentDir, selectedFile.FilterValue())
	return m, chezmoi.EditFile(path)
}

func handleApplyChanges(m Model, selectedFile list.Item) (tea.Model, tea.Cmd) {
	if selectedFile == nil {
		return m, nil
	}

	err := chezmoi.RunChezmoiCommandInteractive("apply")
	if err != nil {
		return m, showToast(err.Error(), 2*time.Second)
	}

	return m, showToast("Changes applied successfully", 2*time.Second)
}

func handleNavigateDirectory(m Model, selectedFile list.Item) (tea.Model, tea.Cmd) {
	if selectedFile == nil {
		return m, nil
	}

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

	filesNewDir = append([]list.Item{helpers.FileEntry{
		Name:       "..",
		Path:       "..",
		IsManaged:  false,
		IsDir:      true,
		BackButton: true,
	}}, filesNewDir...)

	m.Files.SetItems(filesNewDir)
	m.CurrentDir = fullPath
	return m, nil
}

func (m Model) View() string {
	m.Files.Title = "Chezmoi Files"
	m.Files.SetShowTitle(true)
	
	listView := m.Files.View()

	if m.Toast != nil && time.Now().Before(m.Toast.expires) {
		toastView := renderToast(m.Toast.message)
		return listView + "\n" + toastView
	}

	return docStyle.Render(listView)
}
