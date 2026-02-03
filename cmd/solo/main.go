package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ariguillegp/solo/internal/adapters"
	"github.com/ariguillegp/solo/internal/core"
	"github.com/ariguillegp/solo/internal/ports"
	"github.com/ariguillegp/solo/internal/ui"
)

func main() {
	var projectFlag string
	var worktreeFlag string
	var toolFlag string
	var createProjectFlag bool
	var detachFlag bool
	flag.StringVar(&projectFlag, "project", "", "Project container name or path")
	flag.StringVar(&worktreeFlag, "worktree", "", "Worktree name or path")
	flag.StringVar(&toolFlag, "tool", "", "Tool to run (opencode or amp)")
	flag.BoolVar(&createProjectFlag, "create-project", false, "Create the project container if missing")
	flag.BoolVar(&detachFlag, "detach", false, "Create the tmux session without attaching")
	flag.Parse()

	roots := flag.Args()
	if len(roots) == 0 {
		roots = []string{"~/Projects"}
	}

	fs := adapters.NewOSFilesystem()
	sessions := adapters.NewTmuxSession()

	if projectFlag != "" || worktreeFlag != "" || toolFlag != "" || createProjectFlag || detachFlag {
		spec, err := resolveSessionSpec(fs, roots, projectFlag, worktreeFlag, toolFlag, createProjectFlag, detachFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if err := sessions.OpenSession(spec); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	m := ui.New(roots, fs, sessions)
	p := tea.NewProgram(m, tea.WithAltScreen())

	result, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	final := result.(ui.Model)
	if final.SelectedSpec != nil {
		if err := resetTerminal(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if err := sessions.OpenSession(*final.SelectedSpec); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}
	_ = returnToPreviousSession()
}

func resetTerminal() error {
	stty := exec.Command("stty", "sane")
	stty.Stdin = os.Stdin
	stty.Stdout = os.Stdout
	stty.Stderr = os.Stderr
	return stty.Run()
}

func returnToPreviousSession() error {
	if os.Getenv("TMUX") == "" {
		return nil
	}
	cmd := exec.Command("tmux", "switch-client", "-l")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func resolveSessionSpec(fs ports.Filesystem, roots []string, project, worktree, tool string, createProject bool, detach bool) (core.SessionSpec, error) {
	if project == "" {
		return core.SessionSpec{}, errors.New("--project is required")
	}
	if worktree == "" {
		return core.SessionSpec{}, errors.New("--worktree is required")
	}
	if tool == "" {
		return core.SessionSpec{}, errors.New("--tool is required")
	}
	if !core.IsSupportedTool(tool) {
		return core.SessionSpec{}, fmt.Errorf("unsupported tool: %s", tool)
	}

	projectPath, err := resolveProjectPath(fs, roots, project, createProject)
	if err != nil {
		return core.SessionSpec{}, err
	}

	worktreePath, err := resolveWorktreePath(fs, projectPath, worktree)
	if err != nil {
		return core.SessionSpec{}, err
	}

	return core.SessionSpec{DirPath: worktreePath, Tool: tool, Detach: detach}, nil
}

func resolveProjectPath(fs ports.Filesystem, roots []string, project string, createProject bool) (string, error) {
	if looksLikePath(project) {
		path := expandPath(project)
		if path == "" {
			return "", fmt.Errorf("invalid project path")
		}
		if !filepath.IsAbs(path) {
			absPath, err := filepath.Abs(path)
			if err != nil {
				return "", fmt.Errorf("cannot resolve path: %w", err)
			}
			path = absPath
		}
		if !exists(path) {
			if createProject {
				createdPath, err := fs.CreateProject(path)
				if err != nil {
					return "", err
				}
				return createdPath, nil
			}
			return "", fmt.Errorf("project not found: %s", project)
		}
		return path, nil
	}

	for _, root := range roots {
		candidate := filepath.Join(expandPath(root), project)
		if exists(candidate) {
			return candidate, nil
		}
	}

	if createProject {
		if len(roots) == 0 {
			return "", fmt.Errorf("no roots available to create project")
		}
		candidate := filepath.Join(expandPath(roots[0]), project)
		createdPath, err := fs.CreateProject(candidate)
		if err != nil {
			return "", err
		}
		return createdPath, nil
	}

	return "", fmt.Errorf("project not found: %s", project)
}

func resolveWorktreePath(fs ports.Filesystem, projectPath, worktree string) (string, error) {
	if looksLikePath(worktree) {
		path := expandPath(worktree)
		if !filepath.IsAbs(path) {
			absPath, err := filepath.Abs(path)
			if err != nil {
				return "", fmt.Errorf("cannot resolve path: %w", err)
			}
			path = absPath
		}
		if exists(path) {
			return path, nil
		}
		return "", fmt.Errorf("worktree not found: %s", worktree)
	}

	listing, err := fs.ListWorktrees(projectPath)
	if err != nil {
		return "", err
	}
	if listing.Warning != "" {
		return "", fmt.Errorf("%s", listing.Warning)
	}

	for _, wt := range listing.Worktrees {
		if wt.Branch == worktree || wt.Name == worktree {
			return wt.Path, nil
		}
	}

	return fs.CreateWorktree(projectPath, worktree)
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		if home == "" {
			return ""
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

func looksLikePath(s string) bool {
	return filepath.IsAbs(s) ||
		strings.HasPrefix(s, "~/") ||
		strings.HasPrefix(s, "./") ||
		strings.HasPrefix(s, "../") ||
		strings.Contains(s, string(filepath.Separator))
}

func exists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
