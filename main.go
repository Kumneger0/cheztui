package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/kumneger0/chez-tui/chezmoi"
	"github.com/kumneger0/chez-tui/core"
)

var keyBindings = []struct {
	key         string
	description string
}{
	{key: "a", description: "Add file"},
	{key: "r", description: "Remove file"},
	{key: "e", description: "Edit file"},
	{key: "d", description: "Show diff for Current File"},
	{key: "enter", description: "Navigate to directory"},
	{key: "D", description: "Show Diffs For All Files"},
	{key: "A", description: "Apply changes"},
	{key: "L", description: "list all files"},
	{key: "m", description: "Show only managed files"},
	{key: "u", description: "Show only unmanaged files"},
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

		var confirm bool

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Do you want to us to initialize it for you?").
					Affirmative("Yes!").
					Negative("No.").
					Value(&confirm),
			),
		)

		err = form.Run()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if confirm {
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

	bubleList := list.New(managedFiles, core.CustomDelegate{}, 0, 0)
	bubleList.AdditionalFullHelpKeys = func() []key.Binding {
		var keys []key.Binding
		for _, v := range keyBindings {
			newKey := key.NewBinding(key.WithKeys(v.key), key.WithHelp(v.key, v.description))
			keys = append(keys, newKey)
		}
		return keys
	}
	currentDir, _ := os.UserHomeDir()

	m := core.Model{Files: bubleList, CurrentDir: currentDir}

	p := tea.NewProgram(m, tea.WithAltScreen())

	_, err = p.Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
		return
	}
}
