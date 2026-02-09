package core

import (
	"fmt"
	"path/filepath"
	"strings"
)

func SessionDisplayLabel(session SessionInfo) string {
	worktreeName := SessionWorktreeName(session.DirPath)
	if worktreeName == "" {
		return session.Name
	}

	project, branch := SessionWorktreeProjectBranch(worktreeName)
	if project == "" {
		project = worktreeName
	}
	if branch == "" {
		branch = "main"
	}

	label := fmt.Sprintf("%s--%s", project, branch)
	if session.Tool != "" {
		label = fmt.Sprintf("%s--%s", label, session.Tool)
	}
	if branch != "" {
		label = fmt.Sprintf("%s [%s]", label, branch)
	}

	return label
}

func SessionWorktreeName(dirPath string) string {
	dirPath = strings.TrimSpace(dirPath)
	if dirPath == "" {
		return ""
	}
	cleanPath := filepath.Clean(dirPath)
	parts := strings.Split(cleanPath, string(filepath.Separator))
	worktreeIndex := -1
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] == "worktrees" {
			worktreeIndex = i
			break
		}
	}
	if worktreeIndex == -1 || worktreeIndex == 0 || parts[worktreeIndex-1] != ".solo" {
		return ""
	}
	if len(parts) <= worktreeIndex+2 {
		return ""
	}
	project := strings.TrimSpace(parts[worktreeIndex+1])
	branch := strings.TrimSpace(parts[worktreeIndex+2])
	if project == "" || branch == "" {
		return ""
	}
	return fmt.Sprintf("%s--%s", project, branch)
}

func SessionWorktreeProjectBranch(worktreeName string) (string, string) {
	worktreeName = strings.TrimSpace(worktreeName)
	if worktreeName == "" {
		return "", ""
	}
	if !strings.Contains(worktreeName, "--") {
		return filepath.Base(worktreeName), ""
	}

	parts := strings.Split(worktreeName, "--")
	if len(parts) == 1 {
		return worktreeName, ""
	}

	if len(parts) == 2 {
		project := strings.TrimSpace(parts[0])
		branch := strings.TrimSpace(parts[1])
		return project, branch
	}

	project := strings.TrimSpace(strings.Join(parts[:len(parts)-1], "--"))
	branch := strings.TrimSpace(parts[len(parts)-1])
	return project, branch
}
