package core

import (
	"testing"
)

func TestIsSupportedTool_DefaultTools(t *testing.T) {
	if !IsSupportedTool("opencode") {
		t.Error("IsSupportedTool(opencode) = false, want true")
	}
	if !IsSupportedTool("amp") {
		t.Error("IsSupportedTool(amp) = false, want true")
	}
	if !IsSupportedTool("claude") {
		t.Error("IsSupportedTool(claude) = false, want true")
	}
	if !IsSupportedTool("codex") {
		t.Error("IsSupportedTool(codex) = false, want true")
	}
	if !IsSupportedTool(ToolNone) {
		t.Error("IsSupportedTool(none) = false, want true")
	}
	if IsSupportedTool("nonexistent") {
		t.Error("IsSupportedTool(nonexistent) = true, want false")
	}
}
func TestSupportedTools_ReturnsCopy(t *testing.T) {
	got := SupportedTools()
	got[0] = "modified"

	original := SupportedTools()
	if original[0] == "modified" {
		t.Error("SupportedTools() returned a reference, not a copy")
	}
}
