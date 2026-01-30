package adapters

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ariguillegp/solo/internal/core"
)

type TmuxSession struct{}

func NewTmuxSession() *TmuxSession {
	return &TmuxSession{}
}

func (t *TmuxSession) OpenSession(spec core.SessionSpec) error {
	if strings.TrimSpace(spec.DirPath) == "" {
		return fmt.Errorf("session directory is required")
	}
	if strings.TrimSpace(spec.Tool) == "" {
		return fmt.Errorf("session tool is required")
	}

	projectName, worktreeName := sessionParts(spec.DirPath)
	sessionName := strings.Join([]string{
		sanitizeSessionPart(projectName, "project"),
		sanitizeSessionPart(worktreeName, "worktree"),
		sanitizeSessionPart(spec.Tool, "tool"),
	}, "__")

	if spec.Detach {
		return ensureSession(sessionName, spec.DirPath, spec.Tool)
	}

	if os.Getenv("TMUX") != "" {
		if err := ensureSession(sessionName, spec.DirPath, spec.Tool); err != nil {
			return err
		}
		return switchClient(sessionName)
	}

	command := wrapToolCommand(spec.Tool)
	args := []string{"new-session", "-A", "-s", sessionName, "-c", spec.DirPath, command.shell, "-l", "-c", command.cmd}
	cmd := exec.Command("tmux", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func sessionParts(dirPath string) (string, string) {
	cleanPath := filepath.Clean(dirPath)
	worktreeName := filepath.Base(cleanPath)
	parent := filepath.Dir(cleanPath)
	projectName := filepath.Base(parent)
	if projectName == "." || projectName == string(filepath.Separator) || projectName == "" {
		projectName = worktreeName
	}
	return projectName, worktreeName
}

var sessionNamePattern = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)

func sanitizeSessionPart(name, fallback string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return fallback
	}
	name = sessionNamePattern.ReplaceAllString(name, "-")
	if name == "" {
		return fallback
	}
	return name
}

func ensureSession(sessionName, dirPath, tool string) error {
	check := exec.Command("tmux", "has-session", "-t", sessionName)
	if check.Run() == nil {
		return nil
	}

	command := wrapToolCommand(tool)
	args := []string{"new-session", "-d", "-s", sessionName, "-c", dirPath, command.shell, "-l", "-c", command.cmd}
	cmd := exec.Command("tmux", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type shellCommand struct {
	shell string
	cmd   string
}

func wrapToolCommand(tool string) shellCommand {
	shell := os.Getenv("SHELL")
	if strings.TrimSpace(shell) == "" {
		shell = "/bin/sh"
	}
	return shellCommand{
		shell: shell,
		cmd:   fmt.Sprintf("%s; exec %s", tool, shell),
	}
}

func switchClient(sessionName string) error {
	args := []string{"switch-client", "-t", sessionName}
	client := currentClientTTY()
	if client != "" {
		args = []string{"switch-client", "-c", client, "-t", sessionName}
	}
	cmd := exec.Command("tmux", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil && client != "" {
		fallback := exec.Command("tmux", "switch-client", "-t", sessionName)
		fallback.Stdin = os.Stdin
		fallback.Stdout = os.Stdout
		fallback.Stderr = os.Stderr
		return fallback.Run()
	}
	return nil
}

func currentClientTTY() string {
	pane := os.Getenv("TMUX_PANE")
	if pane == "" {
		return ""
	}
	cmd := exec.Command("tmux", "display-message", "-p", "-t", pane, "#{client_tty}")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}
