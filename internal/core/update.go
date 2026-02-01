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
		model, effects, _ := handleKey(m, msg.Key)
		return model, effects

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

	case MsgWorktreeDeleted:
		if msg.Err != nil {
			m.Mode = ModeError
			m.Err = msg.Err
			return m, nil
		}
		m.Mode = ModeWorktree
		m.WorktreeDeletePath = ""
		return m, []Effect{EffLoadWorktrees{ProjectPath: m.SelectedProject}}

	case MsgToolQueryChanged:
		m.ToolQuery = msg.Query
		m.FilteredTools = FilterTools(m.Tools, m.ToolQuery)
		m.ToolIdx = 0
		return m, nil
	}

	return m, nil
}

func UpdateKey(m Model, key string) (Model, []Effect, bool) {
	return handleKey(m, key)
}

func handleKey(m Model, key string) (Model, []Effect, bool) {
	switch m.Mode {
	case ModeBrowsing:
		return handleBrowsingKey(m, key)
	case ModeWorktree:
		return handleWorktreeKey(m, key)
	case ModeWorktreeDeleteConfirm:
		return handleWorktreeDeleteConfirmKey(m, key)
	case ModeTool:
		return handleToolKey(m, key)
	}
	return m, nil, false
}

func handleBrowsingKey(m Model, key string) (Model, []Effect, bool) {
	switch key {
	case "up", "ctrl+k":
		if m.SelectedIdx > 0 {
			m.SelectedIdx--
		}
		return m, nil, true
	case "down", "ctrl+j":
		if m.SelectedIdx < len(m.Filtered)-1 {
			m.SelectedIdx++
		}
		return m, nil, true
	case "enter":
		if dir, ok := m.SelectedDir(); ok {
			m.SelectedProject = dir.Path
			m.Mode = ModeWorktree
			m.WorktreeQuery = ""
			m.WorktreeIdx = 0
			return m, []Effect{EffLoadWorktrees{ProjectPath: dir.Path}}, true
		}
		if m.Query != "" && len(m.RootPaths) > 0 {
			path := m.RootPaths[0] + "/" + m.Query
			return m, []Effect{EffCreateProject{Path: path}}, true
		}
		return m, nil, true
	case "ctrl+n":
		if m.Query != "" && len(m.RootPaths) > 0 {
			path := m.RootPaths[0] + "/" + m.Query
			return m, []Effect{EffCreateProject{Path: path}}, true
		}
		return m, nil, true
	case "esc", "ctrl+c":
		return m, []Effect{EffQuit{}}, true
	}
	return m, nil, false
}

func handleWorktreeKey(m Model, key string) (Model, []Effect, bool) {
	switch key {
	case "up", "ctrl+k":
		if m.WorktreeIdx > 0 {
			m.WorktreeIdx--
		}
		return m, nil, true
	case "down", "ctrl+j":
		if m.WorktreeIdx < len(m.FilteredWT)-1 {
			m.WorktreeIdx++
		}
		return m, nil, true
	case "enter":
		if wt, ok := m.SelectedWorktree(); ok {
			m.SelectedWorktreePath = wt.Path
			m.Mode = ModeTool
			m.ToolQuery = ""
			m.FilteredTools = FilterTools(m.Tools, m.ToolQuery)
			m.ToolIdx = 0
			return m, nil, true
		}
		if m.WorktreeQuery != "" {
			return m, []Effect{EffCreateWorktree{
				ProjectPath: m.SelectedProject,
				BranchName:  m.WorktreeQuery,
			}}, true
		}
		return m, nil, true
	case "ctrl+n":
		if m.WorktreeQuery != "" {
			return m, []Effect{EffCreateWorktree{
				ProjectPath: m.SelectedProject,
				BranchName:  m.WorktreeQuery,
			}}, true
		}
		return m, nil, true
	case "ctrl+d":
		if wt, ok := m.SelectedWorktree(); ok {
			m.Mode = ModeWorktreeDeleteConfirm
			m.WorktreeDeletePath = wt.Path
			return m, nil, true
		}
		return m, nil, true
	case "esc":
		m.Mode = ModeBrowsing
		m.WorktreeQuery = ""
		m.Worktrees = nil
		m.FilteredWT = nil
		m.WorktreeIdx = 0
		m.ProjectWarning = ""
		m.SelectedWorktreePath = ""
		m.WorktreeDeletePath = ""
		return m, nil, true
	case "ctrl+c":
		return m, []Effect{EffQuit{}}, true
	}
	return m, nil, false
}

func handleWorktreeDeleteConfirmKey(m Model, key string) (Model, []Effect, bool) {
	switch key {
	case "enter":
		if m.WorktreeDeletePath != "" {
			return m, []Effect{EffDeleteWorktree{
				ProjectPath:  m.SelectedProject,
				WorktreePath: m.WorktreeDeletePath,
			}}, true
		}
		m.Mode = ModeWorktree
		return m, nil, true
	case "esc":
		m.Mode = ModeWorktree
		m.WorktreeDeletePath = ""
		return m, nil, true
	case "ctrl+c":
		return m, []Effect{EffQuit{}}, true
	}
	return m, nil, false
}

func handleToolKey(m Model, key string) (Model, []Effect, bool) {
	switch key {
	case "up", "ctrl+k":
		if m.ToolIdx > 0 {
			m.ToolIdx--
		}
		return m, nil, true
	case "down", "ctrl+j":
		if m.ToolIdx < len(m.FilteredTools)-1 {
			m.ToolIdx++
		}
		return m, nil, true
	case "enter":
		if tool, ok := m.SelectedTool(); ok && m.SelectedWorktreePath != "" {
			spec := SessionSpec{
				DirPath: m.SelectedWorktreePath,
				Tool:    tool,
			}
			return m, []Effect{EffOpenSession{Spec: spec}}, true
		}
		return m, nil, true
	case "esc":
		m.Mode = ModeWorktree
		m.ToolQuery = ""
		m.ToolIdx = 0
		return m, nil, true
	case "ctrl+c":
		return m, []Effect{EffQuit{}}, true
	}
	return m, nil, false
}

func Init(m Model) (Model, []Effect) {
	return m, []Effect{EffScanDirs{Roots: m.RootPaths}}
}
