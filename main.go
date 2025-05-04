package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	entry, ok := item.(fileEntry)
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

func getChezmoiSourceDir() (string, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userHomeDir, ".local", "share", "chezmoi"), nil
}

var keyBindings = []struct {
	key         string
	description string
}{
	{key: "q", description: "Quit"},
	{key: "a", description: "Add file"},
	{key: "r", description: "Remove file"},
	{key: "e", description: "Edit file"},
	{key: "p", description: "Push to GitHub"},
	{key: "d", description: "Show diff"},
	{key: "A", description: "Apply changes"},
}

const command = "chezmoi"

type fileEntry struct {
	Name      string
	Path      string
	IsManaged bool
}

func (i fileEntry) Title() string       { return i.Name }
func (i fileEntry) Description() string { return i.Name }
func (i fileEntry) FilterValue() string { return i.Name }

type model struct {
	files list.Model
}

func (m model) setWindowTitle() string {
	return "Cheztui"
}

func isChezmoiInstalled() bool {
	_, err := exec.LookPath(command)
	return err == nil
}

func isChezmoiInitialized() bool {
	chezmoiSourceDir, err := getChezmoiSourceDir()
	//TODO:migrate to use better-go
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	_, err = os.Stat(chezmoiSourceDir)
	return err == nil
}

func getChezmoiManagedFiles() ([]list.Item, error) {

	output, err := exec.Command(command, "managed").Output()
	//TODO:migrate to use better-go
	if err != nil {
		return nil, err
	}
	files := strings.Split(string(output), "\n")

	allManagedFilesEntery := []list.Item{}

	for i := range files {
		path := strings.Trim(files[i], " ")
		//migrate to better-go

		allManagedFilesEntery = append(allManagedFilesEntery, fileEntry{Name: path, Path: path, IsManaged: true})
	}
	return allManagedFilesEntery, nil
}

func (m model) Init() tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command(command, "status")
		err := cmd.Run()
		if err != nil {
			return tea.Printf("Error: %v", err)
		}
		return nil
	}

}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "a":
			selectedFile := m.files.SelectedItem()
			if selectedFile != nil {
				fmt.Println(selectedFile.FilterValue())
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
	if !isChezmoiInstalled() {
		fmt.Println("Chezmoi is not installed please install chezmoi first")
		os.Exit(1)
		return
	}

	if !isChezmoiInitialized() {
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
			cmd := exec.Command(command, "init")
			err := cmd.Run()
			if err != nil {
				fmt.Println("Error:", err)
			}
			fmt.Println("Chezmoi initialized successfully")
		} else {
			os.Exit(1)
		}
	}

	managedFiles, err := getChezmoiManagedFiles()

	//TODO:migrate to use better-go
	if err != nil {
		fmt.Println("Error getting managed files:", err)
		os.Exit(1)
		return

	}

	bubleList := list.New(managedFiles, customDelegate{}, 0, 0)
	bubleList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
			key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add file")),
			key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "remove file")),
			key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit file")),
			key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "push to GitHub")),
			key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "show diff")),
			key.NewBinding(key.WithKeys("A"), key.WithHelp("A", "apply changes")),
		}
	}

	m := model{files: bubleList}

	p := tea.NewProgram(m, tea.WithAltScreen())

	styles := list.DefaultStyles()
	styles.HelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))

	_, err = p.Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
		return
	}

}
