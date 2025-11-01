package pane

import (
	"time"

	"github.com/owen-6936/termplex/shell"
)

// PaneManager represents a multitasking workspace within a window.
// It can host one interactive shell and multiple non-interactive shells.
type PaneManager struct {
	ID                   string
	CreatedAt            time.Time
	InteractiveShell     *shell.ShellSession
	NonInteractiveShells []*shell.ShellSession
	Tags                 map[string]string // Optional metadata (e.g. task, env, owner)
}
