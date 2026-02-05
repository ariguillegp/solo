package adapters

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ariguillegp/solo/internal/core"
)

func TestKillSessionIgnoresNoServerRunning(t *testing.T) {
	tmpDir := t.TempDir()
	tmuxPath := filepath.Join(tmpDir, "tmux")

	script := "#!/bin/sh\n" +
		"echo \"no server running on /tmp/tmux-0/default\" 1>&2\n" +
		"exit 1\n"

	if err := os.WriteFile(tmuxPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to write tmux stub: %v", err)
	}

	pathEnv := os.Getenv("PATH")
	pathSep := string(os.PathListSeparator)
	t.Setenv("PATH", tmpDir+pathSep+pathEnv)

	session := &TmuxSession{}
	spec := core.SessionSpec{DirPath: "/tmp/project", Tool: "amp"}
	if err := session.KillSession(spec); err != nil {
		t.Fatalf("expected no error when tmux server is missing: %v", err)
	}
}

func TestKillSessionIgnoresMissingSession(t *testing.T) {
	tmpDir := t.TempDir()
	tmuxPath := filepath.Join(tmpDir, "tmux")

	script := "#!/bin/sh\n" +
		"echo \"can't find session\" 1>&2\n" +
		"exit 1\n"

	if err := os.WriteFile(tmuxPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to write tmux stub: %v", err)
	}

	pathEnv := os.Getenv("PATH")
	pathSep := string(os.PathListSeparator)
	t.Setenv("PATH", tmpDir+pathSep+pathEnv)

	session := &TmuxSession{}
	spec := core.SessionSpec{DirPath: "/tmp/project", Tool: "amp"}
	if err := session.KillSession(spec); err != nil {
		t.Fatalf("expected no error when tmux session is missing: %v", err)
	}
}
