# Chez-TUI

Chez-TUI is a tiny [chezmoi](https://www.chezmoi.io/) wrapper that makes it super easy to manage your dotfiles. So you don't have to remember a bunch of chezmoi commands.

## Why Chez-TUI ?
i use chezmoi to manage my dotfiles and im so lazy to type out the commands all the time, so i made this little tool to make my life easier.

## What It Can Do

- **Manage Files**: Quickly view, add, edit, or remove files managed by chezmoi.
- **File Picker**: Browse your file system and pick files right from the terminal.
- **Shortcut Commands**: Do common tasks like `apply`, `diff`, and `edit` with just a key press.
- **Notifications**: Get instant feedback on what’s happening with toast messages.
- **Switch Views**: Easily toggle between different screens for specific tasks(eg editing a file).


## What You’ll Need

- [chezmoi](https://www.chezmoi.io/): Make sure it’s installed and set up.

## How to Use It

1. Clone the repo:
   ```bash
   git clone https://github.com/kumneger0/cheztui.git
   cd cheztui
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Build the app:
   ```bash
   go build -o cheztui main.go
   ```

4. Run it:
   ```bash
   ./cheztui
   ```





## Key Shortcuts with Equivalent chezmoi Commands

- `a`: Add a file to chezmoi. Equivalent to:
  ```bash
  chezmoi add <file>
  ```
- `r`: Remove a file from chezmoi. Equivalent to:
  ```bash
  chezmoi forget <file>
  ```
- `e`: Edit a file. Equivalent to:
  ```bash
  chezmoi edit <file>
  ```
- `d`: Show the diff of the current file. Equivalent to:
  ```bash
  chezmoi diff <file>
  ```
- `D`: Show diffs for all files. Equivalent to:
  ```bash
  chezmoi diff
  ```
- `enter`: Navigate to a directory (no direct chezmoi equivalent; used for navigation in the TUI).
- `A`: Apply changes. Equivalent to:
  ```bash
  chezmoi apply
  ```
- `L`: Show both managed and unmanaged files. Equivalent to:
  ```bash
  chezmoi managed && chezmoi unmanaged
  ```
- `m`: Show only managed files. Equivalent to:
  ```bash
  chezmoi managed
  ```
- `u`: Show unmanaged files. Equivalent to:
  ```bash
  chezmoi unmanaged
  ```
- `esc`: Close the file picker or go back (no direct chezmoi equivalent; used for navigation in the TUI).
- `ctrl+c` or `q`: Quit the app (no direct chezmoi equivalent; used to exit the TUI).

## Major Dependencies

Chez-TUI is built using the following major dependencies:

- [chezmoi](https://www.chezmoi.io/): The core tool for managing dotfiles.
- [Bubble Tea](https://github.com/charmbracelet/bubbletea): A powerful, fun, and flexible Go framework for building terminal applications.
- [Bubbles](https://github.com/charmbracelet/bubbles): Components for Bubble Tea, like lists, file pickers.
- [Lip Gloss](https://github.com/charmbracelet/lipgloss): A Go library for styling terminal applications.

- [Huh](https://github.com/charmbracelet/huh): A Go library for creating interactive terminal forms.



## Want to Help?

Contributions are welcome! Feel free to open an issue or send a pull request to make Chez TUI even better.

## License

This project is under the MIT License. Check out the `LICENSE` file for details.
