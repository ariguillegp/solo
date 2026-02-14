package ui

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ariguillegp/rivet/internal/core"

	"github.com/ariguillegp/rivet/internal/ports"
	"github.com/ariguillegp/rivet/internal/ui/listmodel"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	core                 core.Model
	input                textinput.Model
	worktreeInput        textinput.Model
	toolInput            textinput.Model
	sessionInput         textinput.Model
	themeInput           textinput.Model
	projectList          listmodel.Model
	worktreeList         listmodel.Model
	toolList             listmodel.Model
	sessionList          listmodel.Model
	themeList            listmodel.Model
	spinner              spinner.Model
	progress             progress.Model
	toolStartingAt       time.Time
	toolStartingDuration time.Duration
	fs                   ports.Filesystem
	sessions             ports.SessionManager
	maxDepth             int
	width                int
	height               int
	SelectedSpec         *core.SessionSpec
	SelectedSessionName  string
	styles               Styles
	themes               []Theme
	filteredThemes       []Theme
	activeThemeIdx       int
	themePickerPrevIdx   int
	showHelp             bool
	showThemePicker      bool
	homeDir              string
	help                 help.Model
	keymap               keyMap
}

func New(roots []string, fs ports.Filesystem, sessions ports.SessionManager) Model {
	ti := textinput.New()
	ti.Prompt = ""
	ti.Focus()

	wti := textinput.New()
	wti.Prompt = ""
	tti := textinput.New()
	tti.Prompt = ""
	sti := textinput.New()
	sti.Prompt = ""
	thi := textinput.New()
	thi.Prompt = ""

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	pr := progress.New()

	homeDir, _ := os.UserHomeDir()

	allThemes := Themes()
	styles := NewStyles(allThemes[0])
	h := newHelpModel()
	km := newKeyMap()
	m := Model{
		core:               core.NewModel(roots),
		input:              ti,
		worktreeInput:      wti,
		toolInput:          tti,
		sessionInput:       sti,
		themeInput:         thi,
		projectList:        newSuggestionList(styles),
		worktreeList:       newSuggestionList(styles),
		toolList:           newSuggestionList(styles),
		sessionList:        newSuggestionList(styles),
		themeList:          newSuggestionList(styles),
		spinner:            sp,
		progress:           pr,
		fs:                 fs,
		sessions:           sessions,
		maxDepth:           2,
		styles:             styles,
		themes:             allThemes,
		filteredThemes:     allThemes,
		activeThemeIdx:     0,
		themePickerPrevIdx: 0,
		homeDir:            homeDir,
		help:               h,
		keymap:             km,
	}
	m.syncProgressTheme(allThemes[0])
	m.applyHelpStyles()
	m.applyListStyles()
	return m
}

func (m *Model) blurInputs() {
	m.input.Blur()
	m.worktreeInput.Blur()
	m.toolInput.Blur()
	m.sessionInput.Blur()
	m.themeInput.Blur()
}

func (m *Model) restoreInputFocus() {
	switch m.core.Mode {
	case core.ModeBrowsing:
		m.input.Focus()
	case core.ModeWorktree:
		m.worktreeInput.Focus()
	case core.ModeTool:
		m.toolInput.Focus()
	case core.ModeSessions:
		m.sessionInput.Focus()
	}
}

func (m *Model) openThemePicker() {
	if len(m.themes) == 0 {
		return
	}

	m.showThemePicker = true
	m.showHelp = false
	m.themePickerPrevIdx = m.activeThemeIdx
	m.themeInput.SetValue("")
	m.filteredThemes = m.themes
	idx := indexOfThemeByName(m.filteredThemes, m.themes[m.activeThemeIdx].Name)
	if idx < 0 {
		idx = 0
	}
	m.themeList.Select(idx)
	m.blurInputs()
	m.themeInput.Focus()
}

func (m *Model) closeThemePicker(apply bool) {
	if !apply {
		m.applyThemeByIndex(m.themePickerPrevIdx)
	}
	m.showThemePicker = false
	m.themeInput.SetValue("")
	m.filteredThemes = m.themes
	m.themeList.Select(0)
	m.themeInput.Blur()
	m.restoreInputFocus()
}

func (m *Model) syncProgressTheme(theme Theme) {
	m.progress.FullColor = string(theme.Accent)
	m.progress.EmptyColor = string(theme.Muted)
}

func (m *Model) applyThemeByIndex(themeIdx int) {
	if themeIdx < 0 || themeIdx >= len(m.themes) {
		return
	}
	theme := m.themes[themeIdx]
	m.activeThemeIdx = themeIdx
	m.styles = NewStyles(theme)
	m.syncProgressTheme(theme)
	m.applyHelpStyles()
	m.applyListStyles()
}

func (m *Model) applyHelpStyles() {
	m.help.Styles.ShortKey = m.styles.Key
	m.help.Styles.ShortDesc = m.styles.Help
	m.help.Styles.FullKey = m.styles.Key
	m.help.Styles.FullDesc = m.styles.Help
}

func (m *Model) applyThemeSelection() {
	idx := m.themeList.Index()
	if len(m.filteredThemes) == 0 || idx < 0 || idx >= len(m.filteredThemes) {
		return
	}
	selected := m.filteredThemes[idx]
	themeIdx := indexOfThemeByName(m.themes, selected.Name)
	if themeIdx < 0 {
		return
	}
	m.applyThemeByIndex(themeIdx)
}

func (m *Model) previewSelectedTheme() {
	idx := m.themeList.Index()
	if len(m.filteredThemes) == 0 || idx < 0 || idx >= len(m.filteredThemes) {
		return
	}
	selected := m.filteredThemes[idx]
	themeIdx := indexOfThemeByName(m.themes, selected.Name)
	if themeIdx < 0 {
		return
	}
	m.applyThemeByIndex(themeIdx)
}

func (m *Model) refreshThemeFilter() {
	m.filteredThemes = FilterThemes(m.themes, m.themeInput.Value())
	m.syncThemeList()
	if len(m.filteredThemes) == 0 {
		m.themeList.Select(0)
		return
	}
	activeName := m.themes[m.activeThemeIdx].Name
	idx := indexOfThemeByName(m.filteredThemes, activeName)
	if idx < 0 {
		m.themeList.Select(0)
	} else {
		m.themeList.Select(idx)
	}
	m.previewSelectedTheme()
}

func indexOfThemeByName(themes []Theme, name string) int {
	for i, theme := range themes {
		if theme.Name == name {
			return i
		}
	}
	return -1
}

func (m Model) displayPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}

	home := strings.TrimSpace(m.homeDir)
	if home == "" {
		return path
	}

	home = filepath.Clean(home)
	cleaned := filepath.Clean(path)
	if cleaned == home {
		return "~"
	}

	prefix := home + string(filepath.Separator)
	if strings.HasPrefix(cleaned, prefix) {
		return "~/" + strings.TrimPrefix(cleaned, prefix)
	}

	return path
}

func (m Model) Init() tea.Cmd {
	coreModel, effects := core.Init(m.core)
	m.core = coreModel
	m.syncLists()
	cmd := m.runEffects(effects)
	return tea.Batch(m.spinner.Tick, cmd)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.applyListStyles()
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		if m.showHelp {
			switch {
			case key.Matches(msg, m.keymap.Back, m.keymap.Toggle):
				m.showHelp = false
				switch m.core.Mode {
				case core.ModeBrowsing:
					m.input.Focus()
				case core.ModeWorktree:
					m.worktreeInput.Focus()
				case core.ModeTool:
					m.toolInput.Focus()
				case core.ModeSessions:
					m.sessionInput.Focus()
				}
				return m, nil
			case key.Matches(msg, m.keymap.Quit):
				return m, tea.Quit
			}
			return m, nil
		}

		if m.showThemePicker {
			switch {
			case key.Matches(msg, m.keymap.Back):
				m.closeThemePicker(false)
				return m, nil
			case key.Matches(msg, m.keymap.Select):
				m.applyThemeSelection()
				m.closeThemePicker(true)
				return m, nil
			case key.Matches(msg, m.keymap.Up, m.keymap.Down):
				var cmd tea.Cmd
				m.themeList, cmd = m.themeList.Update(msg)
				m.previewSelectedTheme()
				return m, cmd
			case key.Matches(msg, m.keymap.Quit):
				return m, tea.Quit
			}

			var cmd tea.Cmd
			m.themeInput, cmd = m.themeInput.Update(msg)
			m.refreshThemeFilter()
			return m, cmd
		}

		if key.Matches(msg, m.keymap.Toggle) && m.core.Mode != core.ModeLoading {
			m.showHelp = true
			m.blurInputs()
			return m, nil
		}

		if key.Matches(msg, m.keymap.Theme) && m.core.Mode != core.ModeLoading {
			m.openThemePicker()
			return m, nil
		}

		prevMode := m.core.Mode

		action, mapped := m.keymap.actionForCore(msg)
		coreModel, effects, handled := core.UpdateKey(m.core, action)
		if !mapped {
			handled = false
		}
		m.core = coreModel
		m.syncLists()
		if spec := extractSessionSpec(effects); spec != nil {
			m.SelectedSpec = spec
		}

		if prevMode == core.ModeBrowsing && m.core.Mode == core.ModeWorktree {
			m.input.Blur()
			m.worktreeInput.Focus()
		}
		if prevMode == core.ModeBrowsing && m.core.Mode == core.ModeProjectDeleteConfirm {
			m.input.Blur()
		}
		if prevMode == core.ModeWorktree && m.core.Mode == core.ModeBrowsing {
			m.worktreeInput.SetValue("")
			m.worktreeInput.Blur()
			m.input.Focus()
		}
		if prevMode == core.ModeWorktree && m.core.Mode == core.ModeTool {
			m.worktreeInput.Blur()
			m.toolInput.SetValue("")
			m.toolInput.Focus()
		}
		if prevMode == core.ModeTool && m.core.Mode == core.ModeToolStarting {
			m.toolInput.Blur()
			m.beginToolStartingProgress(time.Now())
		}
		if prevMode == core.ModeToolStarting && m.core.Mode != core.ModeToolStarting {
			if m.core.Mode == core.ModeTool {
				m.toolInput.Focus()
			}
			m.toolStartingAt = time.Time{}
			m.toolStartingDuration = 0
		}
		if prevMode == core.ModeWorktree && m.core.Mode == core.ModeWorktreeDeleteConfirm {
			m.worktreeInput.Blur()
		}
		if prevMode != core.ModeSessions && m.core.Mode == core.ModeSessions {
			m.blurInputs()
			m.sessionInput.SetValue("")
			m.sessionInput.Focus()
		}
		if prevMode == core.ModeSessions && m.core.Mode != core.ModeSessions {
			m.sessionInput.SetValue("")
			m.sessionInput.Blur()
			if m.core.Mode == core.ModeBrowsing {
				m.input.Focus()
			}
			if m.core.Mode == core.ModeWorktree {
				m.worktreeInput.Focus()
			}
			if m.core.Mode == core.ModeTool {
				m.toolInput.Focus()
			}
		}
		if prevMode == core.ModeTool && m.core.Mode == core.ModeWorktree {
			m.toolInput.SetValue("")
			m.toolInput.Blur()
			m.worktreeInput.Focus()
		}
		if prevMode == core.ModeWorktreeDeleteConfirm && m.core.Mode == core.ModeWorktree {
			m.worktreeInput.Focus()
		}
		if prevMode == core.ModeProjectDeleteConfirm && m.core.Mode == core.ModeBrowsing {
			m.input.Focus()
		}

		if !handled {
			var cmd tea.Cmd
			switch m.core.Mode {
			case core.ModeBrowsing:
				m.input, cmd = m.input.Update(msg)
				cmds = append(cmds, cmd)

				coreModel, effects := core.Update(m.core, core.MsgQueryChanged{Query: m.input.Value()})
				m.core = coreModel
				cmds = append(cmds, m.runEffects(effects))
			case core.ModeWorktree:
				m.worktreeInput, cmd = m.worktreeInput.Update(msg)
				cmds = append(cmds, cmd)

				coreModel, effects := core.Update(m.core, core.MsgWorktreeQueryChanged{Query: m.worktreeInput.Value()})
				m.core = coreModel
				cmds = append(cmds, m.runEffects(effects))
			case core.ModeTool:
				m.toolInput, cmd = m.toolInput.Update(msg)
				cmds = append(cmds, cmd)

				coreModel, effects := core.Update(m.core, core.MsgToolQueryChanged{Query: m.toolInput.Value()})
				m.core = coreModel
				cmds = append(cmds, m.runEffects(effects))
			case core.ModeSessions:
				m.sessionInput, cmd = m.sessionInput.Update(msg)
				cmds = append(cmds, cmd)

				coreModel, effects := core.Update(m.core, core.MsgSessionQueryChanged{Query: m.sessionInput.Value()})
				m.core = coreModel
				cmds = append(cmds, m.runEffects(effects))
			}
		}

		m.syncLists()
		cmds = append(cmds, m.runEffects(effects))
		return m, tea.Batch(cmds...)

	case scanCompletedMsg:
		coreModel, effects := core.Update(m.core, core.MsgScanCompleted{
			Dirs: msg.dirs,
			Err:  msg.err,
		})
		m.core = coreModel
		m.syncLists()
		cmd := m.runEffects(effects)
		return m, cmd

	case projectCreatedMsg:
		coreModel, effects := core.Update(m.core, core.MsgProjectCreated{
			ProjectPath: msg.projectPath,
			Err:         msg.err,
		})
		m.core = coreModel
		m.syncLists()
		if spec := extractSessionSpec(effects); spec != nil {
			m.SelectedSpec = spec
		}
		m.worktreeInput.Focus()
		cmd := m.runEffects(effects)
		return m, cmd

	case projectDeletedMsg:
		coreModel, effects := core.Update(m.core, core.MsgProjectDeleted{
			ProjectPath: msg.projectPath,
			Err:         msg.err,
		})
		m.core = coreModel
		m.syncLists()
		if spec := extractSessionSpec(effects); spec != nil {
			m.SelectedSpec = spec
		}
		if m.core.Mode == core.ModeBrowsing {
			m.input.Focus()
		}
		cmd := m.runEffects(effects)
		return m, cmd

	case worktreesLoadedMsg:
		coreModel, effects := core.Update(m.core, core.MsgWorktreesLoaded{
			Worktrees: msg.worktrees,
			Warning:   msg.warning,
			Err:       msg.err,
		})
		m.core = coreModel
		m.syncLists()
		if spec := extractSessionSpec(effects); spec != nil {
			m.SelectedSpec = spec
		}
		m.worktreeInput.Focus()
		cmd := m.runEffects(effects)
		return m, cmd

	case worktreeCreatedMsg:
		coreModel, effects := core.Update(m.core, core.MsgWorktreeCreated{
			Path: msg.path,
			Err:  msg.err,
		})
		m.core = coreModel
		m.syncLists()
		if spec := extractSessionSpec(effects); spec != nil {
			m.SelectedSpec = spec
		}
		if m.core.Mode == core.ModeTool {
			m.worktreeInput.Blur()
			m.toolInput.SetValue("")
			m.toolInput.Focus()
		}
		cmd := m.runEffects(effects)
		return m, cmd

	case worktreeDeletedMsg:
		coreModel, effects := core.Update(m.core, core.MsgWorktreeDeleted{
			Path: msg.path,
			Err:  msg.err,
		})
		m.core = coreModel
		m.syncLists()
		if spec := extractSessionSpec(effects); spec != nil {
			m.SelectedSpec = spec
		}
		if m.core.Mode == core.ModeWorktree {
			m.worktreeInput.Focus()
		}
		cmd := m.runEffects(effects)
		return m, cmd

	case core.MsgToolPrewarmFailed:
		coreModel, effects := core.Update(m.core, msg)
		m.core = coreModel
		m.syncLists()
		cmd := m.runEffects(effects)
		return m, cmd

	case core.MsgToolPrewarmStarted:
		coreModel, effects := core.Update(m.core, msg)
		m.core = coreModel
		m.syncLists()
		cmd := m.runEffects(effects)
		return m, cmd

	case core.MsgToolPrewarmExisting:
		coreModel, effects := core.Update(m.core, msg)
		m.core = coreModel
		if spec := extractSessionSpec(effects); spec != nil {
			m.SelectedSpec = spec
		}
		cmd := m.runEffects(effects)
		return m, cmd

	case core.MsgToolDelayElapsed:
		coreModel, effects := core.Update(m.core, msg)
		m.core = coreModel
		if spec := extractSessionSpec(effects); spec != nil {
			m.SelectedSpec = spec
		}
		cmd := m.runEffects(effects)
		return m, cmd

	case sessionsLoadedMsg:
		coreModel, effects := core.Update(m.core, core.MsgSessionsLoaded{
			Sessions: msg.sessions,
			Err:      msg.err,
		})
		m.core = coreModel
		cmd := m.runEffects(effects)
		return m, cmd

	case sessionAttachedMsg:
		if msg.err == nil {
			m.SelectedSpec = nil
			m.SelectedSessionName = msg.session.Name
			return m, tea.Quit
		}
		m.core.Mode = core.ModeError
		m.core.Err = msg.err
		return m, nil
	}

	return m, nil
}

type scanCompletedMsg struct {
	dirs []core.DirEntry
	err  error
}

type projectCreatedMsg struct {
	projectPath string
	err         error
}

type projectDeletedMsg struct {
	projectPath string
	err         error
}

type worktreesLoadedMsg struct {
	worktrees []core.Worktree
	warning   string
	err       error
}

type worktreeCreatedMsg struct {
	path string
	err  error
}

type worktreeDeletedMsg struct {
	path string
	err  error
}

type sessionsLoadedMsg struct {
	sessions []core.SessionInfo
	err      error
}

type sessionAttachedMsg struct {
	session core.SessionInfo
	err     error
}

var errNoSessions = errors.New("session manager not configured")

func (m Model) runEffects(effects []core.Effect) tea.Cmd {
	var cmds []tea.Cmd

	for _, eff := range effects {
		switch e := eff.(type) {
		case core.EffScanDirs:
			cmds = append(cmds, m.scanDirsCmd(e.Roots))
		case core.EffCreateProject:
			cmds = append(cmds, m.createProjectCmd(e.Path))
		case core.EffDeleteProject:
			cmds = append(cmds, m.deleteProjectCmd(e.ProjectPath))
		case core.EffLoadWorktrees:
			cmds = append(cmds, m.loadWorktreesCmd(e.ProjectPath))
		case core.EffCreateWorktree:
			cmds = append(cmds, m.createWorktreeCmd(e.ProjectPath, e.BranchName))
		case core.EffDeleteWorktree:
			cmds = append(cmds, m.deleteWorktreeCmd(e.ProjectPath, e.WorktreePath))
		case core.EffPrewarmAllTools:
			cmds = append(cmds, m.prewarmAllToolsCmd(e.DirPath, e.Tools))
		case core.EffCheckToolReady:
			cmds = append(cmds, m.checkToolReadyCmd(e.Spec))
		case core.EffListSessions:
			cmds = append(cmds, m.listSessionsCmd())
		case core.EffAttachSession:
			cmds = append(cmds, m.attachSessionCmd(e.Session))
		case core.EffOpenSession:
			cmds = append(cmds, tea.Quit)
		case core.EffQuit:
			cmds = append(cmds, tea.Quit)
		}
	}

	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

func extractSessionSpec(effects []core.Effect) *core.SessionSpec {
	for _, eff := range effects {
		if e, ok := eff.(core.EffOpenSession); ok {
			spec := e.Spec
			return &spec
		}
	}
	return nil
}

func (m Model) scanDirsCmd(roots []string) tea.Cmd {
	return func() tea.Msg {
		dirs, err := m.fs.ScanDirs(roots, m.maxDepth)
		return scanCompletedMsg{dirs: dirs, err: err}
	}
}

func (m Model) createProjectCmd(path string) tea.Cmd {
	return func() tea.Msg {
		projectPath, err := m.fs.CreateProject(path)
		return projectCreatedMsg{projectPath: projectPath, err: err}
	}
}

func (m Model) deleteProjectCmd(projectPath string) tea.Cmd {
	return func() tea.Msg {
		if m.sessions != nil {
			paths, err := m.fs.ListWorktreePaths(projectPath)
			if err != nil {
				return projectDeletedMsg{projectPath: projectPath, err: err}
			}
			for _, path := range paths {
				for _, tool := range m.core.Tools {
					spec := core.SessionSpec{DirPath: path, Tool: tool}
					if err := m.sessions.KillSession(spec); err != nil {
						return projectDeletedMsg{projectPath: projectPath, err: err}
					}
				}
			}
		}

		err := m.fs.DeleteProject(projectPath)
		return projectDeletedMsg{projectPath: projectPath, err: err}
	}
}

func (m Model) loadWorktreesCmd(projectPath string) tea.Cmd {
	return func() tea.Msg {
		listing, err := m.fs.ListWorktrees(projectPath)
		return worktreesLoadedMsg{worktrees: listing.Worktrees, warning: listing.Warning, err: err}
	}
}

func (m Model) createWorktreeCmd(projectPath, branchName string) tea.Cmd {
	return func() tea.Msg {
		path, err := m.fs.CreateWorktree(projectPath, branchName)
		return worktreeCreatedMsg{path: path, err: err}
	}
}

func (m Model) deleteWorktreeCmd(projectPath, worktreePath string) tea.Cmd {
	return func() tea.Msg {
		if m.sessions != nil {
			for _, tool := range m.core.Tools {
				spec := core.SessionSpec{DirPath: worktreePath, Tool: tool}
				if err := m.sessions.KillSession(spec); err != nil {
					return worktreeDeletedMsg{path: worktreePath, err: err}
				}
			}
		}

		err := m.fs.DeleteWorktree(projectPath, worktreePath)
		return worktreeDeletedMsg{path: worktreePath, err: err}
	}
}

func (m Model) prewarmAllToolsCmd(dirPath string, tools []string) tea.Cmd {
	if m.sessions == nil {
		return nil
	}
	cmds := make([]tea.Cmd, 0, len(tools))
	for _, tool := range tools {
		tool := tool
		spec := core.SessionSpec{DirPath: dirPath, Tool: tool, Detach: true}
		cmds = append(cmds, func() tea.Msg {
			created, err := m.sessions.PrewarmSession(spec)
			if err != nil {
				return core.MsgToolPrewarmFailed{Tool: tool, Err: err}
			}
			if created {
				return core.MsgToolPrewarmStarted{Tool: tool, StartedAt: time.Now()}
			}
			return core.MsgToolPrewarmExisting{Tool: tool}
		})
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

func (m Model) listSessionsCmd() tea.Cmd {
	return func() tea.Msg {
		if m.sessions == nil {
			return sessionsLoadedMsg{}
		}
		sessions, err := m.sessions.ListSessions()
		return sessionsLoadedMsg{sessions: sessions, err: err}
	}
}

func (m Model) attachSessionCmd(session core.SessionInfo) tea.Cmd {
	return func() tea.Msg {
		if m.sessions == nil {
			return sessionAttachedMsg{err: errNoSessions}
		}
		return sessionAttachedMsg{session: session}
	}
}

const toolStartingMinDuration = 200 * time.Millisecond

func (m *Model) beginToolStartingProgress(now time.Time) {
	toolReadyDelay := core.ToolWarmupDelay("")
	if m.core.PendingSpec != nil {
		toolReadyDelay = core.ToolWarmupDelay(m.core.PendingSpec.Tool)
	}
	duration := toolReadyDelay
	if m.core.PendingSpec != nil && m.core.ToolWarmStart != nil {
		if start, ok := m.core.ToolWarmStart[m.core.PendingSpec.Tool]; ok {
			if start.IsZero() {
				duration = toolStartingMinDuration
			} else {
				remaining := toolReadyDelay - now.Sub(start)
				if remaining > 0 {
					duration = remaining
				} else {
					duration = toolStartingMinDuration
				}
			}
		}
	}
	if duration < toolStartingMinDuration {
		duration = toolStartingMinDuration
	}
	m.toolStartingAt = now
	m.toolStartingDuration = duration
}

func (m Model) checkToolReadyCmd(spec core.SessionSpec) tea.Cmd {
	toolReadyDelay := core.ToolWarmupDelay(spec.Tool)
	if m.core.ToolWarmStart != nil {
		if start, ok := m.core.ToolWarmStart[spec.Tool]; ok {
			if start.IsZero() {
				return func() tea.Msg {
					return core.MsgToolDelayElapsed{Tool: spec.Tool}
				}
			}
			remaining := toolReadyDelay - time.Since(start)
			if remaining <= 0 {
				return func() tea.Msg {
					return core.MsgToolDelayElapsed{Tool: spec.Tool}
				}
			}
			return tea.Tick(remaining, func(time.Time) tea.Msg {
				return core.MsgToolDelayElapsed{Tool: spec.Tool}
			})
		}
	}
	return tea.Tick(toolReadyDelay, func(time.Time) tea.Msg {
		return core.MsgToolDelayElapsed{Tool: spec.Tool}
	})
}
