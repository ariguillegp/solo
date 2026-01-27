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

	case MsgCreateDirCompleted:
		if msg.Err != nil {
			m.Mode = ModeError
			m.Err = msg.Err
			return m, nil
		}
		spec := SessionSpec{
			Backend:   m.Backend,
			DirPath:   msg.Path,
			CreateDir: false,
		}
		return m, []Effect{EffOpenSession{Spec: spec}}
	}

	return m, nil
}

func handleKey(m Model, key string) (Model, []Effect) {
	switch m.Mode {
	case ModeBrowsing:
		return handleBrowsingKey(m, key)
	case ModeCreateDir:
		return handleCreateDirKey(m, key)
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
			spec := SessionSpec{
				Backend: m.Backend,
				DirPath: dir.Path,
			}
			return m, []Effect{EffOpenSession{Spec: spec}}
		}
		if m.Query != "" && len(m.RootPaths) > 0 {
			path := m.RootPaths[0] + "/" + m.Query
			return m, []Effect{EffMkdirAll{Path: path}}
		}
	case "ctrl+n":
		if m.Query != "" && len(m.RootPaths) > 0 {
			path := m.RootPaths[0] + "/" + m.Query
			return m, []Effect{EffMkdirAll{Path: path}}
		}
	case "esc", "ctrl+c":
		return m, []Effect{EffQuit{}}
	}
	return m, nil
}

func handleCreateDirKey(m Model, key string) (Model, []Effect) {
	switch key {
	case "enter":
		if m.Query != "" {
			return m, []Effect{EffMkdirAll{Path: m.Query}}
		}
	case "esc":
		m.Mode = ModeBrowsing
	}
	return m, nil
}

func Init(m Model) (Model, []Effect) {
	return m, []Effect{EffScanDirs{Roots: m.RootPaths}}
}
