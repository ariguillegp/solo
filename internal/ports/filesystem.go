package ports

import "github.com/ariguillegp/solo/internal/core"

type Filesystem interface {
	ScanDirs(roots []string, maxDepth int) ([]core.DirEntry, error)
	MkdirAll(path string) (string, error)
}
