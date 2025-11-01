package manifest_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/owen-6936/termplex/assert"
	"github.com/owen-6936/termplex/manifest"
)

func TestLoadFromFile_Success(t *testing.T) {
	// 1. Create a temporary manifest file with known content.
	content := []byte(`{
		"sessionName": "TestSession",
		"sessionTags": { "project": "tester" },
		"windows": [
			{
				"windowName": "TestWindow",
				"panes": [
					{
						"startupShell": { "interactive": true, "command": ["bash", "-i"] },
						"startupCommands": ["echo 'hello'"]
					}
				]
			}
		]
	}`)

	// Use t.TempDir() to create a temporary directory that is automatically cleaned up.
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.termplex.json")
	err := os.WriteFile(filePath, content, 0644)
	assert.NoError(t, err)

	// 2. Call LoadFromFile with the path to the temporary file.
	m, err := manifest.LoadFromFile(filePath)
	assert.NoError(t, err)

	// 3. Assert that the parsed manifest contains the correct data.
	if m.SessionName != "TestSession" {
		t.Errorf("expected sessionName to be 'TestSession', got %q", m.SessionName)
	}
	if m.SessionTags["project"] != "tester" {
		t.Errorf("expected session tag 'project' to be 'tester', got %q", m.SessionTags["project"])
	}
	if len(m.Windows) != 1 {
		t.Fatalf("expected 1 window, got %d", len(m.Windows))
	}
	if m.Windows[0].WindowName != "TestWindow" {
		t.Errorf("expected windowName to be 'TestWindow', got %q", m.Windows[0].WindowName)
	}
	if len(m.Windows[0].Panes) != 1 {
		t.Fatalf("expected 1 pane, got %d", len(m.Windows[0].Panes))
	}
	if !m.Windows[0].Panes[0].StartupShell.Interactive {
		t.Error("expected startupShell.interactive to be true")
	}
	if m.Windows[0].Panes[0].StartupCommands[0] != "echo 'hello'" {
		t.Errorf("expected startup command to be 'echo 'hello'', got %q", m.Windows[0].Panes[0].StartupCommands[0])
	}
}

func TestLoadFromFile_FileNotFound(t *testing.T) {
	// Attempt to load a file that does not exist.
	_, err := manifest.LoadFromFile("non-existent-file.json")

	// Assert that an error was returned.
	if err == nil {
		t.Fatal("expected an error when loading a non-existent file, but got nil")
	}
}

func TestLoadFromFile_InvalidJSON(t *testing.T) {
	// 1. Create a temporary file with malformed JSON.
	content := []byte(`{
		"sessionName": "InvalidSession",
		"windows": [
	}`) // Missing closing brace and bracket

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "invalid.json")
	err := os.WriteFile(filePath, content, 0644)
	assert.NoError(t, err)

	// 2. Attempt to load the malformed file.
	_, err = manifest.LoadFromFile(filePath)

	// 3. Assert that a JSON parsing error was returned.
	if err == nil {
		t.Fatal("expected an error when parsing invalid JSON, but got nil")
	}
	assert.Contains(t, err.Error(), "parse manifest JSON")
}
