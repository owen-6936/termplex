package pane

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
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

// SpawnShell adds a shell to the pane, enforcing modality rules.
func (pm *PaneManager) SpawnShell(env string, interactive bool) (*shell.ShellSession, error) {
	if interactive && !pm.CanSpawnInteractive() {
		return nil, errors.New("interactive shell already active in this pane")
	}

	shellID := uuid.New().String()
	newShell := &shell.ShellSession{
		ID:          shellID,
		Env:         env,
		Interactive: interactive,
		StartedAt:   time.Now(),
	}

	if interactive {
		pm.InteractiveShell = newShell
		fmt.Printf("🐚 Interactive shell spawned: %s in pane %s\n", shellID, pm.ID)
	} else {
		pm.NonInteractiveShells = append(pm.NonInteractiveShells, newShell)
		fmt.Printf("🔧 Non-interactive shell spawned: %s in pane %s\n", shellID, pm.ID)
	}

	return newShell, nil
}

// TerminatePane cleans up all shells in the pane.
func (pm *PaneManager) TerminatePane() {
	if pm.InteractiveShell != nil {
		fmt.Printf("🧹 Terminating interactive shell: %s\n", pm.InteractiveShell.ID)
		pm.InteractiveShell = nil
	}
	for _, s := range pm.NonInteractiveShells {
		fmt.Printf("🧹 Terminating non-interactive shell: %s\n", s.ID)
	}
	pm.NonInteractiveShells = []*shell.ShellSession{}
}
