package ports

import "github.com/ariguillegp/solo/internal/core"

type Filesystem interface {
	ScanDirs(roots []string, maxDepth int) ([]core.DirEntry, error)
	CreateProject(path string) (string, error)
	ListWorktrees(projectPath string) (core.WorktreeListing, error)
	CreateWorktree(projectPath, branchName string) (string, error)
}
