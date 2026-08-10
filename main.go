//go:build linux

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"

	"shelf/internal/fs"
	"shelf/internal/model"
)

// version is set at build time via -ldflags "-X main.version=...".
var version = "dev"

const usage = `Usage: shelf [ctf|box]

  ctf, box     start in that mode, otherwise select interactively

  -h, --help   show this help
  --version    show version
  --config     show the config file path`

func main() {
	mode := ""
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "ctf", "box":
			mode = os.Args[1]
		case "-h", "--help":
			fmt.Println(usage)
			return
		case "--version":
			fmt.Println(version)
			return
		case "--config":
			path, err := fs.Path()
			if err != nil {
				fail(err)
			}
			fmt.Println(path)
			return
		default:
			fmt.Fprintln(os.Stderr, usage)
			os.Exit(1)
		}
	}

	cfg, err := fs.LoadConfig()
	if err != nil {
		fail(err)
	}

	p := tea.NewProgram(model.New(mode, cfg), tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		fail(err)
	}

	fm := final.(model.Model)
	if fm.Err != nil {
		fail(fm.Err)
	}
	if fm.SelectedPath == "" {
		return
	}

	// Replace this process so the command inherits the terminal. Without it an
	// interactive command (tmux attach, an editor) has no tty to draw on.
	argv := cfg.Command(filepath.Base(fm.SelectedPath), fm.SelectedPath)
	sh, err := exec.LookPath(argv[0])
	if err != nil {
		fail(err)
	}
	fail(syscall.Exec(sh, argv, os.Environ()))
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "shelf: %v\n", err)
	os.Exit(1)
}
