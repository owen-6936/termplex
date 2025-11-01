package window

import (
	"time"

	"github.com/owen-6936/termplex/pane"
)

type PaneManager = pane.PaneManager

// WindowManager represents a logical project or domain boundary.
// It owns panes, tracks metadata, and supports contributor tagging.
type WindowManager struct {
	ID        string                  // Unique window ID
	Name      string                  // Optional human-readable name (e.g. "LLM Window")
	CreatedAt time.Time               // Timestamp of window creation
	Tags      map[string]string       // Metadata (e.g. project, owner, type)
	Panes     map[string]*PaneManager // Map of pane IDs to their managers
}
