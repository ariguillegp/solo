package adapters

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
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

		if hasGitMarker(fullPath) {
			continue
		}

		if hasGitChild(fullPath) {
			*dirs = append(*dirs, core.DirEntry{
				Path:   fullPath,
				Name:   name,
				Exists: true,
			})
			continue
		}

		if err := scanDir(fullPath, depth+1, maxDepth, seen, dirs); err != nil {
			return err
		}
	}

	return nil
}

func hasGitMarker(path string) bool {
	_, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil
}

func hasGitChild(path string) bool {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, ".") || ignoreDirs[name] {
			continue
		}
		childPath := filepath.Join(path, name)
		if hasGitMarker(childPath) {
			return true
		}
	}
	return false
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

func (f *OSFilesystem) CreateProject(path string) (string, error) {
	projectPath := expandPath(path)
	mainPath := filepath.Join(projectPath, "main")
	if err := os.MkdirAll(mainPath, 0755); err != nil {
		return "", err
	}

	initCmd := exec.Command("git", "init", "-b", "main")
	initCmd.Dir = mainPath
	if err := initCmd.Run(); err != nil {
		fallbackCmd := exec.Command("git", "init")
		fallbackCmd.Dir = mainPath
		if err := fallbackCmd.Run(); err != nil {
			return "", err
		}
		renameCmd := exec.Command("git", "branch", "-M", "main")
		renameCmd.Dir = mainPath
		if err := renameCmd.Run(); err != nil {
			return "", err
		}
	}

	commitCmd := exec.Command("git", "commit", "--allow-empty", "-m", "Initial commit")
	commitCmd.Dir = mainPath
	if output, err := commitCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("%s: %s", err, string(output))
	}

	return projectPath, nil
}

func (f *OSFilesystem) ListWorktrees(projectPath string) (core.WorktreeListing, error) {
	projectPath = expandPath(projectPath)
	primaryPath, warning, err := findPrimaryRepo(projectPath)
	if err != nil {
		return core.WorktreeListing{}, err
	}
	if warning != "" {
		return core.WorktreeListing{Warning: warning}, nil
	}

	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = primaryPath
	output, err := cmd.Output()
	if err != nil {
		return core.WorktreeListing{}, err
	}

	var worktrees []core.Worktree
	var current core.Worktree

	for _, line := range bytes.Split(output, []byte("\n")) {
		lineStr := string(line)
		switch {
		case strings.HasPrefix(lineStr, "worktree "):
			if current.Path != "" {
				worktrees = append(worktrees, current)
			}
			current = core.Worktree{
				Path: strings.TrimPrefix(lineStr, "worktree "),
			}
		case strings.HasPrefix(lineStr, "branch refs/heads/"):
			current.Branch = strings.TrimPrefix(lineStr, "branch refs/heads/")
		case lineStr == "detached":
			current.Branch = "(detached)"
		}
	}

	if current.Path != "" {
		worktrees = append(worktrees, current)
	}

	var filtered []core.Worktree
	for _, wt := range worktrees {
		if !isWithinProject(projectPath, wt.Path) {
			continue
		}
		wt.Name = filepath.Base(wt.Path)
		if wt.Branch == "" {
			wt.Branch = wt.Name
		}
		filtered = append(filtered, wt)
	}

	return core.WorktreeListing{Worktrees: filtered}, nil
}

func (f *OSFilesystem) CreateWorktree(projectPath, branchName string) (string, error) {
	projectPath = expandPath(projectPath)
	primaryPath, warning, err := findPrimaryRepo(projectPath)
	if err != nil {
		return "", err
	}
	if warning != "" {
		return "", fmt.Errorf("%s", warning)
	}

	cleanBranch := strings.TrimSpace(branchName)
	if cleanBranch == "" {
		return "", fmt.Errorf("branch name cannot be empty")
	}

	worktreeDir := core.SanitizeWorktreeName(cleanBranch)
	worktreePath := filepath.Join(projectPath, worktreeDir)

	cmd := exec.Command("git", "worktree", "add", "-b", cleanBranch, worktreePath)
	cmd.Dir = primaryPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		cmd = exec.Command("git", "worktree", "add", worktreePath, cleanBranch)
		cmd.Dir = primaryPath
		output2, err2 := cmd.CombinedOutput()
		if err2 != nil {
			return "", fmt.Errorf("%s: %s", err2, string(output2))
		}
		return worktreePath, nil
	}
	_ = output

	return worktreePath, nil
}

func findPrimaryRepo(projectPath string) (string, string, error) {
	entries, err := os.ReadDir(projectPath)
	if err != nil {
		return "", "", err
	}

	var primaryRepos []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, ".") || ignoreDirs[name] {
			continue
		}
		childPath := filepath.Join(projectPath, name)
		if isPrimaryRepo(childPath) {
			primaryRepos = append(primaryRepos, childPath)
		}
	}

	if len(primaryRepos) == 0 {
		return "", "Project has no primary repo. Create a 'main' worktree first.", nil
	}
	if len(primaryRepos) > 1 {
		return "", "Project has multiple primary repos. Keep only one worktree with a .git directory.", nil
	}

	return primaryRepos[0], "", nil
}

func isPrimaryRepo(path string) bool {
	info, err := os.Stat(filepath.Join(path, ".git"))
	if err != nil {
		return false
	}
	return info.IsDir()
}

func isWithinProject(projectPath, worktreePath string) bool {
	rel, err := filepath.Rel(projectPath, worktreePath)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	return !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != ".."
}
