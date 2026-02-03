package adapters

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ariguillegp/solo/internal/core"
	"github.com/google/uuid"
)

type OSFilesystem struct{}

func NewOSFilesystem() *OSFilesystem {
	return &OSFilesystem{}
}

const soloWorktreesDir = "~/.solo/worktrees"

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

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

func (f *OSFilesystem) CreateProject(path string) (string, error) {
	projectPath := expandPath(path)
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return "", err
	}

	initCmd := exec.Command("git", "init", "-b", "main", projectPath)
	if err := initCmd.Run(); err != nil {
		fallbackCmd := exec.Command("git", "init", projectPath)
		if err := fallbackCmd.Run(); err != nil {
			return "", err
		}
		renameCmd := exec.Command("git", "-C", projectPath, "symbolic-ref", "HEAD", "refs/heads/main")
		if err := renameCmd.Run(); err != nil {
			return "", err
		}
	}

	return projectPath, nil
}

func (f *OSFilesystem) ListWorktrees(projectPath string) (core.WorktreeListing, error) {
	projectPath = expandPath(projectPath)
	if !hasGitMarker(projectPath) {
		return core.WorktreeListing{Warning: "Project has no repository. Create a project first."}, nil
	}

	pruneCmd := gitCommand(projectPath, "worktree", "prune")
	_ = pruneCmd.Run()

	cmd := gitCommand(projectPath, "worktree", "list", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return core.WorktreeListing{}, err
	}

	projectName := filepath.Base(projectPath)
	soloDir := expandPath(soloWorktreesDir)
	prefix := projectName + "--"

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
		wtClean := filepath.Clean(wt.Path)
		isRoot := wtClean == filepath.Clean(projectPath)
		isUnderSolo := strings.HasPrefix(wtClean, soloDir+string(filepath.Separator)) &&
			strings.HasPrefix(filepath.Base(wtClean), prefix)

		if !isRoot && !isUnderSolo {
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
	if !hasGitMarker(projectPath) {
		return "", fmt.Errorf("Project has no repository. Create a project first.")
	}

	cleanBranch := strings.TrimSpace(branchName)
	if cleanBranch == "" {
		return "", fmt.Errorf("branch name cannot be empty")
	}

	projectName := filepath.Base(projectPath)
	sanitizedBranch := core.SanitizeWorktreeName(cleanBranch)
	shortUUID := uuid.New().String()[:8]
	worktreeDir := fmt.Sprintf("%s--%s--%s", projectName, sanitizedBranch, shortUUID)

	soloDir := expandPath(soloWorktreesDir)
	if err := os.MkdirAll(soloDir, 0755); err != nil {
		return "", err
	}
	worktreePath := filepath.Join(soloDir, worktreeDir)

	hasCommit, _ := repoHasCommit(projectPath)
	if !hasCommit {
		cmd := gitCommand(projectPath, "worktree", "add", "--orphan", "-b", cleanBranch, worktreePath)
		if output, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("%s: %s", err, string(output))
		}
		return worktreePath, nil
	}

	cmd := gitCommand(projectPath, "worktree", "add", "-b", cleanBranch, worktreePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		cmd = gitCommand(projectPath, "worktree", "add", worktreePath, cleanBranch)
		output2, err2 := cmd.CombinedOutput()
		if err2 != nil {
			return "", fmt.Errorf("%s: %s", err2, string(output2))
		}
		return worktreePath, nil
	}
	_ = output

	return worktreePath, nil
}

func (f *OSFilesystem) PruneWorktrees(projectPath string) error {
	projectPath = expandPath(projectPath)
	if !hasGitMarker(projectPath) {
		return nil
	}

	cmd := gitCommand(projectPath, "worktree", "prune")
	_, err := cmd.CombinedOutput()
	return err
}

func (f *OSFilesystem) DeleteWorktree(projectPath, worktreePath string) error {
	projectPath = expandPath(projectPath)
	if !hasGitMarker(projectPath) {
		return fmt.Errorf("Project has no repository. Create a project first.")
	}

	cleanPath := expandPath(worktreePath)
	if cleanPath == "" {
		return fmt.Errorf("worktree path cannot be empty")
	}
	if filepath.Clean(cleanPath) == filepath.Clean(projectPath) {
		return fmt.Errorf("cannot delete the project root worktree")
	}

	soloDir := expandPath(soloWorktreesDir)
	if !strings.HasPrefix(cleanPath, soloDir+string(filepath.Separator)) {
		return fmt.Errorf("can only delete worktrees under %s", soloWorktreesDir)
	}

	if !isRegisteredWorktree(projectPath, cleanPath) {
		return fmt.Errorf("worktree is not registered in git worktree list")
	}

	cmd := gitCommand(projectPath, "worktree", "remove", "--force", cleanPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, string(output))
	}
	return nil
}

func isRegisteredWorktree(projectPath, worktreePath string) bool {
	cmd := gitCommand(projectPath, "worktree", "list", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	cleanTarget := filepath.Clean(worktreePath)
	for _, line := range bytes.Split(output, []byte("\n")) {
		lineStr := string(line)
		if strings.HasPrefix(lineStr, "worktree ") {
			wtPath := filepath.Clean(strings.TrimPrefix(lineStr, "worktree "))
			if wtPath == cleanTarget {
				return true
			}
		}
	}
	return false
}

func gitCommand(repoPath string, args ...string) *exec.Cmd {
	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	return cmd
}

func repoHasCommit(repoPath string) (bool, error) {
	cmd := gitCommand(repoPath, "rev-parse", "--verify", "HEAD")
	if output, err := cmd.CombinedOutput(); err != nil {
		_ = output
		return false, nil
	}
	return true, nil
}
