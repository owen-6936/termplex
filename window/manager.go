package window

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/owen-6936/termplex/pane"
)

// NewWindowManager creates a new window with optional tags and name.
func NewWindowManager(name string, tags map[string]string) *WindowManager {
	return &WindowManager{
		ID:        uuid.New().String(),
		Name:      name,
		CreatedAt: time.Now(),
		Tags:      tags,
		Panes:     make(map[string]*PaneManager),
	}
}

// AddPane creates and registers a new pane in the window.
func (wm *WindowManager) AddPane() (string, error) {
	paneID := uuid.New().String()

	if _, exists := wm.Panes[paneID]; exists {
		return "", errors.New("pane ID collision")
	}

	wm.Panes[paneID] = pane.NewPaneManager(paneID)
	fmt.Printf("🪞 Pane created: %s in window %s\n", paneID, wm.ID)
	return paneID, nil
}

// GetPane retrieves a pane by ID.
func (wm *WindowManager) GetPane(paneID string) (*PaneManager, bool) {
	pane, exists := wm.Panes[paneID]
	return pane, exists
}

// TerminateWindow cleans up all panes in the window.
func (wm *WindowManager) TerminateWindow() {
	// Create a slice of pane IDs to iterate over, as deleting from a map
	// while iterating over it is not safe.
	paneIDs := make([]string, 0, len(wm.Panes))
	for id := range wm.Panes {
		paneIDs = append(paneIDs, id)
	}

	for _, paneID := range paneIDs {
		wm.Panes[paneID].TerminatePane(2 * time.Second)
		fmt.Printf("🧹 Pane terminated: %s in window %s\n", paneID, wm.ID)
		delete(wm.Panes, paneID)
	}
	fmt.Printf("🧹 Window terminated: %s\n", wm.ID)
}
