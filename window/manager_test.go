package window_test

import (
	"testing"

	"github.com/owen-6936/termplex/window"
)

func TestWindowLifecycle(t *testing.T) {
	// 1. Create a new WindowManager.
	wm := window.NewWindowManager("test-window", nil)
	if wm.ID == "" {
		t.Fatal("NewWindowManager failed to assign an ID.")
	}

	// 2. Add a couple of panes to the window.
	paneID1, err := wm.AddPane("test-pane-1")
	if err != nil {
		t.Fatalf("Failed to add first pane: %v", err)
	}
	paneID2, err := wm.AddPane("test-pane-2")
	if err != nil {
		t.Fatalf("Failed to add second pane: %v", err)
	}

	// 3. Verify the panes were added correctly.
	if len(wm.Panes) != 2 {
		t.Fatalf("Expected 2 panes, but found %d", len(wm.Panes))
	}
	if _, exists := wm.GetPane(paneID1); !exists {
		t.Errorf("Pane %s was not found after being added.", paneID1)
	}
	if _, exists := wm.GetPane(paneID2); !exists {
		t.Errorf("Pane %s was not found after being added.", paneID2)
	}

	// 4. Spawn a shell in one of the panes to ensure termination cascades.
	pane1, _ := wm.GetPane(paneID1)
	_, err = pane1.SpawnShell(false, "bash", "-c", "sleep 5")
	if err != nil {
		t.Fatalf("Failed to spawn shell in pane: %v", err)
	}

	// 5. Terminate the window and verify cleanup.
	wm.TerminateWindow()
	if len(wm.Panes) != 0 {
		t.Errorf("Expected 0 panes after TerminateWindow, but found %d", len(wm.Panes))
	}
}
