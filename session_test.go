package termplex

import (
	"testing"
)

func TestSessionLifecycle(t *testing.T) {
    session := "test-session-lifecycle"

    // Ensure clean start
    _ = KillSession(session)

    // Create session
    err := NewSession(session)
    if err != nil {
        t.Fatalf("Failed to create session: %v", err)
    }

    // Verify session exists
    if !HasSession(session) {
        t.Errorf("Expected session %q to exist, but it doesn't", session)
    }

    // Kill session
    err = KillSession(session)
    if err != nil {
        t.Fatalf("Failed to kill session: %v", err)
    }

    // Verify session is gone
    if HasSession(session) {
        t.Errorf("Expected session %q to be gone, but it still exists", session)
    }
}
