package adapters

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ariguillegp/solo/internal/core"
)

type OSFilesystem struct{}

func NewOSFilesystem() *OSFilesystem {
	return &OSFilesystem{}
}

var ignoreDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
	".cache":       true,
	"__pycache__":  true,
	".venv":        true,
	"target":       true,
}

func (f *OSFilesystem) ScanDirs(roots []string, maxDepth int) ([]core.DirEntry, error) {
	var dirs []core.DirEntry
	seen := make(map[string]bool)

	for _, root := range roots {
		root = expandPath(root)
		err := scanDir(root, 0, maxDepth, seen, &dirs)
		if err != nil {
			continue
		}
	}

	return dirs, nil
}

func scanDir(path string, depth, maxDepth int, seen map[string]bool, dirs *[]core.DirEntry) error {
	if depth > maxDepth {
		return nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasPrefix(name, ".") || ignoreDirs[name] {
			continue
		}

		fullPath := filepath.Join(path, name)
		if seen[fullPath] {
			continue
		}
		seen[fullPath] = true

		*dirs = append(*dirs, core.DirEntry{
			Path:   fullPath,
			Name:   name,
			Exists: true,
		})

		scanDir(fullPath, depth+1, maxDepth, seen, dirs)
	}

	return nil
}

func (f *OSFilesystem) MkdirAll(path string) (string, error) {
	path = expandPath(path)
	return path, os.MkdirAll(path, 0755)
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}
