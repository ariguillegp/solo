package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ariguillegp/rivet/internal/core"
)

type stubFilesystem struct {
	createProjectPath      string
	createProjectErr       error
	createProjectCalls     []string
	listWorktreePaths      []string
	listWorktreePathsErr   error
	listWorktreePathsCalls []string
	listing                core.WorktreeListing
	listWorktreesErr       error
	listWorktreesCalls     []string
	createWorktreePath     string
	createWorktreeErr      error
	createWorktreeCalls    []createWorktreeCall
}

type createWorktreeCall struct {
	projectPath string
	branchName  string
}

func (s *stubFilesystem) ScanDirs(_ []string, _ int) ([]core.DirEntry, error) {
	return nil, nil
}

func (s *stubFilesystem) CreateProject(path string) (string, error) {
	s.createProjectCalls = append(s.createProjectCalls, path)
	if s.createProjectErr != nil {
		return "", s.createProjectErr
	}
	if s.createProjectPath != "" {
		return s.createProjectPath, nil
	}
	return path, nil
}

func (s *stubFilesystem) DeleteProject(string) error {
	return nil
}

func (s *stubFilesystem) ListWorktreePaths(projectPath string) ([]string, error) {
	s.listWorktreePathsCalls = append(s.listWorktreePathsCalls, projectPath)
	if s.listWorktreePathsErr != nil {
		return nil, s.listWorktreePathsErr
	}
	return append([]string(nil), s.listWorktreePaths...), nil
}

func (s *stubFilesystem) ListWorktrees(projectPath string) (core.WorktreeListing, error) {
	s.listWorktreesCalls = append(s.listWorktreesCalls, projectPath)
	if s.listWorktreesErr != nil {
		return core.WorktreeListing{}, s.listWorktreesErr
	}
	return s.listing, nil
}

func (s *stubFilesystem) CreateWorktree(projectPath, branchName string) (string, error) {
	s.createWorktreeCalls = append(s.createWorktreeCalls, createWorktreeCall{
		projectPath: projectPath,
		branchName:  branchName,
	})
	if s.createWorktreeErr != nil {
		return "", s.createWorktreeErr
	}
	if s.createWorktreePath != "" {
		return s.createWorktreePath, nil
	}
	return filepath.Join(projectPath, branchName), nil
}

func (s *stubFilesystem) DeleteWorktree(string, string) error {
	return nil
}

func (s *stubFilesystem) PruneWorktrees(string) error {
	return nil
}

func TestResolveSessionSpecRequiresFlags(t *testing.T) {
	fs := &stubFilesystem{}

	_, err := resolveSessionSpec(fs, nil, "", "main", "amp", false, false)
	if err == nil || err.Error() != "--project is required" {
		t.Fatalf("expected missing project error, got %v", err)
	}

	_, err = resolveSessionSpec(fs, nil, "demo", "", "amp", false, false)
	if err == nil || err.Error() != "--worktree is required" {
		t.Fatalf("expected missing worktree error, got %v", err)
	}

	_, err = resolveSessionSpec(fs, nil, "demo", "main", "", false, false)
	if err == nil || err.Error() != "--tool is required" {
		t.Fatalf("expected missing tool error, got %v", err)
	}
}

func TestResolveSessionSpecRejectsUnsupportedTool(t *testing.T) {
	fs := &stubFilesystem{}

	_, err := resolveSessionSpec(fs, nil, "demo", "main", "invalid", false, false)
	if err == nil || !strings.Contains(err.Error(), "unsupported tool") {
		t.Fatalf("expected unsupported tool error, got %v", err)
	}
}

func TestResolveSessionSpecResolvesNamedProjectAndWorktree(t *testing.T) {
	root := t.TempDir()
	projectPath := filepath.Join(root, "demo")
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("failed to create project path: %v", err)
	}

	worktreePath := filepath.Join(root, "worktrees", "feature")
	fs := &stubFilesystem{
		listing: core.WorktreeListing{
			Worktrees: []core.Worktree{
				{Path: worktreePath, Name: "feature", Branch: "feature"},
			},
		},
	}

	spec, err := resolveSessionSpec(fs, []string{root}, "demo", "feature", "amp", false, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.DirPath != worktreePath {
		t.Fatalf("expected worktree path %q, got %q", worktreePath, spec.DirPath)
	}
	if spec.Tool != "amp" {
		t.Fatalf("expected tool amp, got %q", spec.Tool)
	}
	if !spec.Detach {
		t.Fatalf("expected detach to be true")
	}
	if len(fs.createWorktreeCalls) != 0 {
		t.Fatalf("did not expect create worktree call, got %d calls", len(fs.createWorktreeCalls))
	}
}

func TestResolveSessionSpecCreatesProjectAndWorktreeWhenMissing(t *testing.T) {
	root := t.TempDir()
	projectPath := filepath.Join(root, "demo")
	worktreePath := filepath.Join(root, ".rivet", "worktrees", "demo--feature-a")
	fs := &stubFilesystem{
		createProjectPath:  projectPath,
		createWorktreePath: worktreePath,
	}

	spec, err := resolveSessionSpec(fs, []string{root}, "demo", "feature-a", "codex", true, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.DirPath != worktreePath {
		t.Fatalf("expected created worktree path %q, got %q", worktreePath, spec.DirPath)
	}
	if len(fs.createProjectCalls) != 1 || fs.createProjectCalls[0] != projectPath {
		t.Fatalf("unexpected create project calls: %v", fs.createProjectCalls)
	}
	if len(fs.createWorktreeCalls) != 1 {
		t.Fatalf("expected one create worktree call, got %d", len(fs.createWorktreeCalls))
	}
	if fs.createWorktreeCalls[0].projectPath != projectPath || fs.createWorktreeCalls[0].branchName != "feature-a" {
		t.Fatalf("unexpected create worktree call: %+v", fs.createWorktreeCalls[0])
	}
}

func TestResolveProjectPathCreatesMissingPathLikeProject(t *testing.T) {
	root := t.TempDir()
	projectPath := filepath.Join(root, "new-project")
	fs := &stubFilesystem{createProjectPath: projectPath}

	resolved, err := resolveProjectPath(fs, nil, projectPath, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved != projectPath {
		t.Fatalf("expected path %q, got %q", projectPath, resolved)
	}
	if len(fs.createProjectCalls) != 1 || fs.createProjectCalls[0] != projectPath {
		t.Fatalf("unexpected create project calls: %v", fs.createProjectCalls)
	}
}

func TestResolveWorktreePathReturnsListingWarning(t *testing.T) {
	fs := &stubFilesystem{
		listing: core.WorktreeListing{Warning: "Project has no repository. Create a project first."},
	}

	_, err := resolveWorktreePath(fs, "/projects/demo", "feature")
	if err == nil || !strings.Contains(err.Error(), "Project has no repository") {
		t.Fatalf("expected warning error, got %v", err)
	}
}

func TestExpandRootsExpandsHomePrefix(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		t.Skip("home directory not available")
	}

	roots := expandRoots([]string{"~/Projects", "/tmp"})
	if roots[0] != filepath.Join(home, "Projects") {
		t.Fatalf("expected first root to expand to home path, got %q", roots[0])
	}
	if roots[1] != "/tmp" {
		t.Fatalf("expected second root unchanged, got %q", roots[1])
	}
}

func TestLooksLikePathDetectsPathForms(t *testing.T) {
	if !looksLikePath("/tmp/demo") {
		t.Fatalf("expected absolute path to be detected")
	}
	if !looksLikePath("./demo") {
		t.Fatalf("expected relative path to be detected")
	}
	if !looksLikePath("nested/demo") {
		t.Fatalf("expected separator path to be detected")
	}
	if looksLikePath("demo") {
		t.Fatalf("did not expect plain name to be detected as path")
	}
}

func TestExistsReturnsTrueOnlyForDirectories(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(file, []byte("x"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	if !exists(dir) {
		t.Fatalf("expected directory to exist")
	}
	if exists(file) {
		t.Fatalf("expected regular file to be rejected")
	}
	if exists(filepath.Join(dir, "missing")) {
		t.Fatalf("expected missing path to be rejected")
	}
}

func TestResetTerminalInvokesStty(t *testing.T) {
	tmp := t.TempDir()
	logPath := filepath.Join(tmp, "stty.log")
	sttyPath := filepath.Join(tmp, "stty")
	writeExecutable(t, sttyPath, "#!/bin/sh\necho \"$@\" > \""+logPath+"\"\nexit 0\n")

	t.Setenv("PATH", tmp+string(os.PathListSeparator)+os.Getenv("PATH"))

	if err := resetTerminal(); err != nil {
		t.Fatalf("unexpected error resetting terminal: %v", err)
	}

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read stty log: %v", err)
	}
	if strings.TrimSpace(string(content)) != "sane" {
		t.Fatalf("expected stty sane call, got %q", string(content))
	}
}

func TestReturnToPreviousSessionNoopOutsideTmux(t *testing.T) {
	t.Setenv("TMUX", "")
	if err := returnToPreviousSession(); err != nil {
		t.Fatalf("expected no error outside tmux, got %v", err)
	}
}

func TestReturnToPreviousSessionSwitchesClientInsideTmux(t *testing.T) {
	tmp := t.TempDir()
	logPath := filepath.Join(tmp, "tmux.log")
	tmuxPath := filepath.Join(tmp, "tmux")
	writeExecutable(t, tmuxPath, "#!/bin/sh\necho \"$@\" > \""+logPath+"\"\nexit 0\n")

	t.Setenv("PATH", tmp+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("TMUX", "1")

	if err := returnToPreviousSession(); err != nil {
		t.Fatalf("unexpected error switching back to previous session: %v", err)
	}

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read tmux log: %v", err)
	}
	if strings.TrimSpace(string(content)) != "switch-client -l" {
		t.Fatalf("expected switch-client -l call, got %q", string(content))
	}
}

func writeExecutable(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		t.Fatalf("failed to write executable %s: %v", path, err)
	}
}
