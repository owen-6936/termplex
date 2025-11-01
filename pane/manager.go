package pane

import (
	"errors"
	"fmt"
	"os/exec"
	"time"

	"github.com/owen-6936/termplex/shell"
)

// NewPaneManager initializes a new pane with a unique ID.
func NewPaneManager(paneID string) *PaneManager {
	return &PaneManager{
		ID:                   paneID,
		CreatedAt:            time.Now(),
		Tags:                 make(map[string]string),
		NonInteractiveShells: []*shell.ShellSession{},
	}
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

	// Automatically start reading from the shell's stdout and stderr pipes.
	newShell.StartReading(newShell.OutputHandler, newShell.ErrorOutputHandler)

	if interactive {
		pm.InteractiveShell = newShell
		fmt.Printf("🐚 Interactive shell spawned: %s in pane %s\n", newShell.ID, pm.ID)
	} else {
		pm.NonInteractiveShells = append(pm.NonInteractiveShells, newShell)
		fmt.Printf("🔧 Non-interactive shell spawned: %s in pane %s\n", newShell.ID, pm.ID)
	}

	return newShell, nil
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
	if pm.InteractiveShell != nil {
		_, _ = pm.TerminateShell(pm.InteractiveShell.ID, gracePeriod)
		pm.InteractiveShell = nil
	}
	for _, s := range pm.NonInteractiveShells {
		_, _ = pm.TerminateShell(s.ID, gracePeriod)
	}
	pm.NonInteractiveShells = []*shell.ShellSession{}
}
