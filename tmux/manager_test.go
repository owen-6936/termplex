package tmux_test

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/owen-6936/termplex/tmux"
)

// TestTmuxBackendLifecycle is an integration test that requires a real tmux server.
// It validates the full lifecycle: creating a session, adding panes, sending commands,
// capturing output, and terminating the session.
func TestTmuxBackendLifecycle(t *testing.T) {
	// Skip this test if tmux is not installed.
	if !isTmuxAvailable() {
		t.Skip("tmux command not found, skipping integration test")
	}

	// 1. Create a new, detached tmux session with a unique name for this test run.
	sessionName := fmt.Sprintf("termplex-test-%d", time.Now().UnixNano())
	tmuxManager, err := tmux.NewSessionManager(sessionName)
	if err != nil {
		t.Fatalf("Failed to create tmux session: %v", err)
	}
	// Use t.Cleanup to guarantee the tmux session is killed after the test.
	t.Cleanup(func() {
		err := tmuxManager.KillSession()
		if err != nil {
			// Log the error but don't fail the test, as the primary test logic is complete.
			t.Logf("Warning: failed to kill tmux session %s: %v", sessionName, err)
		}
	})

	// 2. Get the initial pane and add a second one.
	if len(tmuxManager.Panes) != 1 {
		t.Fatal("NewSessionManager should create one initial pane")
	}
	pane0 := tmuxManager.Panes[0]

	pane1, err := tmuxManager.AddPane()
	if err != nil {
		t.Fatalf("Failed to add a new pane: %v", err)
	}

	// 3. Send unique commands to each pane.
	err = pane0.SendKeys("echo 'output from pane 0'")
	if err != nil {
		t.Fatalf("Failed to send keys to pane 0: %v", err)
	}
	err = pane1.SendKeys("echo 'output from pane 1'")
	if err != nil {
		t.Fatalf("Failed to send keys to pane 1: %v", err)
	}

	// 4. Wait for commands to execute and capture the output.
	time.Sleep(200 * time.Millisecond) // Give tmux time to process commands.

	output0, err := pane0.Capture()
	if err != nil || !strings.Contains(output0, "output from pane 0") {
		t.Errorf("Expected output from pane 0, but got: %q (err: %v)", output0, err)
	}

	output1, err := pane1.Capture()
	if err != nil || !strings.Contains(output1, "output from pane 1") {
		t.Errorf("Expected output from pane 1, but got: %q (err: %v)", output1, err)
	}
}

// isTmuxAvailable checks if the tmux command exists in the system's PATH.
func isTmuxAvailable() bool {
	_, err := exec.LookPath("tmux")
	return err == nil
}
