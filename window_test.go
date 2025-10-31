package termplex

import (
	"strings"
	"testing"
	"time"
)

func TestNewWindow(t *testing.T) {
    session := "test-session-window"
    windowName := "test-window"

    // Ensure clean session
    _ = KillSession(session)
    err := NewSession(session)
    if err != nil {
        t.Fatalf("Failed to create session: %v", err)
    }

    // Create window
    window, err := NewWindow(session, windowName)
    if err != nil {
        t.Fatalf("Failed to create window: %v", err)
    }
    if window.WindowName == "" {
        t.Errorf("Expected window name, got empty string")
    }

    // Send command to active pane
    err = window.SendKeys("cd /tmp && exec bash")
    if err != nil {
        t.Fatalf("Failed to send keys: %v", err)
    }

    // Give tmux time to process
    time.Sleep(200 * time.Millisecond)

    // Check working directory
    target := session + ":" + window.WindowName + ".0"
    out, err := runTmux("display-message", "-p", "-F", "#{pane_current_path}", "-t", target)
    if err != nil {
        t.Fatalf("Failed to get pane path: %v", err)
    }
    path := strings.TrimSpace(out)
    if path != "/tmp" {
        t.Errorf("Expected /tmp, got %q", path)
    }

    // Cleanup
    _ = KillSession(session)
}
