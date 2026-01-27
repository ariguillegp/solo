package core

type Mode int

const (
	ModeLoading Mode = iota
	ModeBrowsing
	ModeCreateDir
	ModeError
)

type Model struct {
	Mode        Mode
	Query       string
	Dirs        []DirEntry
	Filtered    []DirEntry
	SelectedIdx int
	RootPaths   []string
	Backend     SessionBackend
	Err         error
}

func NewModel(roots []string) Model {
	return Model{
		Mode:      ModeLoading,
		RootPaths: roots,
		Backend:   BackendNative,
	}
}

func (m Model) SelectedDir() (DirEntry, bool) {
	if len(m.Filtered) == 0 || m.SelectedIdx >= len(m.Filtered) {
		return DirEntry{}, false
	}
	return m.Filtered[m.SelectedIdx], true
}
