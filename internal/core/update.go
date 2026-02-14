package core

import "time"

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
		m.WorktreeWarning = ""
		return m, []Effect{EffLoadWorktrees{ProjectPath: msg.ProjectPath}}

	case MsgProjectDeleted:
		if msg.Err != nil {
			m.Mode = ModeError
			m.Err = msg.Err
			return m, nil
		}
		m.Mode = ModeBrowsing
		m.ProjectDeletePath = ""
		m.SelectedProject = ""
		m.WorktreeQuery = ""
		m.Worktrees = nil
		m.FilteredWT = nil
		m.WorktreeIdx = 0
		m.ProjectWarning = ""
		m.SelectedWorktreePath = ""
		m.WorktreeDeletePath = ""
		m.WorktreeWarning = ""
		return m, []Effect{EffScanDirs{Roots: m.RootPaths}}

	case MsgWorktreesLoaded:
		if msg.Err != nil {
			m.Mode = ModeError
			m.Err = msg.Err
			return m, nil
		}
		m.ProjectWarning = msg.Warning
		m.WorktreeWarning = ""
		m.Worktrees = msg.Worktrees
		m.FilteredWT = FilterWorktrees(m.Worktrees, m.WorktreeQuery)
		m.WorktreeIdx = 0
		return m, nil

	case MsgWorktreeQueryChanged:
		m.WorktreeQuery = msg.Query
		m.WorktreeWarning = ""
		m.FilteredWT = FilterWorktrees(m.Worktrees, m.WorktreeQuery)
		m.WorktreeIdx = 0
		return m, nil

	case MsgWorktreeCreated:
		if msg.Err != nil {
			if IsWorktreeExistsError(msg.Err) {
				m.Mode = ModeWorktree
				m.WorktreeWarning = msg.Err.Error()
				return m, nil
			}
			m.Mode = ModeError
			m.Err = msg.Err
			return m, nil
		}
		m.SelectedWorktreePath = msg.Path
		m.WorktreeWarning = ""
		return enterToolMode(m)

	case MsgWorktreeDeleted:
		if msg.Err != nil {
			if IsRecoverableWorktreeDeleteError(msg.Err) {
				m.Mode = ModeWorktree
				m.WorktreeDeletePath = ""
				m.WorktreeWarning = msg.Err.Error()
				return m, nil
			}
			m.Mode = ModeError
			m.Err = msg.Err
			return m, nil
		}
		m.Mode = ModeWorktree
		m.WorktreeDeletePath = ""
		m.WorktreeWarning = ""
		return m, []Effect{EffLoadWorktrees{ProjectPath: m.SelectedProject}}

	case MsgToolQueryChanged:
		m.ToolQuery = msg.Query
		m.FilteredTools = FilterTools(m.Tools, m.ToolQuery)
		m.ToolIdx = 0
		m.ToolError = ""
		return m, nil

	case MsgToolPrewarmFailed:
		if m.ToolErrors == nil {
			m.ToolErrors = make(map[string]string)
		}
		errText := "prewarm failed"
		if msg.Err != nil {
			errText = msg.Err.Error()
		}
		m.ToolErrors[msg.Tool] = errText
		m.ToolWarmupCompleted++
		m.ToolWarmupFailed++
		if m.Mode == ModeToolStarting && m.PendingSpec != nil && m.PendingSpec.Tool == msg.Tool {
			m.ToolError = errText
			m.PendingSpec = nil
			m.Mode = ModeTool
		}
		return m, nil

	case MsgToolPrewarmStarted:
		if m.ToolWarmStart == nil {
			m.ToolWarmStart = make(map[string]time.Time)
		}
		m.ToolWarmStart[msg.Tool] = msg.StartedAt
		m.ToolWarmupCompleted++
		return m, nil

	case MsgToolPrewarmExisting:
		if m.ToolWarmStart == nil {
			m.ToolWarmStart = make(map[string]time.Time)
		}
		m.ToolWarmStart[msg.Tool] = time.Time{}
		if m.ToolErrors != nil {
			delete(m.ToolErrors, msg.Tool)
		}
		m.ToolWarmupCompleted++
		if m.Mode == ModeToolStarting && m.PendingSpec != nil && m.PendingSpec.Tool == msg.Tool {
			spec := *m.PendingSpec
			m.PendingSpec = nil
			m.ToolError = ""
			return m, []Effect{EffOpenSession{Spec: spec}}
		}
		return m, nil

	case MsgToolDelayElapsed:
		if m.Mode == ModeToolStarting && m.PendingSpec != nil && m.PendingSpec.Tool == msg.Tool {
			spec := *m.PendingSpec
			m.PendingSpec = nil
			m.ToolError = ""
			return m, []Effect{EffOpenSession{Spec: spec}}
		}
		return m, nil

	case MsgSessionsLoaded:
		if msg.Err != nil {
			m.Mode = ModeError
			m.Err = msg.Err
			return m, nil
		}
		m.Sessions = msg.Sessions
		m.FilteredSessions = FilterSessions(m.Sessions, m.SessionQuery)
		m.SessionIdx = 0
		return m, nil

	case MsgSessionQueryChanged:
		m.SessionQuery = msg.Query
		m.FilteredSessions = FilterSessions(m.Sessions, m.SessionQuery)
		m.SessionIdx = 0
		return m, nil

	}

	return m, nil
}

func UpdateKey(m Model, key KeyAction) (Model, []Effect, bool) {
	return handleKey(m, key)
}

func handleKey(m Model, key KeyAction) (Model, []Effect, bool) {
	switch m.Mode {
	case ModeBrowsing:
		return handleBrowsingKey(m, key)
	case ModeProjectDeleteConfirm:
		return handleProjectDeleteConfirmKey(m, key)
	case ModeWorktree:
		return handleWorktreeKey(m, key)
	case ModeWorktreeDeleteConfirm:
		return handleWorktreeDeleteConfirmKey(m, key)
	case ModeTool:
		return handleToolKey(m, key)
	case ModeToolStarting:
		return handleToolStartingKey(m, key)
	case ModeSessions:
		return handleSessionsKey(m, key)
	}
	return m, nil, false
}

func handleBrowsingKey(m Model, key KeyAction) (Model, []Effect, bool) {
	switch key {
	case KeyUp:
		if m.SelectedIdx > 0 {
			m.SelectedIdx--
		}
		return m, nil, true
	case KeyDown:
		maxIdx := len(m.Filtered) - 1
		if _, ok := m.CreateProjectPath(); ok {
			maxIdx = len(m.Filtered)
		}
		if m.SelectedIdx < maxIdx {
			m.SelectedIdx++
		}
		return m, nil, true
	case KeyEnter:
		if dir, ok := m.SelectedDir(); ok {
			m.SelectedProject = dir.Path
			m.Mode = ModeWorktree
			m.WorktreeQuery = ""
			m.WorktreeIdx = 0
			return m, []Effect{EffLoadWorktrees{ProjectPath: dir.Path}}, true
		}
		if path, ok := m.CreateProjectPath(); ok {
			return m, []Effect{EffCreateProject{Path: path}}, true
		}
		return m, nil, true
	case KeyDelete:
		if dir, ok := m.SelectedDir(); ok {
			m.Mode = ModeProjectDeleteConfirm
			m.ProjectDeletePath = dir.Path
			return m, nil, true
		}
		return m, nil, true
	case KeySessions:
		return enterSessionsMode(m)
	case KeyBack, KeyQuit:
		return m, []Effect{EffQuit{}}, true
	}
	return m, nil, false
}

func handleProjectDeleteConfirmKey(m Model, key KeyAction) (Model, []Effect, bool) {
	switch key {
	case KeyEnter:
		if m.ProjectDeletePath != "" {
			return m, []Effect{EffDeleteProject{ProjectPath: m.ProjectDeletePath}}, true
		}
		m.Mode = ModeBrowsing
		return m, nil, true
	case KeyBack:
		m.Mode = ModeBrowsing
		m.ProjectDeletePath = ""
		return m, nil, true
	case KeyQuit:
		return m, []Effect{EffQuit{}}, true
	}
	return m, nil, false
}

func handleWorktreeKey(m Model, key KeyAction) (Model, []Effect, bool) {
	switch key {
	case KeyUp:
		if m.WorktreeIdx > 0 {
			m.WorktreeIdx--
		}
		return m, nil, true
	case KeyDown:
		maxIdx := len(m.FilteredWT) - 1
		if _, ok := m.CreateWorktreeName(); ok {
			maxIdx = len(m.FilteredWT)
		}
		if m.WorktreeIdx < maxIdx {
			m.WorktreeIdx++
		}
		return m, nil, true
	case KeyEnter:
		if wt, ok := m.SelectedWorktree(); ok {
			m.SelectedWorktreePath = wt.Path
			m, effects := enterToolMode(m)
			return m, effects, true
		}
		if name, ok := m.CreateWorktreeName(); ok {
			m.WorktreeWarning = ""
			return m, []Effect{EffCreateWorktree{
				ProjectPath: m.SelectedProject,
				BranchName:  name,
			}}, true
		}
		return m, nil, true
	case KeyDelete:
		if wt, ok := m.SelectedWorktree(); ok {
			m.Mode = ModeWorktreeDeleteConfirm
			m.WorktreeDeletePath = wt.Path
			m.WorktreeWarning = ""
			return m, nil, true
		}
		return m, nil, true
	case KeyBack:
		m.Mode = ModeBrowsing
		m.WorktreeQuery = ""
		m.Worktrees = nil
		m.FilteredWT = nil
		m.WorktreeIdx = 0
		m.ProjectWarning = ""
		m.SelectedWorktreePath = ""
		m.WorktreeDeletePath = ""
		m.WorktreeWarning = ""
		return m, nil, true
	case KeySessions:
		return enterSessionsMode(m)
	case KeyQuit:
		return m, []Effect{EffQuit{}}, true
	}
	return m, nil, false
}

func handleWorktreeDeleteConfirmKey(m Model, key KeyAction) (Model, []Effect, bool) {
	switch key {
	case KeyEnter:
		if m.WorktreeDeletePath != "" {
			return m, []Effect{EffDeleteWorktree{
				ProjectPath:  m.SelectedProject,
				WorktreePath: m.WorktreeDeletePath,
			}}, true
		}
		m.Mode = ModeWorktree
		return m, nil, true
	case KeyBack:
		m.Mode = ModeWorktree
		m.WorktreeDeletePath = ""
		return m, nil, true
	case KeyQuit:
		return m, []Effect{EffQuit{}}, true
	}
	return m, nil, false
}

func handleToolKey(m Model, key KeyAction) (Model, []Effect, bool) {
	switch key {
	case KeyUp:
		if m.ToolIdx > 0 {
			m.ToolIdx--
		}
		m.ToolError = ""
		return m, nil, true
	case KeyDown:
		if m.ToolIdx < len(m.FilteredTools)-1 {
			m.ToolIdx++
		}
		m.ToolError = ""
		return m, nil, true
	case KeyEnter:
		if tool, ok := m.SelectedTool(); ok && m.SelectedWorktreePath != "" {
			if errText, ok := m.ToolErrors[tool]; ok && errText != "" {
				m.ToolError = errText
				delete(m.ToolErrors, tool)
				return m, nil, true
			}

			spec := SessionSpec{
				DirPath: m.SelectedWorktreePath,
				Tool:    tool,
			}
			if !ToolNeedsWarmup(tool) {
				m.PendingSpec = nil
				m.ToolError = ""
				return m, []Effect{EffOpenSession{Spec: spec}}, true
			}
			m.PendingSpec = &spec
			m.Mode = ModeToolStarting
			m.ToolError = ""
			return m, []Effect{EffCheckToolReady{Spec: spec}}, true
		}
		return m, nil, true
	case KeyBack:
		m.Mode = ModeWorktree
		m.ToolQuery = ""
		m.ToolIdx = 0
		m.ToolError = ""
		return m, nil, true
	case KeySessions:
		return enterSessionsMode(m)
	case KeyQuit:
		return m, []Effect{EffQuit{}}, true
	}
	return m, nil, false
}

func handleToolStartingKey(m Model, key KeyAction) (Model, []Effect, bool) {
	switch key {
	case KeyBack:
		m.Mode = ModeTool
		m.PendingSpec = nil
		m.ToolError = ""
		return m, nil, true
	case KeyQuit:
		return m, []Effect{EffQuit{}}, true
	}
	return m, nil, false
}

func handleSessionsKey(m Model, key KeyAction) (Model, []Effect, bool) {
	switch key {
	case KeyUp:
		if m.SessionIdx > 0 {
			m.SessionIdx--
		}
		return m, nil, true
	case KeyDown:
		if m.SessionIdx < len(m.FilteredSessions)-1 {
			m.SessionIdx++
		}
		return m, nil, true
	case KeyEnter:
		if session, ok := m.SelectedSession(); ok {
			return m, []Effect{EffAttachSession{Session: session}}, true
		}
		return m, nil, true
	case KeyBack:
		return leaveSessionsMode(m)
	case KeyQuit:
		return m, []Effect{EffQuit{}}, true
	}
	return m, nil, false
}

func enterToolMode(m Model) (Model, []Effect) {
	m.Mode = ModeTool
	m.ToolQuery = ""
	m.FilteredTools = FilterTools(m.Tools, m.ToolQuery)
	m.ToolIdx = 0
	m.ToolError = ""
	m.PendingSpec = nil
	m.ToolWarmStart = make(map[string]time.Time, len(m.Tools))
	m.ToolErrors = make(map[string]string, len(m.Tools))
	warmupTools := toolsNeedingWarmup(m.Tools)
	m.ToolWarmupTotal = len(warmupTools)
	m.ToolWarmupCompleted = 0
	m.ToolWarmupFailed = 0
	return m, []Effect{EffPrewarmAllTools{DirPath: m.SelectedWorktreePath, Tools: warmupTools}}
}

func toolsNeedingWarmup(tools []string) []string {
	if len(tools) == 0 {
		return nil
	}
	filtered := make([]string, 0, len(tools))
	for _, tool := range tools {
		if ToolNeedsWarmup(tool) {
			filtered = append(filtered, tool)
		}
	}
	return filtered
}

func enterSessionsMode(m Model) (Model, []Effect, bool) {
	m.SessionReturnMode = m.Mode
	m.Mode = ModeSessions
	m.SessionQuery = ""
	m.SessionIdx = 0
	m.Sessions = nil
	m.FilteredSessions = nil
	return m, []Effect{EffListSessions{}}, true
}

func leaveSessionsMode(m Model) (Model, []Effect, bool) {
	m.Mode = m.SessionReturnMode
	m.SessionQuery = ""
	m.SessionIdx = 0
	m.Sessions = nil
	m.FilteredSessions = nil
	return m, nil, true
}

func Init(m Model) (Model, []Effect) {
	return m, []Effect{EffScanDirs{Roots: m.RootPaths}}
}
