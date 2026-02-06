package ports

import "github.com/ariguillegp/solo/internal/core"

type SessionManager interface {
	OpenSession(spec core.SessionSpec) error
	PrewarmSession(spec core.SessionSpec) (bool, error)
	KillSession(spec core.SessionSpec) error
	ListSessions() ([]core.SessionInfo, error)
	AttachSession(name string) error
}
