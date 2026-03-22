//go:build linux

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"shelf/internal/fs"
	"shelf/internal/model"
)

func main() {

	cfg, err := fs.LoadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	mode := ""
	if len(os.Args) >= 2 {
		mode = os.Args[1]
		if mode != "ctf" && mode != "box" {
			fmt.Fprintf(os.Stderr, "Usage: shelf {ctf|box}\n")
			os.Exit(1)
		}
	}

	p := tea.NewProgram(model.New(mode, cfg), tea.WithAltScreen())
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
		sessionName := filepath.Base(fm.SelectedPath)

		cmd := cfg.ExpandCmd(sessionName, fm.SelectedPath)
		if err := exec.Command("sh", "-c", cmd).Run(); err != nil {
			fmt.Fprintf(os.Stderr, "cmd: %v\n", err)
		}
	}
}
