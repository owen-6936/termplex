package pane

import (
	"time"

	"github.com/owen-6936/termplex/shell"
)

// PaneOutput represents a piece of output from a shell within a pane,
// providing context about its origin.
type PaneOutput struct {
	ShellID   string
	Timestamp time.Time
	Data      []byte
	IsStderr  bool
}

// PaneManager represents a multitasking workspace within a window.
// It can host one interactive shell and multiple non-interactive shells.
type PaneManager struct {
	ID                   string
	CreatedAt            time.Time
	InteractiveShell     *shell.ShellSession
	NonInteractiveShells []*shell.ShellSession
	Tags                 map[string]string // Optional metadata (e.g. task, env, owner)
	OutputChan           chan PaneOutput   // A multiplexed stream of output from all shells in this pane.
	closeChan            chan struct{}     // Signal to close the output channel and stop forwarding handlers.
}
