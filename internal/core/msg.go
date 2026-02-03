package core

import "time"

type Msg interface {
	isMsg()
}

type MsgScanCompleted struct {
	Dirs []DirEntry
	Err  error
}

func (MsgScanCompleted) isMsg() {}

type MsgProjectCreated struct {
	ProjectPath string
	Err         error
}

func (MsgProjectCreated) isMsg() {}

type MsgProjectDeleted struct {
	ProjectPath string
	Err         error
}

func (MsgProjectDeleted) isMsg() {}

type MsgKeyPress struct {
	Key string
}

func (MsgKeyPress) isMsg() {}

type MsgQueryChanged struct {
	Query string
}

func (MsgQueryChanged) isMsg() {}

type MsgWorktreesLoaded struct {
	Worktrees []Worktree
	Warning   string
	Err       error
}

func (MsgWorktreesLoaded) isMsg() {}

type MsgWorktreeCreated struct {
	Path string
	Err  error
}

func (MsgWorktreeCreated) isMsg() {}

type MsgWorktreeDeleted struct {
	Path string
	Err  error
}

func (MsgWorktreeDeleted) isMsg() {}

type MsgWorktreeQueryChanged struct {
	Query string
}

func (MsgWorktreeQueryChanged) isMsg() {}

type MsgToolQueryChanged struct {
	Query string
}

func (MsgToolQueryChanged) isMsg() {}

type MsgToolPrewarmFailed struct {
	Tool string
	Err  error
}

func (MsgToolPrewarmFailed) isMsg() {}

type MsgToolPrewarmStarted struct {
	Tool      string
	StartedAt time.Time
}

func (MsgToolPrewarmStarted) isMsg() {}

type MsgToolPrewarmExisting struct {
	Tool string
}

func (MsgToolPrewarmExisting) isMsg() {}

type MsgToolDelayElapsed struct {
	Tool string
}

func (MsgToolDelayElapsed) isMsg() {}
