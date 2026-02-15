package adapters

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ariguillegp/rivet/internal/core"
)

func TestOpenSessionDetachCreatesSessionWithoutAttach(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "tmux.log")
	tmuxPath := filepath.Join(tmpDir, "tmux")
	writeExecutable(t, tmuxPath, `#!/bin/sh
echo "$@" >> "$TMUX_LOG"
if [ "$1" = "has-session" ]; then
  exit 1
fi
if [ "$1" = "new-session" ]; then
  exit 0
fi
if [ "$1" = "attach-session" ] || [ "$1" = "switch-client" ]; then
  echo "unexpected command $1" 1>&2
  exit 1
fi
exit 0
`)

	t.Setenv("TMUX_LOG", logPath)
	t.Setenv("PATH", tmpDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("TMUX", "")

	session := &TmuxSession{}
	spec := core.SessionSpec{DirPath: "/tmp/project", Tool: "amp", Detach: true}
	if err := session.OpenSession(spec); err != nil {
		t.Fatalf("unexpected open-session error: %v", err)
	}

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read tmux log: %v", err)
	}
	log := string(content)
	if !strings.Contains(log, "has-session -t =-tmp-project__amp") {
		t.Fatalf("expected has-session check, got log:\n%s", log)
	}
	if !strings.Contains(log, "new-session -d -s -tmp-project__amp") {
		t.Fatalf("expected new-session call, got log:\n%s", log)
	}
	if strings.Contains(log, "attach-session") || strings.Contains(log, "switch-client") {
		t.Fatalf("did not expect attach or switch in detach mode, got log:\n%s", log)
	}
}

func TestOpenSessionInsideTmuxUsesSwitchClient(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "tmux.log")
	tmuxPath := filepath.Join(tmpDir, "tmux")
	writeExecutable(t, tmuxPath, `#!/bin/sh
echo "$@" >> "$TMUX_LOG"
if [ "$1" = "has-session" ]; then
  exit 0
fi
if [ "$1" = "switch-client" ]; then
  exit 0
fi
echo "unexpected command $1" 1>&2
exit 1
`)

	t.Setenv("TMUX_LOG", logPath)
	t.Setenv("PATH", tmpDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("TMUX", "1")

	session := &TmuxSession{}
	spec := core.SessionSpec{DirPath: "/tmp/project", Tool: "amp"}
	if err := session.OpenSession(spec); err != nil {
		t.Fatalf("unexpected open-session error: %v", err)
	}

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read tmux log: %v", err)
	}
	log := string(content)
	if !strings.Contains(log, "has-session -t =-tmp-project__amp") {
		t.Fatalf("expected has-session check, got log:\n%s", log)
	}
	if !strings.Contains(log, "switch-client -t =-tmp-project__amp") {
		t.Fatalf("expected switch-client call, got log:\n%s", log)
	}
	if strings.Contains(log, "attach-session") {
		t.Fatalf("did not expect attach-session inside tmux, got log:\n%s", log)
	}
}

func TestPrewarmSessionReturnsFalseWhenSessionExists(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "tmux.log")
	tmuxPath := filepath.Join(tmpDir, "tmux")
	writeExecutable(t, tmuxPath, `#!/bin/sh
echo "$@" >> "$TMUX_LOG"
if [ "$1" = "has-session" ]; then
  exit 0
fi
if [ "$1" = "new-session" ]; then
  echo "new-session should not be called" 1>&2
  exit 1
fi
exit 0
`)

	t.Setenv("TMUX_LOG", logPath)
	t.Setenv("PATH", tmpDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	session := &TmuxSession{}
	spec := core.SessionSpec{DirPath: "/tmp/project", Tool: "amp"}
	created, err := session.PrewarmSession(spec)
	if err != nil {
		t.Fatalf("unexpected prewarm error: %v", err)
	}
	if created {
		t.Fatalf("expected created=false for existing session")
	}
}

func TestAttachSessionUsesAttachOutsideTmux(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "tmux.log")
	tmuxPath := filepath.Join(tmpDir, "tmux")
	writeExecutable(t, tmuxPath, `#!/bin/sh
echo "$@" >> "$TMUX_LOG"
if [ "$1" = "attach-session" ]; then
  exit 0
fi
echo "unexpected command $1" 1>&2
exit 1
`)

	t.Setenv("TMUX_LOG", logPath)
	t.Setenv("PATH", tmpDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("TMUX", "")

	session := &TmuxSession{}
	if err := session.AttachSession("demo__amp"); err != nil {
		t.Fatalf("unexpected attach error: %v", err)
	}

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read tmux log: %v", err)
	}
	if !strings.Contains(string(content), "attach-session -t =demo__amp") {
		t.Fatalf("expected attach-session call, got log:\n%s", string(content))
	}
}

func TestAttachSessionUsesSwitchInsideTmux(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "tmux.log")
	tmuxPath := filepath.Join(tmpDir, "tmux")
	writeExecutable(t, tmuxPath, `#!/bin/sh
echo "$@" >> "$TMUX_LOG"
if [ "$1" = "switch-client" ]; then
  exit 0
fi
echo "unexpected command $1" 1>&2
exit 1
`)

	t.Setenv("TMUX_LOG", logPath)
	t.Setenv("PATH", tmpDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("TMUX", "1")

	session := &TmuxSession{}
	if err := session.AttachSession("demo__amp"); err != nil {
		t.Fatalf("unexpected attach error: %v", err)
	}

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read tmux log: %v", err)
	}
	if !strings.Contains(string(content), "switch-client -t =demo__amp") {
		t.Fatalf("expected switch-client call, got log:\n%s", string(content))
	}
}

func TestSwitchClientFallsBackToDefaultTarget(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "tmux.log")
	tmuxPath := filepath.Join(tmpDir, "tmux")
	writeExecutable(t, tmuxPath, `#!/bin/sh
echo "$@" >> "$TMUX_LOG"
if [ "$1" = "display-message" ]; then
  echo "/dev/pts/9"
  exit 0
fi
if [ "$1" = "switch-client" ] && [ "$2" = "-c" ]; then
  exit 1
fi
if [ "$1" = "switch-client" ]; then
  exit 0
fi
echo "unexpected command $1" 1>&2
exit 1
`)

	t.Setenv("TMUX_LOG", logPath)
	t.Setenv("PATH", tmpDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("TMUX_PANE", "%1")

	if err := switchClient("demo__amp"); err != nil {
		t.Fatalf("unexpected switch-client error: %v", err)
	}

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read tmux log: %v", err)
	}
	log := string(content)
	if !strings.Contains(log, "switch-client -c /dev/pts/9 -t =demo__amp") {
		t.Fatalf("expected client-specific switch attempt, got log:\n%s", log)
	}
	if !strings.Contains(log, "switch-client -t =demo__amp") {
		t.Fatalf("expected fallback switch attempt, got log:\n%s", log)
	}
}

func TestToolCommandAndTmuxEnvArgs(t *testing.T) {
	t.Setenv("SHELL", "/bin/bash")
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("COLORTERM", "truecolor")
	t.Setenv("COLORFGBG", "15;0")

	shell, args := toolCommand(core.ToolNone)
	if shell != "/bin/bash" {
		t.Fatalf("expected configured shell, got %q", shell)
	}
	if len(args) != 0 {
		t.Fatalf("expected no command args for none tool, got %v", args)
	}

	shell, args = toolCommand("amp")
	if shell != "/bin/bash" {
		t.Fatalf("expected configured shell, got %q", shell)
	}
	if len(args) == 0 || args[len(args)-1] != "amp" {
		t.Fatalf("expected warmup command args for amp, got %v", args)
	}

	envArgs := tmuxEnvArgs("opencode")
	joined := strings.Join(envArgs, " ")
	if !strings.Contains(joined, "OPENCODE_CONFIG_CONTENT=") {
		t.Fatalf("expected opencode environment to be included, got %v", envArgs)
	}
	if !strings.Contains(joined, "TERM=xterm-256color") {
		t.Fatalf("expected TERM to be included, got %v", envArgs)
	}
	if !strings.Contains(joined, "COLORTERM=truecolor") {
		t.Fatalf("expected COLORTERM to be included, got %v", envArgs)
	}
}

func writeExecutable(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		t.Fatalf("failed to write executable %s: %v", path, err)
	}
}
