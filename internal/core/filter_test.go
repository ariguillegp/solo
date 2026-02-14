package core

import (
	"strings"
	"testing"
)

func TestFilterWorktreesMatchesSanitizedBranch(t *testing.T) {
	worktrees := []Worktree{{Path: "/repo/feat", Name: "rivet--feature-test", Branch: "feature/test"}}

	filtered := FilterWorktrees(worktrees, "feature test")
	if len(filtered) != 1 {
		t.Fatalf("expected sanitized branch to match query, got %d", len(filtered))
	}
}

func TestFilterDirsIncludesVeryLongFuzzyMatchesWithNegativeScore(t *testing.T) {
	longName := strings.Repeat("z", 120) + "a"
	dirs := []DirEntry{{Name: longName}, {Name: "bbb"}}

	filtered := FilterDirs(dirs, "a")
	if len(filtered) != 1 {
		t.Fatalf("expected long fuzzy subsequence match to be retained, got %d", len(filtered))
	}
	if filtered[0].Name != longName {
		t.Fatalf("expected long fuzzy match to be returned, got %q", filtered[0].Name)
	}
}

func TestFilterDirsRanksPrefixThenFuzzyWithStableFallback(t *testing.T) {
	dirs := []DirEntry{
		{Name: "alpha"},
		{Name: "alphabet"},
		{Name: "alpine"},
		{Name: "beta"},
	}

	filtered := FilterDirs(dirs, "alp")
	if len(filtered) != 3 {
		t.Fatalf("expected 3 matches, got %d", len(filtered))
	}

	if filtered[0].Name != "alpha" {
		t.Fatalf("expected exact-prefix hits first with stable order, got %q", filtered[0].Name)
	}
	if filtered[1].Name != "alphabet" {
		t.Fatalf("expected stable ordering for same-score prefix hits, got %q", filtered[1].Name)
	}
	if filtered[2].Name != "alpine" {
		t.Fatalf("expected fuzzy hit after prefix hits, got %q", filtered[2].Name)
	}
}

func TestFilterWorktreesRanksExactBeforePrefixAndFuzzy(t *testing.T) {
	worktrees := []Worktree{
		{Name: "feature"},
		{Name: "feature-long"},
		{Name: "fix-feature"},
	}

	filtered := FilterWorktrees(worktrees, "feature")
	if len(filtered) != 3 {
		t.Fatalf("expected 3 matches, got %d", len(filtered))
	}

	if filtered[0].Name != "feature" {
		t.Fatalf("expected exact name first, got %q", filtered[0].Name)
	}
	if filtered[1].Name != "feature-long" {
		t.Fatalf("expected prefix second, got %q", filtered[1].Name)
	}
	if filtered[2].Name != "fix-feature" {
		t.Fatalf("expected fuzzy match last, got %q", filtered[2].Name)
	}
}

func TestFilterToolsRanksByBestMatch(t *testing.T) {
	tools := []string{"codex", "claude", "opencode"}
	filtered := FilterTools(tools, "cod")
	if len(filtered) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(filtered))
	}
	if filtered[0] != "codex" {
		t.Fatalf("expected exact prefix tool first, got %q", filtered[0])
	}
	if filtered[1] != "opencode" {
		t.Fatalf("expected fuzzy tool second, got %q", filtered[1])
	}
}

func TestFilterSessionsRanksNamePrefixBeforePathAndTool(t *testing.T) {
	sessions := []SessionInfo{
		{Name: "alpha", DirPath: "/tmp/zzz", Tool: "none"},
		{Name: "beta", DirPath: "/projects/alpha", Tool: "none"},
		{Name: "gamma", DirPath: "/tmp/zzz", Tool: "alpha-tool"},
	}

	filtered := FilterSessions(sessions, "alpha")
	if len(filtered) != 3 {
		t.Fatalf("expected 3 matches, got %d", len(filtered))
	}
	if filtered[0].Name != "alpha" {
		t.Fatalf("expected session name exact/prefix to rank first, got %q", filtered[0].Name)
	}
	if filtered[1].Name != "gamma" {
		t.Fatalf("expected tool prefix match second, got %q", filtered[1].Name)
	}
	if filtered[2].Name != "beta" {
		t.Fatalf("expected path fuzzy match third, got %q", filtered[2].Name)
	}
}

func TestBrowsingCreateRowStillComesAfterRankedMatches(t *testing.T) {
	m := Model{
		Mode:      ModeBrowsing,
		RootPaths: []string{"/projects"},
		Query:     "app",
		Filtered: []DirEntry{
			{Name: "app"},
			{Name: "apple"},
		},
	}

	updated, _, handled := UpdateKey(m, "down")
	if !handled {
		t.Fatal("expected down to be handled")
	}
	if updated.SelectedIdx != 1 {
		t.Fatalf("expected to move to second ranked match first, got %d", updated.SelectedIdx)
	}

	updated, _, _ = UpdateKey(updated, "down")
	if updated.SelectedIdx != 2 {
		t.Fatalf("expected to move to create row after ranked matches, got %d", updated.SelectedIdx)
	}

	updated, _, _ = UpdateKey(updated, "down")
	if updated.SelectedIdx != 2 {
		t.Fatalf("expected create row to remain max selection, got %d", updated.SelectedIdx)
	}
}
