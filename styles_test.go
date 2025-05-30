package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadStyleConfig_WithTempFile(t *testing.T) {
	const configContent = `{
	  "showProjects": true,
	  "showTags": true,
	  "showFileLines": false,
	  "emojis": {
	    "root": "🎭",
	    "file": "🧪",
	    "suite": "📁"
	  },
	  "styles": [
	    {"name": "enumerator", "color": "150", "bold": true, "italic": false, "faint": false},
	    {"name": "root",       "color": "7",   "bold": true, "italic": false, "faint": false},
	    {"name": "item",       "color": "7",   "bold": true, "italic": false, "faint": false},
	    {"name": "tag",        "color": "6",   "bold": false, "italic": false, "faint": false},
	    {"name": "project",    "color": "8",   "bold": false, "italic": true, "faint": true},
	    {"name": "fileLine",   "color": "7",   "bold": false, "italic": true, "faint": true},
	    {"name": "skipped",    "color": "12",  "bold": false, "italic": true, "faint": false},
	    {"name": "fixme",      "color": "3",   "bold": false, "italic": true, "faint": false},
	    {"name": "fail",       "color": "5",   "bold": false, "italic": true, "faint": false},
	    {"name": "test",       "color": "7",   "bold": false, "italic": false, "faint": false},
	    {"name": "counter",    "color": "7",   "bold": false, "italic": true, "faint": true},
	    {"name": "file",       "color": "7",   "bold": true, "italic": false, "faint": false},
	    {"name": "suite",      "color": "7",   "bold": true, "italic": false, "faint": false}
	  ]
	}`

	// Create a temp directory and override working directory
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, ".pwtree.json")
	if err := os.WriteFile(tmpFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write temp config: %v", err)
	}

	// Set working directory to the temp dir
	originalWD, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change working dir: %v", err)
	}
	defer os.Chdir(originalWD)

	ci := false
	ciMode = &ci

	styles, display, emojis := loadStyleConfig()

	if !display.ShowProjects {
		t.Error("Expected ShowProjects = true")
	}
	if !display.ShowTags {
		t.Error("Expected ShowTags = true")
	}
	if display.ShowFileLines {
		t.Error("Expected ShowFileLines = false")
	}

	if emojis.Root != "🎭" {
		t.Errorf("Expected Root emoji 🎭, got %q", emojis.Root)
	}
	if emojis.File != "🧪" {
		t.Errorf("Expected File emoji 🧪, got %q", emojis.File)
	}
	if emojis.Suite != "📁" {
		t.Errorf("Expected Suite emoji 📁, got %q", emojis.Suite)
	}

	if !styles["root"].GetBold() {
		t.Error("Expected 'root' to be bold")
	}
	if styles["tag"].GetBold() {
		t.Error("Expected 'tag' to NOT be bold")
	}
	if !styles["project"].GetItalic() || !styles["project"].GetFaint() {
		t.Error("Expected 'project' to be italic and faint")
	}
	if !styles["fail"].GetItalic() {
		t.Error("Expected 'fail' to be italic")
	}
	if !styles["enumerator"].GetBold() {
		t.Error("Expected 'enumerator' to be bold")
	}
}
