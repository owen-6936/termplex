package pane

import (
	"fmt"
	"sync"
	"time"

	"github.com/owen-6936/termplex/shell"
)

// NewPaneManager initializes a new pane with a unique ID.
func NewPaneManager(paneID, name string) *PaneManager {
	pm := &PaneManager{
		ID:        paneID,
		Name:      name,
		CreatedAt: time.Now(),
		// Each pane gets its own dedicated shell manager.
		Shells:     shell.NewShellManager(nil),
		Tags:       make(map[string]string),
		OutputChan: make(chan PaneOutput, 100), // Buffered channel
		closeChan:  make(chan struct{}),
	}
	pm.tagsCond = sync.NewCond(&pm.tagsMu)
	// Start a single goroutine to forward all output from the shell manager.
	go pm.forwardShellOutput()
	return pm
}

// SpawnShell creates and registers a new shell process within the pane.
func (pm *PaneManager) SpawnShell(interactive bool, command ...string) (*shell.ShellSession, error) {
	if interactive {
		// If an interactive shell already exists, gracefully terminate it before spawning the new one.
		if pm.InteractiveShell != nil {
			fmt.Printf("🔄 Replacing existing interactive shell %s\n", pm.InteractiveShell.ID)
			// Use a 5-second grace period as suggested.
			_, _ = pm.TerminateShell(pm.InteractiveShell.ID, 5*time.Second)
		}
	}

	// Delegate shell creation to the pane's own shell manager.
	newShell, err := pm.Shells.SpawnShell(interactive, command...)
	if err != nil {
		return nil, fmt.Errorf("failed to spawn shell via manager: %w", err)
	}

	if interactive {
		pm.InteractiveShell = newShell
	}
	// The shell manager already prints a spawn message.

	return newShell, nil
}

// forwardShellOutput listens to the dedicated shell manager's output channel
// and forwards it to the pane's own channel. This is the new, clean abstraction.
func (pm *PaneManager) forwardShellOutput() {
	for {
		select {
		case output, ok := <-pm.Shells.OutputChan:
			if !ok {
				return // ShellManager's channel was closed.
			}
			// Convert shell.PaneOutput to pane.PaneOutput
			pm.OutputChan <- PaneOutput{
				ShellID:   output.ShellID,
				Timestamp: output.Timestamp,
				Data:      output.Data,
				IsStderr:  output.IsStderr}
		case <-pm.closeChan:
			return
		}
	}
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

// SendCommand delegates to the pane's shell manager to send a command to a specific shell.
func (pm *PaneManager) SendCommand(shellID, command string) error {
	_, err := pm.Shells.SendCommand(shellID, command)
	return err
}

// TerminateShell attempts a graceful shutdown of a specific shell session.
func (pm *PaneManager) TerminateShell(shellID string, gracePeriod time.Duration) (bool, error) {
	// Delegate termination to the pane's shell manager.
	// Also, check if the shell being terminated is the active interactive one.
	if pm.InteractiveShell != nil && pm.InteractiveShell.ID == shellID {
		pm.InteractiveShell = nil
	}

	return true, pm.Shells.TerminateShell(shellID)
}

// TerminatePane cleans up all shells in the pane by gracefully shutting them down.
func (pm *PaneManager) TerminatePane(gracePeriod time.Duration) {
	// Signal to all forwarding handlers that they should stop sending to OutputChan.
	close(pm.closeChan)
	pm.Shells.TerminateAllShells()

	// Close the main output channel to signal the end of the stream.
	close(pm.OutputChan)
}
