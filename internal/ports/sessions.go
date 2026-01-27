package ports

import "github.com/ariguillegp/solo/internal/core"

type SessionManager interface {
	OpenSession(spec core.SessionSpec) error
}
