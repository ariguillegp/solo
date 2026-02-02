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
