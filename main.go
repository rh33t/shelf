package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"shelf/internal/fs"
	"shelf/internal/model"
)

func main() {
	mode := ""
	if len(os.Args) >= 2 {
		mode = os.Args[1]
		if mode != "ctf" && mode != "box" {
			fmt.Fprintf(os.Stderr, "Usage: shelf {ctf|box}\n")
			os.Exit(1)
		}
	}

	p := tea.NewProgram(model.New(mode), tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fm := final.(model.Model)
	if fm.Err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", fm.Err)
		os.Exit(1)
	}

	if fm.SelectedPath != "" {
		// Derive the tmux session name from the selected directory name.
		sessionName := filepath.Base(fm.SelectedPath)

		// Spawn a tmux session rooted at the selected path.
		// If not inside tmux, print the attach command for the caller to run.
		if err := fs.SpawnTmux(sessionName, fm.SelectedPath); err != nil {
			fmt.Fprintf(os.Stderr, "tmux: %v\n", err)
		} else if os.Getenv("TMUX") == "" {
			fmt.Printf("tmux attach -t %s\n", sessionName)
		}
	}
}
