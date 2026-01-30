package core

import (
	"os"
	"strings"
	"time"
)

type DirEntry struct {
	Path     string
	Name     string
	Score    int
	Exists   bool
	LastUsed time.Time
}

type SessionSpec struct {
	DirPath string
	Tool    string
	Detach  bool
}

type Worktree struct {
	Path   string
	Name   string
	Branch string
}

type WorktreeListing struct {
	Worktrees []Worktree
	Warning   string
}

var defaultTools = []string{"opencode", "amp"}

var AllowedTools = append([]string(nil), defaultTools...)

func ConfigureAllowedToolsFromEnv() {
	raw := strings.TrimSpace(os.Getenv("SOLO_TOOLS"))
	if raw == "" {
		return
	}
	tools := parseToolList(raw)
	if len(tools) == 0 {
		return
	}
	ConfigureAllowedTools(tools)
}

func ConfigureAllowedTools(tools []string) {
	if len(tools) == 0 {
		AllowedTools = append([]string(nil), defaultTools...)
		return
	}
	AllowedTools = uniqueTools(tools)
}

func SupportedTools() []string {
	return append([]string(nil), AllowedTools...)
}

func IsSupportedTool(tool string) bool {
	for _, allowed := range AllowedTools {
		if tool == allowed {
			return true
		}
	}
	return false
}

func parseToolList(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		switch r {
		case ',', '\n', '\t', ' ':
			return true
		default:
			return false
		}
	})
	if len(parts) == 0 {
		return nil
	}
	return uniqueTools(parts)
}

func uniqueTools(tools []string) []string {
	seen := make(map[string]struct{}, len(tools))
	unique := make([]string, 0, len(tools))
	for _, tool := range tools {
		clean := strings.TrimSpace(tool)
		if clean == "" {
			continue
		}
		if _, ok := seen[clean]; ok {
			continue
		}
		seen[clean] = struct{}{}
		unique = append(unique, clean)
	}
	return unique
}

func SanitizeWorktreeName(name string) string {
	clean := strings.ReplaceAll(name, "/", "-")
	clean = strings.ReplaceAll(clean, " ", "-")
	return clean
}
