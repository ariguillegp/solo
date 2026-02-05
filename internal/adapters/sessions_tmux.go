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
	sessionName, err := sessionNameFor(spec)
	if err != nil {
		return err
	}

	if _, err := ensureSession(sessionName, spec.DirPath, spec.Tool); err != nil {
		return err
	}

	if spec.Detach {
		return nil
	}

	if os.Getenv("TMUX") != "" {
		return switchClient(sessionName)
	}

	cmd := exec.Command("tmux", "attach-session", "-t", tmuxSessionTarget(sessionName))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (t *TmuxSession) PrewarmSession(spec core.SessionSpec) (bool, error) {
	sessionName, err := sessionNameFor(spec)
	if err != nil {
		return false, err
	}
	return ensureSession(sessionName, spec.DirPath, spec.Tool)
}

func (t *TmuxSession) KillSession(spec core.SessionSpec) error {
	sessionName, err := sessionNameFor(spec)
	if err != nil {
		return err
	}

	cmd := exec.Command("tmux", "kill-session", "-t", tmuxSessionTarget(sessionName))
	if output, err := cmd.CombinedOutput(); err != nil {
		if strings.Contains(string(output), "can't find session") ||
			strings.Contains(string(output), "no server running") {
			return nil
		}
		return fmt.Errorf("failed to kill tmux session: %w (output: %s)", err, string(output))
	}
	return nil
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

func sessionNameFor(spec core.SessionSpec) (string, error) {
	if strings.TrimSpace(spec.DirPath) == "" {
		return "", fmt.Errorf("session directory is required")
	}
	if strings.TrimSpace(spec.Tool) == "" {
		return "", fmt.Errorf("session tool is required")
	}

	cleanPath := filepath.Clean(spec.DirPath)
	return strings.Join([]string{
		sanitizeSessionPart(cleanPath, "worktree"),
		sanitizeSessionPart(spec.Tool, "tool"),
	}, "__"), nil
}

func ensureSession(sessionName, dirPath, tool string) (bool, error) {
	check := exec.Command("tmux", "has-session", "-t", tmuxSessionTarget(sessionName))
	if check.Run() == nil {
		return false, nil
	}

	shell, commandArgs := toolCommand(tool)
	args := []string{"new-session", "-d", "-s", sessionName}
	args = append(args, tmuxEnvArgs(tool)...)
	args = append(args, "-c", dirPath, shell)
	args = append(args, commandArgs...)
	cmd := exec.Command("tmux", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("failed to create tmux session: %w (output: %s)", err, string(output))
	}

	return true, nil
}

func switchClient(sessionName string) error {
	args := []string{"switch-client", "-t", tmuxSessionTarget(sessionName)}
	client := currentClientTTY()
	if client != "" {
		args = []string{"switch-client", "-c", client, "-t", tmuxSessionTarget(sessionName)}
	}
	cmd := exec.Command("tmux", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil && client != "" {
		fallback := exec.Command("tmux", "switch-client", "-t", tmuxSessionTarget(sessionName))
		fallback.Stdin = os.Stdin
		fallback.Stdout = os.Stdout
		fallback.Stderr = os.Stderr
		return fallback.Run()
	}
	return nil
}

func tmuxSessionTarget(sessionName string) string {
	return "=" + sessionName
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

func toolCommand(tool string) (string, []string) {
	shell := os.Getenv("SHELL")
	if strings.TrimSpace(shell) == "" {
		shell = "/bin/sh"
	}
	return shell, []string{"-c", `"$1"; exec "$0"`, shell, tool}
}

func tmuxEnvArgs(tool string) []string {
	keys := []string{
		"COLORFGBG",
		"COLORTERM",
		"TERM",
		"TERM_PROGRAM",
		"TERM_PROGRAM_VERSION",
	}
	args := make([]string, 0, len(keys)*2+2)

	for _, env := range core.ToolEnv(tool) {
		args = append(args, "-e", env)
	}

	for _, key := range keys {
		val := strings.TrimSpace(os.Getenv(key))
		if val == "" {
			continue
		}
		args = append(args, "-e", key+"="+val)
	}

	return args
}
