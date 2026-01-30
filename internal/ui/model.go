package ui

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ariguillegp/solo/internal/core"
	"github.com/ariguillegp/solo/internal/ports"
)

type Model struct {
	core          core.Model
	input         textinput.Model
	worktreeInput textinput.Model
	toolInput     textinput.Model
	spinner       spinner.Model
	fs            ports.Filesystem
	maxDepth      int
	width         int
	height        int
	SelectedSpec  *core.SessionSpec
	ignoreToolHit bool
}

func New(roots []string, fs ports.Filesystem) Model {
	ti := textinput.New()
	ti.Focus()

	wti := textinput.New()
	tti := textinput.New()

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	return Model{
		core:          core.NewModel(roots),
		input:         ti,
		worktreeInput: wti,
		toolInput:     tti,
		spinner:       sp,
		fs:            fs,
		maxDepth:      2,
	}
}

func (m Model) Init() tea.Cmd {
	coreModel, effects := core.Init(m.core)
	m.core = coreModel
	cmd := m.runEffects(effects)
	return tea.Batch(m.spinner.Tick, cmd)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		key := msg.String()

		if m.core.Mode == core.ModeBrowsing {
			switch key {
			case "up", "down", "ctrl+k", "ctrl+j", "enter", "ctrl+n", "esc", "ctrl+c":
				coreModel, effects := core.Update(m.core, core.MsgKeyPress{Key: key})
				m.core = coreModel
				if spec := extractSessionSpec(effects); spec != nil {
					m.SelectedSpec = spec
				}
				if m.core.Mode == core.ModeWorktree {
					m.input.Blur()
					m.worktreeInput.Focus()
				}
				cmd := m.runEffects(effects)
				return m, cmd
			default:
				var cmd tea.Cmd
				m.input, cmd = m.input.Update(msg)
				cmds = append(cmds, cmd)

				coreModel, effects := core.Update(m.core, core.MsgQueryChanged{Query: m.input.Value()})
				m.core = coreModel
				cmds = append(cmds, m.runEffects(effects))
				return m, tea.Batch(cmds...)
			}
		}

		if m.core.Mode == core.ModeWorktree {
			switch key {
			case "up", "down", "ctrl+k", "ctrl+j", "enter", "ctrl+n", "esc", "ctrl+c":
				prevMode := m.core.Mode
				coreModel, effects := core.Update(m.core, core.MsgKeyPress{Key: key})
				m.core = coreModel
				if spec := extractSessionSpec(effects); spec != nil {
					m.SelectedSpec = spec
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
					m.ignoreToolHit = true
				}
				cmd := m.runEffects(effects)
				return m, cmd
			default:
				var cmd tea.Cmd
				m.worktreeInput, cmd = m.worktreeInput.Update(msg)
				cmds = append(cmds, cmd)

				coreModel, effects := core.Update(m.core, core.MsgWorktreeQueryChanged{Query: m.worktreeInput.Value()})
				m.core = coreModel
				cmds = append(cmds, m.runEffects(effects))
				return m, tea.Batch(cmds...)
			}
		}

		if m.core.Mode == core.ModeTool {
			switch key {
			case "up", "down", "ctrl+k", "ctrl+j", "enter", "esc", "ctrl+c":
				if m.ignoreToolHit {
					m.ignoreToolHit = false
					if key == "enter" {
						return m, nil
					}
				}
				prevMode := m.core.Mode
				coreModel, effects := core.Update(m.core, core.MsgKeyPress{Key: key})
				m.core = coreModel
				if spec := extractSessionSpec(effects); spec != nil {
					m.SelectedSpec = spec
				}
				if prevMode == core.ModeTool && m.core.Mode == core.ModeWorktree {
					m.toolInput.SetValue("")
					m.toolInput.Blur()
					m.worktreeInput.Focus()
					m.ignoreToolHit = false
				}
				cmd := m.runEffects(effects)
				return m, cmd
			default:
				var cmd tea.Cmd
				m.toolInput, cmd = m.toolInput.Update(msg)
				cmds = append(cmds, cmd)

				coreModel, effects := core.Update(m.core, core.MsgToolQueryChanged{Query: m.toolInput.Value()})
				m.core = coreModel
				cmds = append(cmds, m.runEffects(effects))
				return m, tea.Batch(cmds...)
			}
		}

	case scanCompletedMsg:
		coreModel, effects := core.Update(m.core, core.MsgScanCompleted{
			Dirs: msg.dirs,
			Err:  msg.err,
		})
		m.core = coreModel
		cmd := m.runEffects(effects)
		return m, cmd

	case projectCreatedMsg:
		coreModel, effects := core.Update(m.core, core.MsgProjectCreated{
			ProjectPath: msg.projectPath,
			Err:         msg.err,
		})
		m.core = coreModel
		if spec := extractSessionSpec(effects); spec != nil {
			m.SelectedSpec = spec
		}
		m.worktreeInput.Focus()
		cmd := m.runEffects(effects)
		return m, cmd

	case worktreesLoadedMsg:
		coreModel, effects := core.Update(m.core, core.MsgWorktreesLoaded{
			Worktrees: msg.worktrees,
			Warning:   msg.warning,
			Err:       msg.err,
		})
		m.core = coreModel
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

type worktreesLoadedMsg struct {
	worktrees []core.Worktree
	warning   string
	err       error
}

type worktreeCreatedMsg struct {
	path string
	err  error
}

func (m Model) runEffects(effects []core.Effect) tea.Cmd {
	var cmds []tea.Cmd

	for _, eff := range effects {
		switch e := eff.(type) {
		case core.EffScanDirs:
			cmds = append(cmds, m.scanDirsCmd(e.Roots))
		case core.EffCreateProject:
			cmds = append(cmds, m.createProjectCmd(e.Path))
		case core.EffLoadWorktrees:
			cmds = append(cmds, m.loadWorktreesCmd(e.ProjectPath))
		case core.EffCreateWorktree:
			cmds = append(cmds, m.createWorktreeCmd(e.ProjectPath, e.BranchName))
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
