package core

import (
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

type SessionInfo struct {
	Name    string
	DirPath string
	Tool    string
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

type ToolDefinition struct {
	Name string
	Env  []string
}

const ToolNone = "none"

var toolDefinitions = []ToolDefinition{
	{
		Name: "opencode",
		Env:  []string{`OPENCODE_CONFIG_CONTENT={"theme":"gruvbox"}`},
	},
	{Name: "claude"},
	{Name: "amp"},
	{Name: ToolNone},
}

func SupportedTools() []string {
	names := make([]string, 0, len(toolDefinitions))
	for _, tool := range toolDefinitions {
		names = append(names, tool.Name)
	}
	return names
}

func IsSupportedTool(tool string) bool {
	for _, allowed := range toolDefinitions {
		if tool == allowed.Name {
			return true
		}
	}
	return false
}

func ToolEnv(tool string) []string {
	tool = strings.TrimSpace(tool)
	if tool == "" {
		return nil
	}
	for _, def := range toolDefinitions {
		if def.Name == tool {
			return append([]string(nil), def.Env...)
		}
	}
	return nil
}

func ToolNeedsWarmup(tool string) bool {
	tool = strings.TrimSpace(tool)
	if tool == "" {
		return false
	}
	return tool != ToolNone
}

func SanitizeWorktreeName(name string) string {
	clean := strings.ReplaceAll(name, " ", "-")
	return clean
}

func IsValidWorktreeName(name string) bool {
	if strings.TrimSpace(name) == "" {
		return false
	}
	if strings.HasPrefix(name, "/") || strings.HasSuffix(name, "/") {
		return false
	}
	if strings.Contains(name, "//") {
		return false
	}
	for _, segment := range strings.Split(name, "/") {
		if segment == "." || segment == ".." || segment == "" {
			return false
		}
	}
	for i := 0; i < len(name); i++ {
		ch := name[i]
		if ch == '/' || ch == '-' || ch == '_' {
			continue
		}
		if (ch < '0' || ch > '9') && (ch < 'A' || ch > 'Z') && (ch < 'a' || ch > 'z') {
			return false
		}
	}
	return true
}
