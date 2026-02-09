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

func (f *OSFilesystem) DeleteProject(projectPath string) error {
	projectPath = expandPath(projectPath)
	if !hasGitMarker(projectPath) {
		return fmt.Errorf("Project has no repository. Create a project first.")
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
			return fmt.Errorf("%s: %s", err, string(output))
		}
	}

	if err := os.RemoveAll(projectClean); err != nil {
		return err
	}
	return nil
}

func (f *OSFilesystem) ListWorktreePaths(projectPath string) ([]string, error) {
	projectPath = expandPath(projectPath)
	if !hasGitMarker(projectPath) {
		return nil, fmt.Errorf("Project has no repository. Create a project first.")
	}
	if err := f.migrateLegacyWorktrees(projectPath); err != nil {
		return nil, err
	}
	return listWorktreePathsRaw(projectPath)
}

func (f *OSFilesystem) ListWorktrees(projectPath string) (core.WorktreeListing, error) {
	projectPath = expandPath(projectPath)
	if !hasGitMarker(projectPath) {
		return core.WorktreeListing{Warning: "Project has no repository. Create a project first."}, nil
	}
	if err := f.migrateLegacyWorktrees(projectPath); err != nil {
		return core.WorktreeListing{}, err
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
	soloDirClean := filepath.Clean(soloDir)
	projectDir := filepath.Join(soloDirClean, projectName)
	projectDirPrefix := projectDir + string(filepath.Separator)

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
		isUnderSoloProject := strings.HasPrefix(wtClean, projectDirPrefix)

		if !isRoot && !isUnderSoloProject {
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
	if err := f.migrateLegacyWorktrees(projectPath); err != nil {
		return "", err
	}

	cleanBranch := strings.TrimSpace(branchName)
	if cleanBranch == "" {
		return "", fmt.Errorf("branch name cannot be empty")
	}
	if !core.IsValidWorktreeName(cleanBranch) {
		return "", core.ErrInvalidWorktreeName
	}

	projectName := filepath.Base(projectPath)
	sanitizedBranch := core.SanitizeWorktreeName(cleanBranch)

	soloDir := expandPath(soloWorktreesDir)
	projectDir := filepath.Join(soloDir, projectName)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return "", err
	}
	_ = f.PruneWorktrees(projectPath)
	worktreePath := filepath.Join(projectDir, filepath.FromSlash(sanitizedBranch))
	if worktreePathTaken(projectPath, worktreePath) {
		return "", fmt.Errorf("%w for branch %s", core.ErrWorktreeExists, cleanBranch)
	}

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

func worktreePathTaken(projectPath, worktreePath string) bool {
	if _, err := os.Stat(worktreePath); err == nil {
		return true
	} else if !os.IsNotExist(err) {
		return true
	}

	return isRegisteredWorktree(projectPath, worktreePath)
}

func isShortUUIDPart(part string) bool {
	if len(part) != 8 {
		return false
	}
	for i := 0; i < len(part); i++ {
		ch := part[i]
		if (ch < '0' || ch > '9') && (ch < 'a' || ch > 'f') {
			return false
		}
	}
	return true
}

func listWorktreePathsRaw(projectPath string) ([]string, error) {
	cmd := gitCommand(projectPath, "worktree", "list", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var paths []string
	for _, line := range bytes.Split(output, []byte("\n")) {
		lineStr := string(line)
		if strings.HasPrefix(lineStr, "worktree ") {
			paths = append(paths, strings.TrimPrefix(lineStr, "worktree "))
		}
	}
	return paths, nil
}

func (f *OSFilesystem) migrateLegacyWorktrees(projectPath string) error {
	projectPath = expandPath(projectPath)
	if !hasGitMarker(projectPath) {
		return fmt.Errorf("Project has no repository. Create a project first.")
	}

	projectName := filepath.Base(projectPath)
	if projectName == "" || projectName == "." || projectName == string(filepath.Separator) {
		return fmt.Errorf("cannot determine project name for migration")
	}

	paths, err := listWorktreePathsRaw(projectPath)
	if err != nil {
		return err
	}

	soloDir := expandPath(soloWorktreesDir)
	legacyPrefix := projectName + "--"
	legacyRoot := filepath.Clean(soloDir)
	projectDir := filepath.Join(soloDir, projectName)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return err
	}

	for _, wtPath := range paths {
		wtClean := filepath.Clean(wtPath)
		if filepath.Dir(wtClean) != legacyRoot {
			continue
		}
		base := filepath.Base(wtClean)
		if !strings.HasPrefix(base, legacyPrefix) {
			continue
		}
		branch, ok := legacyBranchFromName(base, legacyPrefix)
		if !ok {
			return fmt.Errorf("cannot migrate legacy worktree %s", base)
		}
		if strings.TrimSpace(branch) == "" {
			return fmt.Errorf("cannot migrate legacy worktree %s", base)
		}
		target := filepath.Join(projectDir, branch)
		if filepath.Clean(target) == wtClean {
			continue
		}
		if worktreePathTaken(projectPath, target) {
			return fmt.Errorf("cannot migrate %s: target already exists", base)
		}

		moveCmd := gitCommand(projectPath, "worktree", "move", wtClean, target)
		output, err := moveCmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("%s: %s", err, string(output))
		}
	}

	return nil
}

func legacyBranchFromName(name, prefix string) (string, bool) {
	if !strings.HasPrefix(name, prefix) {
		return "", false
	}
	trimmed := strings.TrimPrefix(name, prefix)
	trimmed = strings.TrimSpace(trimmed)
	if trimmed == "" {
		return "", false
	}
	parts := strings.Split(trimmed, "--")
	if len(parts) == 0 {
		return "", false
	}
	if len(parts) == 1 {
		return strings.TrimSpace(parts[0]), true
	}
	last := strings.TrimSpace(parts[len(parts)-1])
	if isShortUUIDPart(last) {
		branch := strings.TrimSpace(strings.Join(parts[:len(parts)-1], "--"))
		if branch == "" {
			return "", false
		}
		return branch, true
	}
	return strings.TrimSpace(strings.Join(parts, "--")), true
}

func (f *OSFilesystem) PruneWorktrees(projectPath string) error {
	projectPath = expandPath(projectPath)
	if !hasGitMarker(projectPath) {
		return nil
	}
	if err := f.migrateLegacyWorktrees(projectPath); err != nil {
		return err
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
	if err := f.migrateLegacyWorktrees(projectPath); err != nil {
		return err
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
