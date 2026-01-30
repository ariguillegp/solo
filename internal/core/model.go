package core

type Mode int

const (
	ModeLoading Mode = iota
	ModeBrowsing
	ModeWorktree
	ModeTool
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
	ProjectWarning       string
	Worktrees            []Worktree
	FilteredWT           []Worktree
	WorktreeIdx          int
	WorktreeQuery        string
	Tools                []string
	ToolIdx              int
}

func NewModel(roots []string) Model {
	return Model{
		Mode:      ModeLoading,
		RootPaths: roots,
		Tools:     SupportedTools(),
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
	if len(m.Tools) == 0 || m.ToolIdx >= len(m.Tools) {
		return "", false
	}
	return m.Tools[m.ToolIdx], true
}
