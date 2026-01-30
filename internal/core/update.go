package core

func Update(m Model, msg Msg) (Model, []Effect) {
	switch msg := msg.(type) {
	case MsgScanCompleted:
		if msg.Err != nil {
			m.Mode = ModeError
			m.Err = msg.Err
			return m, nil
		}
		m.Mode = ModeBrowsing
		m.Dirs = msg.Dirs
		m.Filtered = FilterDirs(m.Dirs, m.Query)
		m.SelectedIdx = 0
		return m, nil

	case MsgQueryChanged:
		m.Query = msg.Query
		m.Filtered = FilterDirs(m.Dirs, m.Query)
		m.SelectedIdx = 0
		return m, nil

	case MsgKeyPress:
		return handleKey(m, msg.Key)

	case MsgProjectCreated:
		if msg.Err != nil {
			m.Mode = ModeError
			m.Err = msg.Err
			return m, nil
		}
		m.SelectedProject = msg.ProjectPath
		m.Mode = ModeWorktree
		m.WorktreeQuery = ""
		m.WorktreeIdx = 0
		m.ProjectWarning = ""
		return m, []Effect{EffLoadWorktrees{ProjectPath: msg.ProjectPath}}

	case MsgWorktreesLoaded:
		if msg.Err != nil {
			m.Mode = ModeError
			m.Err = msg.Err
			return m, nil
		}
		m.ProjectWarning = msg.Warning
		m.Worktrees = msg.Worktrees
		m.FilteredWT = FilterWorktrees(m.Worktrees, m.WorktreeQuery)
		m.WorktreeIdx = 0
		return m, nil

	case MsgWorktreeQueryChanged:
		m.WorktreeQuery = msg.Query
		m.FilteredWT = FilterWorktrees(m.Worktrees, m.WorktreeQuery)
		m.WorktreeIdx = 0
		return m, nil

	case MsgWorktreeCreated:
		if msg.Err != nil {
			m.Mode = ModeError
			m.Err = msg.Err
			return m, nil
		}
		m.SelectedWorktreePath = msg.Path
		m.Mode = ModeTool
		m.ToolQuery = ""
		m.FilteredTools = FilterTools(m.Tools, m.ToolQuery)
		m.ToolIdx = 0
		return m, nil

	case MsgToolQueryChanged:
		m.ToolQuery = msg.Query
		m.FilteredTools = FilterTools(m.Tools, m.ToolQuery)
		m.ToolIdx = 0
		return m, nil
	}

	return m, nil
}

func handleKey(m Model, key string) (Model, []Effect) {
	switch m.Mode {
	case ModeBrowsing:
		return handleBrowsingKey(m, key)
	case ModeWorktree:
		return handleWorktreeKey(m, key)
	case ModeTool:
		return handleToolKey(m, key)
	}
	return m, nil
}

func handleBrowsingKey(m Model, key string) (Model, []Effect) {
	switch key {
	case "up", "ctrl+k":
		if m.SelectedIdx > 0 {
			m.SelectedIdx--
		}
	case "down", "ctrl+j":
		if m.SelectedIdx < len(m.Filtered)-1 {
			m.SelectedIdx++
		}
	case "enter":
		if dir, ok := m.SelectedDir(); ok {
			m.SelectedProject = dir.Path
			m.Mode = ModeWorktree
			m.WorktreeQuery = ""
			m.WorktreeIdx = 0
			return m, []Effect{EffLoadWorktrees{ProjectPath: dir.Path}}
		}
		if m.Query != "" && len(m.RootPaths) > 0 {
			path := m.RootPaths[0] + "/" + m.Query
			return m, []Effect{EffCreateProject{Path: path}}
		}
	case "ctrl+n":
		if m.Query != "" && len(m.RootPaths) > 0 {
			path := m.RootPaths[0] + "/" + m.Query
			return m, []Effect{EffCreateProject{Path: path}}
		}
	case "esc", "ctrl+c":
		return m, []Effect{EffQuit{}}
	}
	return m, nil
}

func handleWorktreeKey(m Model, key string) (Model, []Effect) {
	switch key {
	case "up", "ctrl+k":
		if m.WorktreeIdx > 0 {
			m.WorktreeIdx--
		}
	case "down", "ctrl+j":
		if m.WorktreeIdx < len(m.FilteredWT)-1 {
			m.WorktreeIdx++
		}
	case "enter":
		if wt, ok := m.SelectedWorktree(); ok {
			m.SelectedWorktreePath = wt.Path
			m.Mode = ModeTool
			m.ToolQuery = ""
			m.FilteredTools = FilterTools(m.Tools, m.ToolQuery)
			m.ToolIdx = 0
			return m, nil
		}
		if m.WorktreeQuery != "" {
			return m, []Effect{EffCreateWorktree{
				ProjectPath: m.SelectedProject,
				BranchName:  m.WorktreeQuery,
			}}
		}
	case "ctrl+n":
		if m.WorktreeQuery != "" {
			return m, []Effect{EffCreateWorktree{
				ProjectPath: m.SelectedProject,
				BranchName:  m.WorktreeQuery,
			}}
		}
	case "esc":
		m.Mode = ModeBrowsing
		m.WorktreeQuery = ""
		m.Worktrees = nil
		m.FilteredWT = nil
		m.WorktreeIdx = 0
		m.ProjectWarning = ""
		m.SelectedWorktreePath = ""
	case "ctrl+c":
		return m, []Effect{EffQuit{}}
	}
	return m, nil
}

func handleToolKey(m Model, key string) (Model, []Effect) {
	switch key {
	case "up", "ctrl+k":
		if m.ToolIdx > 0 {
			m.ToolIdx--
		}
	case "down", "ctrl+j":
		if m.ToolIdx < len(m.FilteredTools)-1 {
			m.ToolIdx++
		}
	case "enter":
		if tool, ok := m.SelectedTool(); ok && m.SelectedWorktreePath != "" {
			spec := SessionSpec{
				DirPath: m.SelectedWorktreePath,
				Tool:    tool,
			}
			return m, []Effect{EffOpenSession{Spec: spec}}
		}
	case "esc":
		m.Mode = ModeWorktree
		m.ToolQuery = ""
		m.ToolIdx = 0
	case "ctrl+c":
		return m, []Effect{EffQuit{}}
	}
	return m, nil
}

func Init(m Model) (Model, []Effect) {
	return m, []Effect{EffScanDirs{Roots: m.RootPaths}}
}
