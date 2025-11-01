package pane_test

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/owen-6936/termplex/pane"
)

func TestPaneOutputMultiplexing(t *testing.T) {
	// 1. Initialize a PaneManager.
	pm := pane.NewPaneManager("test-mux-pane")
	// Ensure the pane and its resources are cleaned up when the test finishes.
	t.Cleanup(func() { pm.TerminatePane(2 * time.Second) })

	const numShells = 5
	// We expect two outputs (stdout, stderr) from each shell.
	expectedMessageCount := numShells * 2

	// A map to track the unique outputs we expect to receive.
	expectedOutputs := make(map[string]bool)
	for i := 0; i < numShells; i++ {
		expectedOutputs[fmt.Sprintf("stdout-from-shell-%d", i)] = true
		expectedOutputs[fmt.Sprintf("stderr-from-shell-%d", i)] = true
	}

	// 2. Spawn multiple shells concurrently.
	var wg sync.WaitGroup
	for i := 0; i < numShells; i++ {
		wg.Add(1)
		go func(shellIndex int) {
			defer wg.Done()
			stdoutMsg := fmt.Sprintf("stdout-from-shell-%d", shellIndex)
			stderrMsg := fmt.Sprintf("stderr-from-shell-%d", shellIndex)
			command := fmt.Sprintf("echo '%s'; >&2 echo '%s'", stdoutMsg, stderrMsg)

			_, err := pm.SpawnShell(false, "bash", "-c", command)
			if err != nil {
				// t.Errorf is thread-safe and can be called from goroutines.
				t.Errorf("Failed to spawn shell %d: %v", shellIndex, err)
			}
		}(i)
	}
	wg.Wait() // Wait for all spawning goroutines to complete.

	// 3. Concurrently, consume messages from the multiplexed output channel.
	receivedCount := 0
	receivedOutputs := make(map[string]bool)
	timeout := time.After(5 * time.Second)

	for receivedCount < expectedMessageCount {
		select {
		case output, ok := <-pm.OutputChan:
			if !ok {
				t.Fatal("OutputChan was closed prematurely")
			}
			// Trim whitespace from the received data to match our expected keys.
			data := strings.TrimSpace(string(output.Data))
			if _, exists := expectedOutputs[data]; exists {
				if !receivedOutputs[data] {
					receivedOutputs[data] = true
					receivedCount++
				}
			}
		case <-timeout:
			t.Fatalf("Timed out waiting for shell outputs. Received %d of %d messages.", receivedCount, expectedMessageCount)
		}
	}

	// 4. Final verification.
	if len(receivedOutputs) != expectedMessageCount {
		t.Errorf("Expected to receive %d unique messages, but got %d", expectedMessageCount, len(receivedOutputs))
		for expected := range expectedOutputs {
			if !receivedOutputs[expected] {
				t.Logf("Missing expected output: %s", expected)
			}
		}
	}
}

func TestPaneManagerSingleInteractiveShellRule(t *testing.T) {
	// 1. Initialize a PaneManager.
	pm := pane.NewPaneManager("test-interactive-rule-pane")
	t.Cleanup(func() { pm.TerminatePane(2 * time.Second) })

	// 2. Successfully spawn the first interactive shell.
	firstShell, err := pm.SpawnShell(true, "bash", "-i")
	if err != nil {
		t.Fatalf("Failed to spawn the first interactive shell: %v", err)
	}
	if pm.InteractiveShell == nil || pm.InteractiveShell.ID != firstShell.ID {
		t.Fatal("PaneManager did not correctly register the first interactive shell.")
	}

	// 3. Attempt to spawn a second interactive shell. This should fail.
	secondShell, err := pm.SpawnShell(true, "bash", "-i")

	// 4. Verify the failure.
	if err == nil {
		t.Fatal("Expected an error when spawning a second interactive shell, but got nil.")
	}
	if secondShell != nil {
		t.Error("A second shell object was returned even though spawning failed.")
	}

	// 5. Verify that the original interactive shell is still in place.
	if pm.InteractiveShell.ID != firstShell.ID {
		t.Error("The original interactive shell was replaced or removed after a failed spawn attempt.")
	}
}
