package core

type Effect interface {
	isEffect()
}

type EffScanDirs struct {
	Roots []string
}

func (EffScanDirs) isEffect() {}

type EffCreateProject struct {
	Path string
}

func (EffCreateProject) isEffect() {}

type EffDeleteProject struct {
	ProjectPath string
}

func (EffDeleteProject) isEffect() {}

type EffOpenSession struct {
	Spec SessionSpec
}

func (EffOpenSession) isEffect() {}

type EffQuit struct{}

func (EffQuit) isEffect() {}

type EffLoadWorktrees struct {
	ProjectPath string
}

func (EffLoadWorktrees) isEffect() {}

type EffCreateWorktree struct {
	ProjectPath string
	BranchName  string
}

func (EffCreateWorktree) isEffect() {}

type EffDeleteWorktree struct {
	ProjectPath  string
	WorktreePath string
}

func (EffDeleteWorktree) isEffect() {}

type EffPrewarmAllTools struct {
	DirPath string
	Tools   []string
}

func (EffPrewarmAllTools) isEffect() {}

type EffCheckToolReady struct {
	Spec SessionSpec
}

func (EffCheckToolReady) isEffect() {}

type EffListSessions struct{}

func (EffListSessions) isEffect() {}

type EffAttachSession struct {
	Session SessionInfo
}

func (EffAttachSession) isEffect() {}
