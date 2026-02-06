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

	parts := strings.Split(worktreeName, "--")
	if len(parts) == 1 {
		return worktreeName, ""
	}

	if len(parts) == 2 {
		project := strings.TrimSpace(parts[0])
		branch := strings.TrimSpace(parts[1])
		return project, branch
	}

	if isShortUUIDPart(parts[len(parts)-1]) {
		branch := strings.TrimSpace(parts[len(parts)-2])
		project := strings.TrimSpace(strings.Join(parts[:len(parts)-2], "--"))
		return project, branch
	}

	project := strings.TrimSpace(strings.Join(parts[:len(parts)-1], "--"))
	branch := strings.TrimSpace(parts[len(parts)-1])
	return project, branch
}

func isShortUUIDPart(part string) bool {
	if len(part) != 8 {
		return false
	}
	for i := 0; i < len(part); i++ {
		ch := part[i]
		if (ch < '0' || ch > '9') && (ch < 'a' || ch > 'f') {
			return false
		}
	}
	return true
}
