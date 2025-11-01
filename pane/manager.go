package pane

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/owen-6936/termplex/shell"
)

// NewPaneManager initializes a new pane with a unique ID.
func NewPaneManager(paneID string) *PaneManager {
	pm := &PaneManager{
		ID:                   paneID,
		CreatedAt:            time.Now(),
		Tags:                 make(map[string]string), // A map is not a slice, so it remains.
		NonInteractiveShells: nil,
		OutputChan:           make(chan PaneOutput, 100), // Buffered channel
		closeChan:            make(chan struct{}),
	}
	pm.tagsCond = sync.NewCond(&pm.tagsMu)
	return pm
}

// CanSpawnInteractive checks if the pane is free for an interactive shell.
func (pm *PaneManager) CanSpawnInteractive() bool {
	return pm.InteractiveShell == nil
}

// SpawnShell creates and registers a new shell process within the pane.
func (pm *PaneManager) SpawnShell(interactive bool, command ...string) (*shell.ShellSession, error) {
	if interactive && !pm.CanSpawnInteractive() {
		return nil, errors.New("interactive shell already active in this pane")
	}

	if len(command) == 0 {
		return nil, errors.New("SpawnShell requires a command to execute (e.g., \"bash\", \"-i\")")
	}
	cmd := exec.Command(command[0], command[1:]...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		// If starting the command fails, ensure we close the pipes to prevent resource leaks.
		_ = stdin.Close()
		_ = stdout.Close()
		_ = stderr.Close()
		return nil, fmt.Errorf("failed to start command %v: %w", command, err)
	}

	// The shell session now holds the real process details
	newShell := &shell.ShellSession{
		ID:          fmt.Sprintf("pane-%s-shell-%d", pm.ID, time.Now().UnixNano()),
		Cmd:         cmd,
		Stdin:       stdin,
		Stdout:      stdout,
		Stderr:      stderr,
		StartedAt:   time.Now(),
		Interactive: interactive,
	}

	// Define custom handlers that forward output to the pane's multiplexed channel.
	stdoutHandler := func(output []byte) {
		// First, call the default handler to buffer output locally on the shell.
		newShell.OutputHandler(output)

		// Then, send the output to the pane's combined stream.
		select {
		case pm.OutputChan <- PaneOutput{ShellID: newShell.ID, Timestamp: time.Now(), Data: bytes.Clone(output)}:
		case <-pm.closeChan: // Avoid blocking if the pane is terminated.
		}
	}

	stderrHandler := func(output []byte) {
		// Call the default handler to buffer stderr locally.
		newShell.ErrorOutputHandler(output)

		// Send to the combined stream, marked as stderr.
		select {
		case pm.OutputChan <- PaneOutput{ShellID: newShell.ID, Timestamp: time.Now(), Data: bytes.Clone(output), IsStderr: true}:
		case <-pm.closeChan:
		}
	}

	// Start reading from the shell's pipes using our custom handlers.
	newShell.StartReading(stdoutHandler, stderrHandler)

	if interactive {
		pm.InteractiveShell = newShell
		fmt.Printf("🐚 Interactive shell spawned: %s in pane %s\n", newShell.ID, pm.ID)
	} else {
		pm.NonInteractiveShells = append(pm.NonInteractiveShells, newShell)
		fmt.Printf("🔧 Non-interactive shell spawned: %s in pane %s\n", newShell.ID, pm.ID)
	}

	return newShell, nil
}

// AddTag safely adds or updates a tag on the pane and notifies any waiting listeners.
func (pm *PaneManager) AddTag(key, value string) {
	pm.tagsMu.Lock()
	defer pm.tagsMu.Unlock()

	fmt.Printf("MILESTONE: Pane %s tagged '%s' = '%s'\n", pm.ID, key, value)
	pm.Tags[key] = value

	// Wake up all goroutines waiting on a tag change.
	pm.tagsCond.Broadcast()
}

// WaitForTag blocks until a specific tag has a specific value, or until the timeout is reached.
func (pm *PaneManager) WaitForTag(key, value string, timeout time.Duration) error {
	pm.tagsMu.Lock()
	defer pm.tagsMu.Unlock()

	// Channel to signal when the condition is met or timed out.
	done := make(chan struct{})

	go func() {
		// This loop is the core of the waiting logic.
		for pm.Tags[key] != value {
			// cond.Wait() atomically unlocks the mutex and waits for a signal.
			// When woken up, it re-locks the mutex before proceeding.
			pm.tagsCond.Wait()
		}
		// The condition is met, close the channel to unblock the select.
		close(done)
	}()

	select {
	case <-done:
		// The tag was found.
		return nil
	case <-time.After(timeout):
		// The timeout was reached.
		return fmt.Errorf("timed out waiting for tag '%s' = '%s' on pane %s", key, value, pm.ID)
	}
}

// TerminateShell attempts a graceful shutdown of a specific shell session.
func (pm *PaneManager) TerminateShell(shellID string, gracePeriod time.Duration) (bool, error) {
	if pm.InteractiveShell != nil && pm.InteractiveShell.ID == shellID {
		fmt.Printf("🧹 Terminating interactive shell: %s\n", pm.InteractiveShell.ID)
		err := pm.InteractiveShell.Close(gracePeriod)
		pm.InteractiveShell = nil
		return true, err
	}

	for i, s := range pm.NonInteractiveShells {
		if s.ID == shellID {
			fmt.Printf("🧹 Terminating non-interactive shell: %s\n", s.ID)
			err := s.Close(gracePeriod)
			// Remove the shell from the slice
			pm.NonInteractiveShells = append(pm.NonInteractiveShells[:i], pm.NonInteractiveShells[i+1:]...)
			return true, err
		}
	}

	return false, fmt.Errorf("shell %s not found in pane %s", shellID, pm.ID)
}

// TerminatePane cleans up all shells in the pane by gracefully shutting them down.
func (pm *PaneManager) TerminatePane(gracePeriod time.Duration) {
	// Signal to all forwarding handlers that they should stop sending to OutputChan.
	close(pm.closeChan)

	// Terminate the interactive shell if it exists.
	if pm.InteractiveShell != nil {
		_, _ = pm.TerminateShell(pm.InteractiveShell.ID, gracePeriod)
	}

	// Terminate non-interactive shells by iterating backward to safely remove elements.
	for i := len(pm.NonInteractiveShells) - 1; i >= 0; i-- {
		shell := pm.NonInteractiveShells[i]
		_, _ = pm.TerminateShell(shell.ID, gracePeriod)
	}

	// Close the main output channel to signal the end of the stream.
	close(pm.OutputChan)
}
