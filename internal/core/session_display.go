package core

import (
	"fmt"
	"path/filepath"
	"strings"
)

const soloWorktreesMarker = "worktrees-"

func SessionDisplayLabel(session SessionInfo) string {
	worktreeName := SessionWorktreeName(session.DirPath)
	if worktreeName == "" {
		return session.Name
	}

	project, branch := SessionWorktreeProjectBranch(worktreeName)
	project = trimProjectHash(project)
	label := strings.TrimSpace(branch)
	if strings.TrimSpace(project) != "" && strings.TrimSpace(branch) != "" {
		label = fmt.Sprintf("%s/%s", project, branch)
	}
	if label == "" {
		label = worktreeName
	}
	if session.Tool != "" {
		label = fmt.Sprintf("%s - %s", label, session.Tool)
	}

	return label
}

func SessionWorktreeName(dirPath string) string {
	dirPath = strings.TrimSpace(dirPath)
	if dirPath == "" {
		return ""
	}

	if strings.Contains(dirPath, string(filepath.Separator)) {
		base := filepath.Base(dirPath)
		if base != "." && base != string(filepath.Separator) {
			return base
		}
	}

	if idx := strings.LastIndex(dirPath, soloWorktreesMarker); idx >= 0 {
		name := strings.TrimSpace(dirPath[idx+len(soloWorktreesMarker):])
		if name != "" {
			return name
		}
	}

	return dirPath
}

func SessionWorktreeProjectBranch(worktreeName string) (string, string) {
	worktreeName = strings.TrimSpace(worktreeName)
	if worktreeName == "" {
		return "", ""
	}

	idx := strings.LastIndex(worktreeName, "--")
	if idx == -1 {
		return worktreeName, ""
	}

	project := strings.TrimSpace(worktreeName[:idx])
	branch := strings.TrimSpace(worktreeName[idx+len("--"):])
	return project, branch
}

func trimProjectHash(project string) string {
	project = strings.TrimSpace(project)
	if len(project) < 8 {
		return project
	}
	sepIdx := len(project) - 7
	if project[sepIdx] != '-' {
		return project
	}
	suffix := project[sepIdx+1:]
	for i := 0; i < len(suffix); i++ {
		ch := suffix[i]
		if (ch < '0' || ch > '9') && (ch < 'a' || ch > 'f') {
			return project
		}
	}
	return project[:sepIdx]
}
