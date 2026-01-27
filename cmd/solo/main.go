package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ariguillegp/solo/internal/adapters"
	"github.com/ariguillegp/solo/internal/ui"
)

func main() {
	roots := []string{"~/Work/tries"}

	if len(os.Args) > 1 {
		roots = os.Args[1:]
	}

	fs := adapters.NewOSFilesystem()
	sessions := adapters.NewNativeSession()

	m := ui.New(roots, fs, sessions)
	p := tea.NewProgram(m, tea.WithAltScreen())

	result, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	final := result.(ui.Model)
	if final.SelectedDir != "" {
		openShell(final.SelectedDir)
	}
}

func openShell(dir string) {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	// Reset terminal state
	stty := exec.Command("stty", "sane")
	stty.Stdin = os.Stdin
	stty.Run()

	if err := os.Chdir(dir); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	syscall.Exec(shell, []string{shell}, os.Environ())
}
