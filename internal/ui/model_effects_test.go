package ui

import (
	"errors"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ariguillegp/rivet/internal/core"
)

type fakeFilesystem struct {
	scanDirsEntries         []core.DirEntry
	scanDirsErr             error
	scanDirsCalls           int
	scanDirsRoots           []string
	scanDirsDepth           int
	createProjectPath       string
	createProjectErr        error
	createProjectCalls      []string
	deleteProjectErr        error
	deleteProjectCalls      []string
	listWorktreePathsResult []string
	listWorktreePathsErr    error
	listWorktreePathsCalls  []string
	listWorktreesListing    core.WorktreeListing
	listWorktreesErr        error
	listWorktreesCalls      []string
	createWorktreePath      string
	createWorktreeErr       error
	createWorktreeCalls     []createWorktreeCall
	deleteWorktreeErr       error
	deleteWorktreeCalls     []deleteWorktreeCall
}

func (f *fakeFilesystem) ScanDirs(roots []string, maxDepth int) ([]core.DirEntry, error) {
	f.scanDirsCalls++
	f.scanDirsRoots = append([]string(nil), roots...)
	f.scanDirsDepth = maxDepth
	if f.scanDirsErr != nil {
		return nil, f.scanDirsErr
	}
	return append([]core.DirEntry(nil), f.scanDirsEntries...), nil
}

func (f *fakeFilesystem) CreateProject(path string) (string, error) {
	f.createProjectCalls = append(f.createProjectCalls, path)
	if f.createProjectErr != nil {
		return "", f.createProjectErr
	}
	if f.createProjectPath != "" {
		return f.createProjectPath, nil
	}
	return path, nil
}

func (f *fakeFilesystem) DeleteProject(projectPath string) error {
	f.deleteProjectCalls = append(f.deleteProjectCalls, projectPath)
	return f.deleteProjectErr
}

func (f *fakeFilesystem) ListWorktreePaths(projectPath string) ([]string, error) {
	f.listWorktreePathsCalls = append(f.listWorktreePathsCalls, projectPath)
	if f.listWorktreePathsErr != nil {
		return nil, f.listWorktreePathsErr
	}
	return append([]string(nil), f.listWorktreePathsResult...), nil
}

func (f *fakeFilesystem) ListWorktrees(projectPath string) (core.WorktreeListing, error) {
	f.listWorktreesCalls = append(f.listWorktreesCalls, projectPath)
	if f.listWorktreesErr != nil {
		return core.WorktreeListing{}, f.listWorktreesErr
	}
	return f.listWorktreesListing, nil
}

func (f *fakeFilesystem) CreateWorktree(projectPath, branchName string) (string, error) {
	f.createWorktreeCalls = append(f.createWorktreeCalls, createWorktreeCall{
		projectPath: projectPath,
		branchName:  branchName,
	})
	if f.createWorktreeErr != nil {
		return "", f.createWorktreeErr
	}
	if f.createWorktreePath != "" {
		return f.createWorktreePath, nil
	}
	return projectPath + "/" + branchName, nil
}

func (f *fakeFilesystem) DeleteWorktree(projectPath, worktreePath string) error {
	f.deleteWorktreeCalls = append(f.deleteWorktreeCalls, deleteWorktreeCall{
		projectPath: projectPath,
		worktree:    worktreePath,
	})
	return f.deleteWorktreeErr
}

func (f *fakeFilesystem) PruneWorktrees(string) error {
	return nil
}

type createWorktreeCall struct {
	projectPath string
	branchName  string
}

type deleteWorktreeCall struct {
	projectPath string
	worktree    string
}

type fakeSessionManager struct {
	openErr          error
	prewarmFn        func(spec core.SessionSpec) (bool, error)
	killErr          error
	listSessionsResp []core.SessionInfo
	listSessionsErr  error
	attachErr        error
	openCalls        []core.SessionSpec
	prewarmCalls     []core.SessionSpec
	killCalls        []core.SessionSpec
	attachCalls      []string
}

func (f *fakeSessionManager) OpenSession(spec core.SessionSpec) error {
	f.openCalls = append(f.openCalls, spec)
	return f.openErr
}

func (f *fakeSessionManager) PrewarmSession(spec core.SessionSpec) (bool, error) {
	f.prewarmCalls = append(f.prewarmCalls, spec)
	if f.prewarmFn != nil {
		return f.prewarmFn(spec)
	}
	return false, nil
}

func (f *fakeSessionManager) KillSession(spec core.SessionSpec) error {
	f.killCalls = append(f.killCalls, spec)
	return f.killErr
}

func (f *fakeSessionManager) ListSessions() ([]core.SessionInfo, error) {
	if f.listSessionsErr != nil {
		return nil, f.listSessionsErr
	}
	return append([]core.SessionInfo(nil), f.listSessionsResp...), nil
}

func (f *fakeSessionManager) AttachSession(name string) error {
	f.attachCalls = append(f.attachCalls, name)
	return f.attachErr
}

func TestScanDirsCmdReturnsScanCompletedMsg(t *testing.T) {
	fs := &fakeFilesystem{
		scanDirsEntries: []core.DirEntry{{Path: "/projects/demo", Name: "demo"}},
	}
	m := New(nil, fs, nil)

	msg := m.scanDirsCmd([]string{"/projects"})()
	scan, ok := msg.(scanCompletedMsg)
	if !ok {
		t.Fatalf("expected scanCompletedMsg, got %T", msg)
	}
	if fs.scanDirsCalls != 1 {
		t.Fatalf("expected one scan call, got %d", fs.scanDirsCalls)
	}
	if fs.scanDirsDepth != 2 {
		t.Fatalf("expected max depth 2, got %d", fs.scanDirsDepth)
	}
	if len(scan.dirs) != 1 || scan.dirs[0].Name != "demo" {
		t.Fatalf("unexpected scan result: %+v", scan.dirs)
	}
}

func TestCreateProjectCmdReturnsProjectCreatedMsg(t *testing.T) {
	fs := &fakeFilesystem{createProjectPath: "/projects/new"}
	m := New(nil, fs, nil)

	msg := m.createProjectCmd("/projects/new")()
	created, ok := msg.(projectCreatedMsg)
	if !ok {
		t.Fatalf("expected projectCreatedMsg, got %T", msg)
	}
	if created.projectPath != "/projects/new" {
		t.Fatalf("unexpected project path: %q", created.projectPath)
	}
}

func TestLoadWorktreesCmdReturnsListing(t *testing.T) {
	fs := &fakeFilesystem{
		listWorktreesListing: core.WorktreeListing{
			Worktrees: []core.Worktree{{Path: "/projects/demo/main", Branch: "main"}},
			Warning:   "note",
		},
	}
	m := New(nil, fs, nil)

	msg := m.loadWorktreesCmd("/projects/demo")()
	loaded, ok := msg.(worktreesLoadedMsg)
	if !ok {
		t.Fatalf("expected worktreesLoadedMsg, got %T", msg)
	}
	if loaded.warning != "note" {
		t.Fatalf("expected warning note, got %q", loaded.warning)
	}
	if len(loaded.worktrees) != 1 || loaded.worktrees[0].Branch != "main" {
		t.Fatalf("unexpected worktree payload: %+v", loaded.worktrees)
	}
}

func TestCreateAndDeleteWorktreeCmds(t *testing.T) {
	fs := &fakeFilesystem{createWorktreePath: "/projects/demo/feature-x"}
	m := New(nil, fs, nil)

	createdMsg := m.createWorktreeCmd("/projects/demo", "feature-x")()
	created, ok := createdMsg.(worktreeCreatedMsg)
	if !ok {
		t.Fatalf("expected worktreeCreatedMsg, got %T", createdMsg)
	}
	if created.path != "/projects/demo/feature-x" {
		t.Fatalf("unexpected created path %q", created.path)
	}

	deletedMsg := m.deleteWorktreeCmd("/projects/demo", "/projects/demo/feature-x")()
	deleted, ok := deletedMsg.(worktreeDeletedMsg)
	if !ok {
		t.Fatalf("expected worktreeDeletedMsg, got %T", deletedMsg)
	}
	if deleted.path != "/projects/demo/feature-x" {
		t.Fatalf("unexpected deleted path %q", deleted.path)
	}
	if len(fs.deleteWorktreeCalls) != 1 {
		t.Fatalf("expected one delete call, got %d", len(fs.deleteWorktreeCalls))
	}
}

func TestDeleteProjectCmdKillsSessionsBeforeDeleting(t *testing.T) {
	fs := &fakeFilesystem{
		listWorktreePathsResult: []string{
			"/projects/demo/main",
			"/projects/demo/feature",
		},
	}
	sessions := &fakeSessionManager{}
	m := New(nil, fs, sessions)

	msg := m.deleteProjectCmd("/projects/demo")()
	deleted, ok := msg.(projectDeletedMsg)
	if !ok {
		t.Fatalf("expected projectDeletedMsg, got %T", msg)
	}
	if deleted.err != nil {
		t.Fatalf("unexpected delete error: %v", deleted.err)
	}
	if len(fs.deleteProjectCalls) != 1 || fs.deleteProjectCalls[0] != "/projects/demo" {
		t.Fatalf("unexpected delete project calls: %v", fs.deleteProjectCalls)
	}
	expectedKills := len(fs.listWorktreePathsResult) * len(m.core.Tools)
	if len(sessions.killCalls) != expectedKills {
		t.Fatalf("expected %d kill-session calls, got %d", expectedKills, len(sessions.killCalls))
	}
}

func TestDeleteProjectCmdStopsOnSessionKillError(t *testing.T) {
	fs := &fakeFilesystem{
		listWorktreePathsResult: []string{"/projects/demo/main"},
	}
	sessions := &fakeSessionManager{killErr: errors.New("kill failed")}
	m := New(nil, fs, sessions)

	msg := m.deleteProjectCmd("/projects/demo")()
	deleted, ok := msg.(projectDeletedMsg)
	if !ok {
		t.Fatalf("expected projectDeletedMsg, got %T", msg)
	}
	if deleted.err == nil || deleted.err.Error() != "kill failed" {
		t.Fatalf("expected kill failed error, got %v", deleted.err)
	}
	if len(fs.deleteProjectCalls) != 0 {
		t.Fatalf("did not expect project delete call on kill error")
	}
}

func TestDeleteWorktreeCmdStopsOnSessionKillError(t *testing.T) {
	fs := &fakeFilesystem{}
	sessions := &fakeSessionManager{killErr: errors.New("kill failed")}
	m := New(nil, fs, sessions)

	msg := m.deleteWorktreeCmd("/projects/demo", "/projects/demo/main")()
	deleted, ok := msg.(worktreeDeletedMsg)
	if !ok {
		t.Fatalf("expected worktreeDeletedMsg, got %T", msg)
	}
	if deleted.err == nil || deleted.err.Error() != "kill failed" {
		t.Fatalf("expected kill failed error, got %v", deleted.err)
	}
	if len(fs.deleteWorktreeCalls) != 0 {
		t.Fatalf("did not expect delete worktree call on kill error")
	}
}

func TestPrewarmAllToolsCmdReturnsMessagesForCreatedExistingAndFailed(t *testing.T) {
	sessions := &fakeSessionManager{
		prewarmFn: func(spec core.SessionSpec) (bool, error) {
			switch spec.Tool {
			case "amp":
				return true, nil
			case "codex":
				return false, nil
			default:
				return false, errors.New("prewarm failed")
			}
		},
	}
	m := New(nil, &fakeFilesystem{}, sessions)

	cmd := m.prewarmAllToolsCmd("/projects/demo/main", []string{"amp", "codex", "claude"})
	msgs := runCmd(cmd)
	if len(msgs) != 3 {
		t.Fatalf("expected three messages, got %d", len(msgs))
	}

	seenStarted := false
	seenExisting := false
	seenFailed := false
	for _, msg := range msgs {
		switch typed := msg.(type) {
		case core.MsgToolPrewarmStarted:
			seenStarted = typed.Tool == "amp" && !typed.StartedAt.IsZero()
		case core.MsgToolPrewarmExisting:
			seenExisting = typed.Tool == "codex"
		case core.MsgToolPrewarmFailed:
			seenFailed = typed.Tool == "claude" && typed.Err != nil
		}
	}
	if !seenStarted || !seenExisting || !seenFailed {
		t.Fatalf("expected started/existing/failed messages, got %#v", msgs)
	}
}

func TestPrewarmAllToolsCmdReturnsNilWithoutSessions(t *testing.T) {
	m := New(nil, &fakeFilesystem{}, nil)
	if cmd := m.prewarmAllToolsCmd("/projects/demo/main", []string{"amp"}); cmd != nil {
		t.Fatalf("expected nil command when sessions manager is missing")
	}
}

func TestListSessionsCmdAndAttachSessionCmd(t *testing.T) {
	sessions := &fakeSessionManager{
		listSessionsResp: []core.SessionInfo{{Name: "demo__amp"}},
	}
	m := New(nil, &fakeFilesystem{}, sessions)

	msg := m.listSessionsCmd()()
	loaded, ok := msg.(sessionsLoadedMsg)
	if !ok {
		t.Fatalf("expected sessionsLoadedMsg, got %T", msg)
	}
	if len(loaded.sessions) != 1 || loaded.sessions[0].Name != "demo__amp" {
		t.Fatalf("unexpected sessions payload: %+v", loaded.sessions)
	}

	attachMsg := m.attachSessionCmd(core.SessionInfo{Name: "demo__amp"})()
	attached, ok := attachMsg.(sessionAttachedMsg)
	if !ok {
		t.Fatalf("expected sessionAttachedMsg, got %T", attachMsg)
	}
	if attached.err != nil || attached.session.Name != "demo__amp" {
		t.Fatalf("unexpected attached payload: %+v", attached)
	}
}

func TestAttachSessionCmdWithoutManagerReturnsError(t *testing.T) {
	m := New(nil, &fakeFilesystem{}, nil)

	msg := m.attachSessionCmd(core.SessionInfo{Name: "demo__amp"})()
	attached, ok := msg.(sessionAttachedMsg)
	if !ok {
		t.Fatalf("expected sessionAttachedMsg, got %T", msg)
	}
	if !errors.Is(attached.err, errNoSessions) {
		t.Fatalf("expected errNoSessions, got %v", attached.err)
	}
}

func TestCheckToolReadyCmdImmediateWhenWarmStartIsZero(t *testing.T) {
	m := New(nil, &fakeFilesystem{}, nil)
	m.core.ToolWarmStart = map[string]time.Time{
		"amp": time.Time{},
	}

	msg := m.checkToolReadyCmd(core.SessionSpec{Tool: "amp"})()
	delay, ok := msg.(core.MsgToolDelayElapsed)
	if !ok {
		t.Fatalf("expected MsgToolDelayElapsed, got %T", msg)
	}
	if delay.Tool != "amp" {
		t.Fatalf("expected amp delay message, got %q", delay.Tool)
	}
}

func TestRunEffectsIncludesQuitCommands(t *testing.T) {
	m := New(nil, &fakeFilesystem{}, nil)
	msgs := runCmd(m.runEffects([]core.Effect{
		core.EffOpenSession{Spec: core.SessionSpec{DirPath: "/projects/demo/main", Tool: "amp"}},
		core.EffQuit{},
	}))
	if len(msgs) != 2 {
		t.Fatalf("expected two quit messages, got %d", len(msgs))
	}
	for _, msg := range msgs {
		if _, ok := msg.(tea.QuitMsg); !ok {
			t.Fatalf("expected tea.QuitMsg, got %T", msg)
		}
	}
}

func runCmd(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}

	msg := cmd()
	if msg == nil {
		return nil
	}

	if batch, ok := msg.(tea.BatchMsg); ok {
		var out []tea.Msg
		for _, nested := range batch {
			out = append(out, runCmd(nested)...)
		}
		return out
	}

	return []tea.Msg{msg}
}
