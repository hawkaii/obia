package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hawkaii/obia/internal/config"
	"github.com/hawkaii/obia/internal/tui"
	"github.com/hawkaii/obia/internal/version"
)

func main() {
	for _, arg := range os.Args[1:] {
		if arg == "--version" || arg == "-v" {
			commit := version.Commit
			if len(commit) > 7 {
				commit = commit[:7]
			}
			fmt.Printf("obia %s (%s, %s)\n", version.Version, commit, version.Date)
			return
		}
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	app := tui.NewApp(cfg)
	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
