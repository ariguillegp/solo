package core

import (
	"path/filepath"
	"strings"
	"time"
)

type Mode int

const (
	ModeLoading Mode = iota
	ModeBrowsing
	ModeProjectDeleteConfirm
	ModeWorktree
	ModeWorktreeDeleteConfirm
	ModeTool
	ModeToolStarting
	ModeSessions
	ModeError
)

type Model struct {
	Mode                 Mode
	Query                string
	Dirs                 []DirEntry
	Filtered             []DirEntry
	SelectedIdx          int
	RootPaths            []string
	Err                  error
	SelectedProject      string
	SelectedWorktreePath string
	WorktreeDeletePath   string
	ProjectDeletePath    string
	ProjectWarning       string
	Worktrees            []Worktree
	FilteredWT           []Worktree
	WorktreeIdx          int
	WorktreeQuery        string
	Tools                []string
	FilteredTools        []string
	ToolQuery            string
	ToolIdx              int
	ToolWarmStart        map[string]time.Time
	ToolErrors           map[string]string
	ToolError            string
	PendingSpec          *SessionSpec
	SessionReturnMode    Mode
	Sessions             []SessionInfo
	FilteredSessions     []SessionInfo
	SessionQuery         string
	SessionIdx           int
}

func NewModel(roots []string) Model {
	tools := SupportedTools()
	return Model{
		Mode:          ModeLoading,
		RootPaths:     roots,
		Tools:         tools,
		FilteredTools: tools,
	}
}

func (m Model) SelectedDir() (DirEntry, bool) {
	if len(m.Filtered) == 0 || m.SelectedIdx >= len(m.Filtered) {
		return DirEntry{}, false
	}
	return m.Filtered[m.SelectedIdx], true
}

func (m Model) SelectedWorktree() (Worktree, bool) {
	if len(m.FilteredWT) == 0 || m.WorktreeIdx >= len(m.FilteredWT) {
		return Worktree{}, false
	}
	return m.FilteredWT[m.WorktreeIdx], true
}

func (m Model) SelectedTool() (string, bool) {
	if len(m.FilteredTools) == 0 || m.ToolIdx >= len(m.FilteredTools) {
		return "", false
	}
	return m.FilteredTools[m.ToolIdx], true
}

func (m Model) SelectedSession() (SessionInfo, bool) {
	if len(m.FilteredSessions) == 0 || m.SessionIdx >= len(m.FilteredSessions) {
		return SessionInfo{}, false
	}
	return m.FilteredSessions[m.SessionIdx], true
}

func (m Model) CreateProjectPath() (string, bool) {
	if m.Query == "" || len(m.RootPaths) == 0 {
		return "", false
	}
	query := strings.TrimSpace(m.Query)
	if query == "" || filepath.IsAbs(query) {
		return "", false
	}
	path := filepath.Join(m.RootPaths[0], query)
	for _, dir := range m.Dirs {
		if dir.Path == path {
			return "", false
		}
	}
	return path, true
}

func (m Model) CreateWorktreeName() (string, bool) {
	name := strings.TrimSpace(m.WorktreeQuery)
	if name == "" {
		return "", false
	}
	sanitized := SanitizeWorktreeName(name)
	if sanitized == "" {
		return "", false
	}
	for _, wt := range m.Worktrees {
		if wt.Branch == name {
			return "", false
		}
		if sanitized != "" && SanitizeWorktreeName(wt.Branch) == sanitized {
			return "", false
		}
	}
	return name, true
}
