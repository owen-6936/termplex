package shell_test

import (
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/owen-6936/termplex/assert"
	"github.com/owen-6936/termplex/shell"
)

// newTestShell is a helper function to create a self-contained shell.ShellSession
// for testing, completely bypassing the pane and window managers. This ensures
// that tests for the 'shell' package are true unit tests.
func newTestShell(t *testing.T, command ...string) *shell.ShellSession {
	t.Helper()

	if len(command) == 0 {
		t.Fatal("newTestShell requires a command to execute")
	}
	cmd := exec.Command(command[0], command[1:]...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("Failed to create stdin pipe: %v", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("Failed to create stdout pipe: %v", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("Failed to create stderr pipe: %v", err)
	}

	session := &shell.ShellSession{
		ID:          fmt.Sprintf("test-shell-%d", time.Now().UnixNano()),
		Cmd:         cmd,
		Stdin:       stdin,
		Stdout:      stdout,
		Stderr:      stderr,
		StartedAt:   time.Now(),
		Interactive: true, // Assume interactive for most tests
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start command %v", command)
	}

	// Automatically start capturing output to the session's internal buffers.
	session.StartReading(session.OutputHandler, session.ErrorOutputHandler)

	// Ensure the shell is terminated when the test completes.
	t.Cleanup(func() {
		_ = session.Close(2 * time.Second)
	})

	return session
}

func TestShellLifecycleAndOutputBuffering(t *testing.T) {
	// Spawn a shell that writes to both stdout and stderr
	command := []string{"bash", "-c", "echo 'hello stdout' && >&2 echo 'hello stderr'"}
	session := newTestShell(t, command...)

	// The process runs and exits quickly. Wait a moment for output to be captured.
	time.Sleep(200 * time.Millisecond)

	// Check if stdout was captured
	stdout := session.OutputBuf.String()
	assert.Contains(t, stdout, "hello stdout")

	// Check if stderr was captured
	stderr := session.StderrBuf.String()
	assert.Contains(t, stderr, "hello stderr")
}

func TestSendCommandAndWait(t *testing.T) {
	// Spawn a persistent interactive shell
	session := newTestShell(t, "bash", "-i")

	// Give the shell a moment to initialize
	time.Sleep(200 * time.Millisecond)

	// Send a command and wait for the response
	testPhrase := "this is a synchronous test"
	output, err := session.SendCommandAndWait("echo '" + testPhrase + "'")
	assert.NoError(t, err)

	// The output should contain our phrase, trimmed of whitespace and shell prompts
	assert.Contains(t, output, testPhrase)

	// Verify the output was also captured in the main buffer
	fullOutput := session.OutputBuf.String()
	assert.Contains(t, fullOutput, testPhrase)
}

func TestSendCommandAndWait_NoPromptLeakage(t *testing.T) {
	// This test ensures that the output from SendCommandAndWait does not contain
	// the shell prompt, which can include ANSI escape codes.
	session := newTestShell(t, "bash", "-i")
	time.Sleep(200 * time.Millisecond) // Allow shell to initialize

	output, err := session.SendCommandAndWait("echo -n 'clean output'")
	assert.NoError(t, err)

	// Check that the output is exactly what we expect and nothing more.
	if output != "clean output" {
		t.Errorf("Expected output to be exactly 'clean output', but got %q", output)
	}
}

func TestShellCloseWithFallback(t *testing.T) {
	// Spawn a shell that ignores the standard 'exit' signal by trapping SIGTERM
	// and then sleeps, forcing our fallback to trigger.
	command := []string{"bash", "-c", "trap '' TERM; echo 'ready'; sleep 5"}
	session := newTestShell(t, command...)

	// Wait for the "ready" signal to ensure the process is running
	time.Sleep(200 * time.Millisecond)
	assert.Contains(t, session.OutputBuf.String(), "ready")

	// Attempt to close the session with a very short grace period.
	// This should fail gracefully and trigger the force-kill fallback.
	gracePeriod := 100 * time.Millisecond
	startTime := time.Now()
	err := session.Close(gracePeriod)
	duration := time.Since(startTime)

	// After a force-kill, Wait() returns an error. We expect this.
	assert.True(t, err != nil, "Expected an error from Wait() after a forced kill, but got nil")

	// The close operation should have taken longer than the grace period because
	// it had to wait for the timeout before force-killing.
	assert.True(t, duration > gracePeriod, "The close operation finished too quickly (%v), suggesting the fallback was not triggered", duration)

	t.Logf("Successfully triggered fallback termination for shell %s after %v", session.ID, duration)
}
