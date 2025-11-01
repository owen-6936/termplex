package session_test

import (
	"testing"
	"time"

	"github.com/owen-6936/termplex/assert"
	"github.com/owen-6936/termplex/testenv"
)

func TestCreateSessionFromManifest_Harness(t *testing.T) {
	// 1. Use the new test harness to create a session from the example manifest.
	// All setup and teardown is handled automatically.
	sm, sessionID := testenv.NewSessionFromManifest(t, "../example.termplex.json")

	// 2. Get the created session and its components to validate them.
	s, exists := sm.GetSession(sessionID)
	assert.True(t, exists, "Session with ID %s should exist", sessionID)

	// 3. Assert that the session was created with the correct data.
	assert.True(t, s.Name == "WebAppDev", "Expected session name 'WebAppDev', got %s", s.Name)
	assert.True(t, s.Tags["project"] == "termplex-demo", "Expected session tag 'project' to be 'termplex-demo'")

	// 4. Assert that the window and its panes were created.
	assert.True(t, len(s.WindowRefs) == 1, "Expected session to have 1 window, got %d", len(s.WindowRefs))

	var windowID string
	for id := range s.WindowRefs {
		windowID = id // Get the ID of the single window.
	}

	wm, exists := sm.Windows[windowID]
	assert.True(t, exists, "Window with ID %s should exist in SessionManager", windowID)
	assert.True(t, len(wm.Panes) == 2, "Expected window to have 2 panes, got %d", len(wm.Panes))

	// 5. Briefly check that a shell process is running in one of the panes.
	// A more detailed test could inspect the output buffers.
	time.Sleep(200 * time.Millisecond) // Allow time for shells to spawn.
	for _, pane := range wm.Panes {
		assert.True(t, pane.InteractiveShell != nil || len(pane.NonInteractiveShells) > 0, "Expected pane %s to have at least one shell", pane.ID)
	}
}
