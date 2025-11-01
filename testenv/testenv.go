package testenv

import (
	"os"
	"os/exec"
	"testing"
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
