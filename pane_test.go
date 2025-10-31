package termplex

import (
	"strings"
	"testing"
	"time"
)

func TestNewPaneAndShellLifecycle(t *testing.T) {
    session := "test-session-pane"
    windowName := "test-window"

    // Ensure clean session
    _ = KillSession(session)
    if err := NewSession(session); err != nil {
        t.Fatalf("Failed to create session: %v", err)
    }

    // Create initial window
    window, err := NewWindow(session, windowName)
    if err != nil {
        t.Fatalf("Failed to create window: %v", err)
    }

    // Create a new pane in the window
    pane, err := NewPane(session, window.WindowName)
    if err != nil {
        t.Fatalf("Failed to create new pane: %v", err)
    }

    // Start a shell in /tmp
    err = pane.StartShell("/tmp")
    if err != nil {
        t.Fatalf("Failed to start shell in pane: %v", err)
    }

    time.Sleep(300 * time.Millisecond)

    // Validate working directory
    path, err := pane.GetCurrentPath()
    if err != nil {
        t.Fatalf("Failed to get current path: %v", err)
    }
    if path != "/tmp" {
        t.Errorf("Expected /tmp, got %q", path)
    }

    // Validate active command
    cmd, err := pane.GetCurrentCommand()
    if err != nil {
        t.Fatalf("Failed to get current command: %v", err)
    }
    if !strings.Contains(cmd, "bash") {
        t.Errorf("Expected bash, got %q", cmd)
    }

    // Cleanup
    _ = KillSession(session)
}
