package adapters

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindPrimaryRepoWarnings(t *testing.T) {
	t.Run("no primary repo", func(t *testing.T) {
		projectPath := t.TempDir()

		primary, warning, err := findPrimaryRepo(projectPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if primary != "" {
			t.Fatalf("expected empty primary repo path, got %q", primary)
		}
		if warning != "Project has no primary repo. Create a 'main' worktree first." {
			t.Fatalf("unexpected warning: %q", warning)
		}
	})

	t.Run("multiple primary repos", func(t *testing.T) {
		projectPath := t.TempDir()
		primaryDir := filepath.Join(projectPath, "main")
		otherDir := filepath.Join(projectPath, "secondary")

		if err := os.MkdirAll(filepath.Join(primaryDir, ".git"), 0755); err != nil {
			t.Fatalf("failed to create primary repo marker: %v", err)
		}
		if err := os.MkdirAll(filepath.Join(otherDir, ".git"), 0755); err != nil {
			t.Fatalf("failed to create secondary repo marker: %v", err)
		}

		primary, warning, err := findPrimaryRepo(projectPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if primary != "" {
			t.Fatalf("expected empty primary repo path, got %q", primary)
		}
		if warning != "Project has multiple primary repos. Keep only one worktree with a .git directory." {
			t.Fatalf("unexpected warning: %q", warning)
		}
	})
}

func TestCreateWorktreeFailsWithoutPrimaryRepo(t *testing.T) {
	projectPath := t.TempDir()

	fs := &OSFilesystem{}
	_, err := fs.CreateWorktree(projectPath, "feature/test")
	if err == nil {
		t.Fatal("expected error when no primary repo exists")
	}
	if err.Error() != "Project has no primary repo. Create a 'main' worktree first." {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateWorktreeCreatesSanitizedPath(t *testing.T) {
	setGitIdentity(t)

	projectPath := t.TempDir()
	primaryPath := filepath.Join(projectPath, "main")
	if err := os.MkdirAll(primaryPath, 0755); err != nil {
		t.Fatalf("failed to create primary path: %v", err)
	}

	initRepo(t, primaryPath)
	writeFile(t, filepath.Join(primaryPath, "README.md"), []byte("init\n"))
	runGit(t, primaryPath, "add", ".")
	runGit(t, primaryPath, "commit", "-m", "init")

	fs := &OSFilesystem{}
	worktreePath, err := fs.CreateWorktree(projectPath, "feature/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := filepath.Join(projectPath, "feature-test")
	if worktreePath != expected {
		t.Fatalf("expected worktree path %q, got %q", expected, worktreePath)
	}
	if _, err := os.Stat(worktreePath); err != nil {
		t.Fatalf("expected worktree path to exist: %v", err)
	}
}

func TestListWorktreesWarnsWhenNoPrimaryRepo(t *testing.T) {
	projectPath := t.TempDir()

	fs := &OSFilesystem{}
	listing, err := fs.ListWorktrees(projectPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if listing.Warning != "Project has no primary repo. Create a 'main' worktree first." {
		t.Fatalf("unexpected warning: %q", listing.Warning)
	}
	if len(listing.Worktrees) != 0 {
		t.Fatalf("expected no worktrees, got %d", len(listing.Worktrees))
	}
}

func TestListWorktreesWarnsWhenMultiplePrimaryRepos(t *testing.T) {
	projectPath := t.TempDir()
	primaryDir := filepath.Join(projectPath, "main")
	secondaryDir := filepath.Join(projectPath, "secondary")

	if err := os.MkdirAll(filepath.Join(primaryDir, ".git"), 0755); err != nil {
		t.Fatalf("failed to create primary repo marker: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(secondaryDir, ".git"), 0755); err != nil {
		t.Fatalf("failed to create secondary repo marker: %v", err)
	}

	fs := &OSFilesystem{}
	listing, err := fs.ListWorktrees(projectPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if listing.Warning != "Project has multiple primary repos. Keep only one worktree with a .git directory." {
		t.Fatalf("unexpected warning: %q", listing.Warning)
	}
	if len(listing.Worktrees) != 0 {
		t.Fatalf("expected no worktrees, got %d", len(listing.Worktrees))
	}
}

func TestDeleteWorktreeRemovesPath(t *testing.T) {
	setGitIdentity(t)

	projectPath := t.TempDir()
	primaryPath := filepath.Join(projectPath, "main")
	if err := os.MkdirAll(primaryPath, 0755); err != nil {
		t.Fatalf("failed to create primary path: %v", err)
	}

	initRepo(t, primaryPath)
	writeFile(t, filepath.Join(primaryPath, "README.md"), []byte("init\n"))
	runGit(t, primaryPath, "add", ".")
	runGit(t, primaryPath, "commit", "-m", "init")

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

func TestCreateProjectCreatesPrimaryRepo(t *testing.T) {
	setGitIdentity(t)

	projectPath := filepath.Join(t.TempDir(), "solo-project")

	fs := &OSFilesystem{}
	createdPath, err := fs.CreateProject(projectPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if createdPath != projectPath {
		t.Fatalf("expected project path %q, got %q", projectPath, createdPath)
	}

	mainPath := filepath.Join(projectPath, "main")
	if _, err := os.Stat(filepath.Join(mainPath, ".git")); err != nil {
		t.Fatalf("expected git repo in main path: %v", err)
	}

	cmd := exec.Command("git", "log", "-1", "--pretty=%B")
	cmd.Dir = mainPath
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("git log failed: %v", err)
	}
	if strings.TrimSpace(string(output)) != "Initial commit" {
		t.Fatalf("unexpected initial commit message: %q", strings.TrimSpace(string(output)))
	}
}

func setGitIdentity(t *testing.T) {
	t.Helper()
	t.Setenv("GIT_AUTHOR_NAME", "solo")
	t.Setenv("GIT_AUTHOR_EMAIL", "solo@example.com")
	t.Setenv("GIT_COMMITTER_NAME", "solo")
	t.Setenv("GIT_COMMITTER_EMAIL", "solo@example.com")
}

func initRepo(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("git", "init", "-b", "main")
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		cmd = exec.Command("git", "init")
		cmd.Dir = dir
		if output2, err2 := cmd.CombinedOutput(); err2 != nil {
			t.Fatalf("git init failed: %v: %s", err2, string(output2))
		}
		cmd = exec.Command("git", "branch", "-M", "main")
		cmd.Dir = dir
		if output2, err2 := cmd.CombinedOutput(); err2 != nil {
			t.Fatalf("git branch -M failed: %v: %s", err2, string(output2))
		}
	} else {
		_ = output
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v: %s", args, err, string(output))
	}
}

func writeFile(t *testing.T, path string, contents []byte) {
	t.Helper()
	if err := os.WriteFile(path, contents, 0644); err != nil {
		t.Fatalf("write file failed: %v", err)
	}
}
