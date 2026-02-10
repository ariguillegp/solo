package adapters

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateWorktreeFailsWithoutGitRepo(t *testing.T) {
	projectPath := t.TempDir()

	fs := &OSFilesystem{}
	_, err := fs.CreateWorktree(projectPath, "feature/test")
	if err == nil {
		t.Fatal("expected error when no .git exists")
	}
	if err.Error() != "Project has no repository. Create a project first." {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateWorktreeCreatesUnderSoloDir(t *testing.T) {
	projectPath := t.TempDir()
	initRepo(t, projectPath)

	fs := &OSFilesystem{}
	worktreePath, err := fs.CreateWorktree(projectPath, "feature/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	home, _ := os.UserHomeDir()
	soloDir := filepath.Join(home, ".solo", "worktrees")
	if !strings.HasPrefix(worktreePath, soloDir) {
		t.Fatalf("expected worktree under %s, got %s", soloDir, worktreePath)
	}

	projectPrefix := projectWorktreePrefix(projectPath)
	wtName := filepath.Base(worktreePath)
	expectedName := projectPrefix + "--feature-test"
	if wtName != expectedName {
		t.Fatalf("expected worktree name %q, got %q", expectedName, wtName)
	}

	if _, err := os.Stat(worktreePath); err != nil {
		t.Fatalf("expected worktree path to exist: %v", err)
	}

	t.Cleanup(func() {
		_ = os.RemoveAll(worktreePath)
	})
}

func TestCreateWorktreeRejectsDuplicateBranch(t *testing.T) {
	projectPath := t.TempDir()
	initRepo(t, projectPath)

	fs := &OSFilesystem{}
	worktreePath, err := fs.CreateWorktree(projectPath, "feature/test")
	if err != nil {
		t.Fatalf("unexpected error creating worktree: %v", err)
	}
	defer func() {
		_ = fs.DeleteWorktree(projectPath, worktreePath)
	}()

	_, err = fs.CreateWorktree(projectPath, "feature test")
	if err == nil {
		t.Fatal("expected error when worktree already exists for branch")
	}
	if !strings.Contains(err.Error(), "worktree already exists") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateWorktreeUsesUniquePrefixPerProject(t *testing.T) {
	root := t.TempDir()
	projectA := filepath.Join(root, "client", "api")
	projectB := filepath.Join(root, "personal", "api")
	initRepo(t, projectA)
	initRepo(t, projectB)

	fs := &OSFilesystem{}
	wtA, err := fs.CreateWorktree(projectA, "feature-a")
	if err != nil {
		t.Fatalf("unexpected error creating worktree for project A: %v", err)
	}
	defer func() {
		_ = fs.DeleteWorktree(projectA, wtA)
	}()

	wtB, err := fs.CreateWorktree(projectB, "feature-a")
	if err != nil {
		t.Fatalf("unexpected error creating worktree for project B: %v", err)
	}
	defer func() {
		_ = fs.DeleteWorktree(projectB, wtB)
	}()

	if filepath.Base(wtA) == filepath.Base(wtB) {
		t.Fatalf("expected distinct worktree directories, got %q", filepath.Base(wtA))
	}
}

func TestListWorktreesWarnsWhenNoGitRepo(t *testing.T) {
	projectPath := t.TempDir()

	fs := &OSFilesystem{}
	listing, err := fs.ListWorktrees(projectPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if listing.Warning != "Project has no repository. Create a project first." {
		t.Fatalf("unexpected warning: %q", listing.Warning)
	}
	if len(listing.Worktrees) != 0 {
		t.Fatalf("expected no worktrees, got %d", len(listing.Worktrees))
	}
}

func TestListWorktreesIncludesRootAndSoloWorktrees(t *testing.T) {
	projectPath := t.TempDir()
	initRepo(t, projectPath)

	fs := &OSFilesystem{}

	listing, err := fs.ListWorktrees(projectPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if listing.Warning != "" {
		t.Fatalf("unexpected warning: %q", listing.Warning)
	}
	if len(listing.Worktrees) != 1 {
		t.Fatalf("expected 1 worktree (root), got %d", len(listing.Worktrees))
	}

	worktreePath, err := fs.CreateWorktree(projectPath, "feature/list")
	if err != nil {
		t.Fatalf("unexpected error creating worktree: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(worktreePath)
	})

	listing2, err := fs.ListWorktrees(projectPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(listing2.Worktrees) != 2 {
		t.Fatalf("expected 2 worktrees (root + created), got %d", len(listing2.Worktrees))
	}
}

func TestDeleteWorktreeRemovesPath(t *testing.T) {
	projectPath := t.TempDir()
	initRepo(t, projectPath)

	fs := &OSFilesystem{}
	worktreePath, err := fs.CreateWorktree(projectPath, "feature/delete")
	if err != nil {
		t.Fatalf("unexpected error creating worktree: %v", err)
	}

	if err := fs.DeleteWorktree(projectPath, worktreePath); err != nil {
		t.Fatalf("unexpected error deleting worktree: %v", err)
	}
	if _, err := os.Stat(worktreePath); !os.IsNotExist(err) {
		t.Fatalf("expected worktree path to be removed")
	}
}

func TestDeleteWorktreeRejectsProjectRoot(t *testing.T) {
	projectPath := t.TempDir()
	initRepo(t, projectPath)

	fs := &OSFilesystem{}
	err := fs.DeleteWorktree(projectPath, projectPath)
	if err == nil {
		t.Fatal("expected error when deleting project root")
	}
	if !strings.Contains(err.Error(), "cannot delete the project root") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteWorktreeRejectsPathsOutsideSoloDir(t *testing.T) {
	projectPath := t.TempDir()
	initRepo(t, projectPath)

	outsidePath := filepath.Join(t.TempDir(), "outside-worktree")

	fs := &OSFilesystem{}
	err := fs.DeleteWorktree(projectPath, outsidePath)
	if err == nil {
		t.Fatal("expected error when deleting outside solo dir")
	}
	if !strings.Contains(err.Error(), "can only delete worktrees under") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateProjectCreatesGitRepo(t *testing.T) {
	projectPath := filepath.Join(t.TempDir(), "solo-project")

	fs := &OSFilesystem{}
	createdPath, err := fs.CreateProject(projectPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if createdPath != projectPath {
		t.Fatalf("expected project path %q, got %q", projectPath, createdPath)
	}

	gitPath := filepath.Join(projectPath, ".git")
	if _, err := os.Stat(gitPath); err != nil {
		t.Fatalf("expected .git in project path: %v", err)
	}
}

func TestDeleteProjectRemovesRootAndWorktrees(t *testing.T) {
	projectPath := t.TempDir()
	initRepo(t, projectPath)

	fs := &OSFilesystem{}
	worktreePath, err := fs.CreateWorktree(projectPath, "feature/delete-project")
	if err != nil {
		t.Fatalf("unexpected error creating worktree: %v", err)
	}

	if err := fs.DeleteProject(projectPath); err != nil {
		t.Fatalf("unexpected error deleting project: %v", err)
	}
	if _, err := os.Stat(projectPath); !os.IsNotExist(err) {
		t.Fatalf("expected project path to be removed")
	}
	if _, err := os.Stat(worktreePath); !os.IsNotExist(err) {
		t.Fatalf("expected worktree path to be removed")
	}
}

func TestScanDirsFindsGitProjects(t *testing.T) {
	root := t.TempDir()
	projectPath := filepath.Join(root, "repo")
	initRepo(t, projectPath)

	_ = os.MkdirAll(filepath.Join(root, "plain-dir"), 0755)

	fs := &OSFilesystem{}
	entries, err := fs.ScanDirs([]string{root}, 2)
	if err != nil {
		t.Fatalf("unexpected scan error: %v", err)
	}

	var found bool
	for _, entry := range entries {
		if entry.Path == projectPath {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected to find project %q in scan results", projectPath)
	}
}

func TestScanDirsStopsAtProject(t *testing.T) {
	root := t.TempDir()
	projectPath := filepath.Join(root, "repo")
	initRepo(t, projectPath)

	nestedPath := filepath.Join(projectPath, "nested")
	if err := os.MkdirAll(filepath.Join(nestedPath, ".git"), 0755); err != nil {
		t.Fatalf("unexpected error creating nested git marker: %v", err)
	}

	fs := &OSFilesystem{}
	entries, err := fs.ScanDirs([]string{root}, 4)
	if err != nil {
		t.Fatalf("unexpected scan error: %v", err)
	}

	var foundRoot bool
	var foundNested bool
	for _, entry := range entries {
		switch entry.Path {
		case projectPath:
			foundRoot = true
		case nestedPath:
			foundNested = true
		}
	}
	if !foundRoot {
		t.Fatalf("expected to find project %q in scan results", projectPath)
	}
	if foundNested {
		t.Fatalf("did not expect nested project %q in scan results", nestedPath)
	}
}

func TestListWorktreesFiltersNonSoloWorktrees(t *testing.T) {
	projectPath := t.TempDir()
	initRepo(t, projectPath)

	outsidePath := filepath.Join(t.TempDir(), "outside-worktree")
	cmd := exec.Command("git", "-C", projectPath, "worktree", "add", outsidePath, "-b", "feature/outside")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git worktree add failed: %v: %s", err, string(output))
	}
	defer func() {
		cleanupCmd := exec.Command("git", "-C", projectPath, "worktree", "remove", "--force", outsidePath)
		_ = cleanupCmd.Run()
		_ = os.RemoveAll(outsidePath)
	}()

	fs := &OSFilesystem{}
	listing, err := fs.ListWorktrees(projectPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, wt := range listing.Worktrees {
		if filepath.Clean(wt.Path) == filepath.Clean(outsidePath) {
			t.Fatalf("did not expect outside worktree %q in listing", outsidePath)
		}
	}
}

func TestDeleteWorktreeRejectsUnregisteredWorktree(t *testing.T) {
	projectPath := t.TempDir()
	initRepo(t, projectPath)

	home, _ := os.UserHomeDir()
	soloDir := filepath.Join(home, ".solo", "worktrees")
	projectPrefix := projectWorktreePrefix(projectPath)
	worktreePath := filepath.Join(soloDir, projectPrefix+"--fake-branch")

	if err := os.MkdirAll(worktreePath, 0755); err != nil {
		t.Fatalf("unexpected error creating unregistered worktree: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(worktreePath)
	})

	fs := &OSFilesystem{}
	err := fs.DeleteWorktree(projectPath, worktreePath)
	if err == nil {
		t.Fatal("expected error when deleting unregistered worktree")
	}
	if !strings.Contains(err.Error(), "not registered") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func initRepo(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("git", "init", "-b", "main", dir)
	if output, err := cmd.CombinedOutput(); err != nil {
		cmd = exec.Command("git", "init", dir)
		if output2, err2 := cmd.CombinedOutput(); err2 != nil {
			t.Fatalf("git init failed: %v: %s", err2, string(output2))
		}
		cmd = exec.Command("git", "-C", dir, "symbolic-ref", "HEAD", "refs/heads/main")
		if output2, err2 := cmd.CombinedOutput(); err2 != nil {
			t.Fatalf("git branch -M failed: %v: %s", err2, string(output2))
		}
	} else {
		_ = output
	}

	cmd = exec.Command("git", "-C", dir, "commit", "--allow-empty", "-m", "initial commit")
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=Test",
		"GIT_AUTHOR_EMAIL=test@test.com",
		"GIT_COMMITTER_NAME=Test",
		"GIT_COMMITTER_EMAIL=test@test.com",
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit failed: %v: %s", err, string(output))
	}
}
