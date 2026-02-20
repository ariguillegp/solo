package adapters

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ariguillegp/rivet/internal/core"
)

type OSFilesystem struct{}

func NewOSFilesystem() *OSFilesystem {
	return &OSFilesystem{}
}

const rivetWorktreesDir = "~/.rivet/worktrees"

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
	if err := os.MkdirAll(projectPath, 0o755); err != nil {
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

func (f *OSFilesystem) DeleteProject(projectPath string) error {
	projectPath = expandPath(projectPath)
	if !hasGitMarker(projectPath) {
		return fmt.Errorf("project has no repository; create a project first")
	}

	_ = f.PruneWorktrees(projectPath)

	worktreePaths, err := f.ListWorktreePaths(projectPath)
	if err != nil {
		return err
	}

	projectClean := filepath.Clean(projectPath)
	for _, wtPath := range worktreePaths {
		cleanPath := filepath.Clean(wtPath)
		if cleanPath == filepath.Clean(projectPath) {
			continue
		}
		if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
			continue
		}
		cmd := gitCommand(projectPath, "worktree", "remove", "--force", cleanPath)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("%w: %s", err, string(output))
		}
	}

	return os.RemoveAll(projectClean)
}

func (f *OSFilesystem) ListWorktreePaths(projectPath string) ([]string, error) {
	projectPath = expandPath(projectPath)
	if !hasGitMarker(projectPath) {
		return nil, fmt.Errorf("project has no repository; create a project first")
	}

	cmd := gitCommand(projectPath, "worktree", "list", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var paths []string
	for line := range bytes.SplitSeq(output, []byte("\n")) {
		lineStr := string(line)
		if after, ok := strings.CutPrefix(lineStr, "worktree "); ok {
			paths = append(paths, after)
		}
	}
	return paths, nil
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
	projectID := projectWorktreePrefix(projectPath)
	rivetDir := expandPath(rivetWorktreesDir)
	prefix := projectID + "--"
	legacyPrefix := projectName + "--"

	var worktrees []core.Worktree
	var current core.Worktree

	for line := range bytes.SplitSeq(output, []byte("\n")) {
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
		isUnderRivet := strings.HasPrefix(wtClean, rivetDir+string(filepath.Separator)) &&
			(strings.HasPrefix(filepath.Base(wtClean), prefix) || strings.HasPrefix(filepath.Base(wtClean), legacyPrefix))

		if !isRoot && !isUnderRivet {
			continue
		}

		wt.Name = filepath.Base(wt.Path)
		filtered = append(filtered, wt)
	}

	return core.WorktreeListing{Worktrees: filtered}, nil
}

func (f *OSFilesystem) CreateWorktree(projectPath, branchName string) (string, error) {
	projectPath = expandPath(projectPath)
	if !hasGitMarker(projectPath) {
		return "", fmt.Errorf("project has no repository; create a project first")
	}

	cleanBranch := strings.TrimSpace(branchName)
	if cleanBranch == "" {
		return "", fmt.Errorf("branch name cannot be empty")
	}

	projectID := projectWorktreePrefix(projectPath)
	sanitizedBranch := core.SanitizeWorktreeName(cleanBranch)
	if sanitizedBranch == "" {
		return "", fmt.Errorf("branch name cannot be empty")
	}
	worktreeDir := fmt.Sprintf("%s--%s", projectID, sanitizedBranch)

	listing, err := f.ListWorktrees(projectPath)
	if err != nil {
		return "", err
	}
	for _, wt := range listing.Worktrees {
		if strings.TrimSpace(wt.Branch) == cleanBranch {
			return "", core.WorktreeExistsError{Branch: cleanBranch}
		}
		if sanitizedBranch != "" && core.SanitizeWorktreeName(wt.Branch) == sanitizedBranch {
			return "", core.WorktreeExistsError{Branch: cleanBranch}
		}
	}

	rivetDir := expandPath(rivetWorktreesDir)
	if err := os.MkdirAll(rivetDir, 0o755); err != nil {
		return "", err
	}
	worktreePath := filepath.Join(rivetDir, worktreeDir)

	hasCommit := repoHasCommit(projectPath)
	if !hasCommit {
		cmd := gitCommand(projectPath, "worktree", "add", "--orphan", "-b", cleanBranch, worktreePath)
		if output, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("%w: %s", err, string(output))
		}
		return worktreePath, nil
	}

	cmd := gitCommand(projectPath, "worktree", "add", "-b", cleanBranch, worktreePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if bytes.Contains(output, []byte("already exists")) && bytes.Contains(output, []byte(cleanBranch)) {
			return "", core.WorktreeExistsError{Branch: cleanBranch}
		}
		cmd = gitCommand(projectPath, "worktree", "add", "--orphan", "-b", cleanBranch, worktreePath)
		output2, err2 := cmd.CombinedOutput()
		if err2 != nil {
			if bytes.Contains(output2, []byte("already exists")) && bytes.Contains(output2, []byte(cleanBranch)) {
				return "", core.WorktreeExistsError{Branch: cleanBranch}
			}
			return "", fmt.Errorf("%w: %s", err2, string(output2))
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
		return fmt.Errorf("project has no repository; create a project first")
	}

	cleanPath := expandPath(worktreePath)
	if cleanPath == "" {
		return fmt.Errorf("worktree path cannot be empty")
	}
	if filepath.Clean(cleanPath) == filepath.Clean(projectPath) {
		return core.ErrWorktreeDeleteRoot
	}

	rivetDir := expandPath(rivetWorktreesDir)
	if !strings.HasPrefix(cleanPath, rivetDir+string(filepath.Separator)) {
		return fmt.Errorf("%w: %s", core.ErrWorktreeDeleteOutsideRoot, rivetWorktreesDir)
	}

	if !isRegisteredWorktree(projectPath, cleanPath) {
		return core.ErrWorktreeUnregistered
	}

	cmd := gitCommand(projectPath, "worktree", "remove", "--force", cleanPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
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
	for line := range bytes.SplitSeq(output, []byte("\n")) {
		lineStr := string(line)
		if after, ok := strings.CutPrefix(lineStr, "worktree "); ok {
			wtPath := filepath.Clean(after)
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

func repoHasCommit(repoPath string) bool {
	cmd := gitCommand(repoPath, "rev-parse", "--verify", "HEAD")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func projectWorktreePrefix(projectPath string) string {
	cleanPath := filepath.Clean(projectPath)
	projectName := filepath.Base(cleanPath)
	hasher := sha1.Sum([]byte(cleanPath))
	suffix := hex.EncodeToString(hasher[:])[:6]
	return fmt.Sprintf("%s-%s", projectName, suffix)
}
