package testenv

import (
	"os"
	"os/exec"
	"testing"

	"github.com/owen-6936/termplex/assert"
	"github.com/owen-6936/termplex/session"
)

// IsInCI checks if the test is running in a common Continuous Integration environment.
func IsInCI() bool {
	// GitHub Actions, Travis CI, CircleCI, GitLab CI, and others set this.
	return os.Getenv("CI") != ""
}

// isTmuxAvailable checks if the tmux command exists in the system's PATH.
func isTmuxAvailable() bool {
	_, err := exec.LookPath("tmux")
	return err == nil
}

// SkipIfNoTmux skips the current test if the tmux command is not available.
func SkipIfNoTmux(t *testing.T) {
	if !isTmuxAvailable() {
		t.Skip("tmux command not found, skipping integration test")
	}
}

// NewSessionFromManifest provides a reproducible test harness for creating a session
// from a manifest file. It handles session creation, error checking, and automatic
// teardown via t.Cleanup. It returns the session manager and the ID of the created session.
func NewSessionFromManifest(t *testing.T, manifestPath string) (*session.SessionManager, string) {
	t.Helper()

	// 1. Create a new SessionManager.
	sm := session.NewSessionManager(5) // Default window limit for tests.

	// 2. Create the session from the manifest.
	sessionID, err := sm.CreateSessionFromManifest(manifestPath)
	assert.NoError(t, err)
	assert.True(t, sessionID != "", "CreateSessionFromManifest returned an empty session ID")

	// 3. Register a cleanup function to terminate the session after the test.
	t.Cleanup(func() {
		err := sm.TerminateSession(sessionID)
		assert.NoError(t, err)
	})

	return sm, sessionID
}
