package ui

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ariguillegp/solo/internal/core"
	"github.com/ariguillegp/solo/internal/ports"
)

type Model struct {
	core        core.Model
	input       textinput.Model
	spinner     spinner.Model
	fs          ports.Filesystem
	sessions    ports.SessionManager
	maxDepth    int
	width       int
	height      int
	SelectedDir string
}

func New(roots []string, fs ports.Filesystem, sessions ports.SessionManager) Model {
	ti := textinput.New()
	ti.Focus()

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	return Model{
		core:     core.NewModel(roots),
		input:    ti,
		spinner:  sp,
		fs:       fs,
		sessions: sessions,
		maxDepth: 1,
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
				if dir := extractSelectedDir(effects); dir != "" {
					m.SelectedDir = dir
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

		if m.core.Mode == core.ModeCreateDir {
			switch key {
			case "enter", "esc":
				coreModel, effects := core.Update(m.core, core.MsgKeyPress{Key: key})
				m.core = coreModel
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

	case scanCompletedMsg:
		coreModel, effects := core.Update(m.core, core.MsgScanCompleted{
			Dirs: msg.dirs,
			Err:  msg.err,
		})
		m.core = coreModel
		cmd := m.runEffects(effects)
		return m, cmd

	case createDirCompletedMsg:
		coreModel, effects := core.Update(m.core, core.MsgCreateDirCompleted{
			Path: msg.path,
			Err:  msg.err,
		})
		m.core = coreModel
		if dir := extractSelectedDir(effects); dir != "" {
			m.SelectedDir = dir
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

type createDirCompletedMsg struct {
	path string
	err  error
}

func (m Model) runEffects(effects []core.Effect) tea.Cmd {
	var cmds []tea.Cmd

	for _, eff := range effects {
		switch e := eff.(type) {
		case core.EffScanDirs:
			cmds = append(cmds, m.scanDirsCmd(e.Roots))
		case core.EffMkdirAll:
			cmds = append(cmds, m.mkdirCmd(e.Path))
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

func extractSelectedDir(effects []core.Effect) string {
	for _, eff := range effects {
		if e, ok := eff.(core.EffOpenSession); ok {
			return e.Spec.DirPath
		}
	}
	return ""
}

func (m Model) scanDirsCmd(roots []string) tea.Cmd {
	return func() tea.Msg {
		dirs, err := m.fs.ScanDirs(roots, m.maxDepth)
		return scanCompletedMsg{dirs: dirs, err: err}
	}
}

func (m Model) mkdirCmd(path string) tea.Cmd {
	return func() tea.Msg {
		expandedPath, err := m.fs.MkdirAll(path)
		return createDirCompletedMsg{path: expandedPath, err: err}
	}
}
