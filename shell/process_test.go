package shell_test

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

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
	if !strings.Contains(stdout, "hello stdout") {
		t.Errorf("Expected stdout buffer to contain 'hello stdout', but got: %q", stdout)
	}

	// Check if stderr was captured
	stderr := session.StderrBuf.String()
	if !strings.Contains(stderr, "hello stderr") {
		t.Errorf("Expected stderr buffer to contain 'hello stderr', but got: %q", stderr)
	}
}

func TestSendCommandAndWait(t *testing.T) {
	// Spawn a persistent interactive shell
	session := newTestShell(t, "bash", "-i")

	// Give the shell a moment to initialize
	time.Sleep(200 * time.Millisecond)

	// Send a command and wait for the response
	testPhrase := "this is a synchronous test"
	output, err := session.SendCommandAndWait("echo '" + testPhrase + "'")
	if err != nil {
		t.Fatalf("SendCommandAndWait failed: %v", err)
	}

	// The output should contain our phrase, trimmed of whitespace and shell prompts
	if !strings.Contains(output, testPhrase) {
		t.Errorf("Expected output to contain %q, but got: %q", testPhrase, output)
	}

	// Verify the output was also captured in the main buffer
	fullOutput := session.OutputBuf.String()
	if !strings.Contains(fullOutput, testPhrase) {
		t.Errorf("Expected main buffer to contain %q after SendCommandAndWait, but it didn't", testPhrase)
	}
}

func TestShellCloseWithFallback(t *testing.T) {
	// Spawn a shell that ignores the standard 'exit' signal by trapping SIGTERM
	// and then sleeps, forcing our fallback to trigger.
	command := []string{"bash", "-c", "trap '' TERM; echo 'ready'; sleep 5"}
	session := newTestShell(t, command...)

	// Wait for the "ready" signal to ensure the process is running
	time.Sleep(200 * time.Millisecond)
	if !strings.Contains(session.OutputBuf.String(), "ready") {
		t.Fatal("Shell did not print 'ready' signal")
	}

	// Attempt to close the session with a very short grace period.
	// This should fail gracefully and trigger the force-kill fallback.
	gracePeriod := 100 * time.Millisecond
	startTime := time.Now()
	err := session.Close(gracePeriod)
	duration := time.Since(startTime)

	// After a force-kill, Wait() returns an error. We expect this.
	// If err is nil, it means the process exited cleanly before the kill, which is wrong.
	if err == nil {
		t.Errorf("Expected an error from Wait() after a forced kill, but got nil")
	}

	// The close operation should have taken longer than the grace period because
	// it had to wait for the timeout before force-killing.
	if duration < gracePeriod {
		t.Errorf("The close operation finished too quickly (%v), suggesting the fallback was not triggered", duration)
	}

	t.Logf("Successfully triggered fallback termination for shell %s after %v", session.ID, duration)
}
