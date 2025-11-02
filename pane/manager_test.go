package pane_test

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/owen-6936/termplex/pane"
	"github.com/owen-6936/termplex/window"
)

func TestPaneOutputMultiplexing(t *testing.T) {
	// 1. Initialize a PaneManager.
	pm := pane.NewPaneManager("test-mux-pane", "multiplexer")
	// Ensure the pane and its resources are cleaned up when the test finishes.
	t.Cleanup(func() { pm.TerminatePane(2 * time.Second) })

	const numShells = 5
	// We expect two outputs (stdout, stderr) from each shell.
	expectedMessageCount := numShells * 2

	// A map to track the unique outputs we expect to receive.
	expectedOutputs := make(map[string]bool) // A map is not a slice, so it remains.
	for i := range numShells {
		expectedOutputs[fmt.Sprintf("stdout-from-shell-%d", i)] = true
		expectedOutputs[fmt.Sprintf("stderr-from-shell-%d", i)] = true
	}

	// 2. Spawn multiple shells concurrently.
	var wg sync.WaitGroup
	for i := range numShells {
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
	receivedOutputs := make(map[string]bool) // A map is not a slice, so it remains.
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
	pm := pane.NewPaneManager("test-interactive-rule-pane", "interactive-tester")
	t.Cleanup(func() { pm.TerminatePane(2 * time.Second) })

	// 2. Successfully spawn the first interactive shell.
	firstShell, err := pm.SpawnShell(true, "bash", "-i")
	if err != nil {
		t.Fatalf("Failed to spawn the first interactive shell: %v", err)
	}
	if pm.InteractiveShell == nil || pm.InteractiveShell.ID != firstShell.ID {
		t.Fatal("PaneManager did not correctly register the first interactive shell.")
	}

	// by gracefully replacing the first one.
	secondShell, err := pm.SpawnShell(true, "bash", "-i", "-c", "echo 'new shell'")

	// 4. Verify the failure.
	if err != nil {
		t.Fatalf("Expected second interactive shell to spawn successfully, but got error: %v", err)
	}
	if secondShell == nil {
		t.Fatal("Expected a new shell object, but got nil.")
	}

	// 5. Verify that the new shell is now the active interactive shell.
	if pm.InteractiveShell.ID == firstShell.ID {
		t.Error("PaneManager did not replace the old interactive shell with the new one.")
	}
	if pm.InteractiveShell.ID != secondShell.ID {
		t.Errorf("Expected interactive shell to be %s, but got %s", secondShell.ID, pm.InteractiveShell.ID)
	}
}

func TestPaneOutputMetadata(t *testing.T) {
	// This test verifies that the metadata (ShellID, IsStderr) on the
	// PaneOutput struct is accurate when multiplexing output.

	// 1. Initialize a PaneManager.
	pm := pane.NewPaneManager("test-metadata-pane", "metadata-tester")
	t.Cleanup(func() { pm.TerminatePane(2 * time.Second) })

	// 2. Spawn two distinct shells.
	shell1, err := pm.SpawnShell(false, "bash", "-c", "echo 'out 1'; >&2 echo 'err 1'")
	if err != nil {
		t.Fatalf("Failed to spawn shell 1: %v", err)
	}

	shell2, err := pm.SpawnShell(false, "bash", "-c", "echo 'out 2'; >&2 echo 'err 2'")
	if err != nil {
		t.Fatalf("Failed to spawn shell 2: %v", err)
	}

	// 3. Consume the four expected messages from the output channel.
	expectedMessages := 4
	receivedMessages := 0
	timeout := time.After(5 * time.Second)

	// Use maps to track received outputs for each shell and stream.
	received := make(map[string]map[string]bool)
	received[shell1.ID] = make(map[string]bool)
	received[shell2.ID] = make(map[string]bool)

	for receivedMessages < expectedMessages {
		select {
		case output, ok := <-pm.OutputChan:
			if !ok {
				t.Fatal("OutputChan was closed prematurely")
			}
			data := strings.TrimSpace(string(output.Data))

			// Mark the received message based on its metadata.
			if output.IsStderr {
				received[output.ShellID]["stderr"] = (data == "err 1" || data == "err 2")
			} else {
				received[output.ShellID]["stdout"] = (data == "out 1" || data == "out 2")
			}
			receivedMessages++

		case <-timeout:
			t.Fatalf("Timed out waiting for shell outputs. Received %d of %d messages.", receivedMessages, expectedMessages)
		}
	}

	// 4. Assert that we received one stdout and one stderr from each shell.
	if !(received[shell1.ID]["stdout"] && received[shell1.ID]["stderr"]) {
		t.Errorf("Did not receive both stdout and stderr for shell 1. Got: %v", received[shell1.ID])
	}
	if !(received[shell2.ID]["stdout"] && received[shell2.ID]["stderr"]) {
		t.Errorf("Did not receive both stdout and stderr for shell 2. Got: %v", received[shell2.ID])
	}

}

func TestInteractiveShellOutput(t *testing.T) {
	// This is a minimal test to isolate the behavior of a single interactive shell.
	// If this fails, the problem is in the core I/O handling for interactive shells.

	// 1. Create a single pane.
	pm := pane.NewPaneManager("test-interactive-pane", "interactive-output-tester")
	t.Cleanup(func() { pm.TerminatePane(2 * time.Second) })

	// 2. Spawn one interactive shell.
	shell, err := pm.SpawnShell(true, "bash", "-i")
	if err != nil {
		t.Fatalf("Failed to spawn interactive shell: %v", err)
	}
	time.Sleep(200 * time.Millisecond) // Allow shell to initialize.

	// 3. Send a simple command.
	testPhrase := "testing interactive output"
	if err := pm.SendCommand(shell.ID, "echo '"+testPhrase+"'"); err != nil {
		t.Fatalf("Failed to send command: %v", err)
	}

	// 4. Listen on the output channel, accumulating data.
	timeout := time.After(3 * time.Second)
	var receivedOutput strings.Builder
	for {
		select {
		case output := <-pm.OutputChan:
			receivedOutput.Write(output.Data)
			if strings.Contains(receivedOutput.String(), testPhrase) {
				return // Success!
			}
		case <-timeout:
			t.Fatalf("Timed out waiting for interactive shell output. Accumulated: %q", receivedOutput.String())
		}
	}
}

func TestSendCommandToShellInPane(t *testing.T) {
	// This test simulates a more realistic scenario where we look up a pane
	// by name and then send a command to a shell running within it.

	// 1. Setup a mock window manager to hold our pane.
	wm := window.NewWindowManager("Api Server", nil)

	// 2. Create a pane with a specific name.
	paneName := "worker-pane"
	pmId, err := wm.AddPane(paneName)
	if err != nil {
		t.Fatalf("Failed to add pane: %v", err)
	}
	pm, exists := wm.GetPane(pmId)
	if !exists {
		t.Fatalf("Failed to find pane with ID %s", pmId)
	}
	t.Cleanup(func() { pm.TerminatePane(2 * time.Second) })

	// 3. Spawn an interactive shell inside the pane.
	shell, err := pm.SpawnShell(true, "bash", "-i")
	if err != nil {
		t.Fatalf("Failed to spawn shell: %v", err)
	}
	time.Sleep(200 * time.Millisecond) // Allow shell to initialize.

	// 4. Find the pane by its name and send a command to the shell.
	testPhrase := "hello from named pane"
	if err := pm.SendCommand(shell.ID, "echo '"+testPhrase+"'"); err != nil {
		t.Fatalf("Failed to send command to shell: %v", err)
	}

	// 5. Assert that the command's output was received on the pane's public OutputChan.
	// We must accumulate output because an interactive shell might send the
	// command echo and the command result in separate chunks.
	timeout := time.After(2 * time.Second)
	var receivedOutput strings.Builder
	for {
		select {
		case output, ok := <-pm.OutputChan:
			if !ok {
				t.Fatal("OutputChan was closed prematurely")
			}
			receivedOutput.Write(output.Data)
			if strings.Contains(receivedOutput.String(), testPhrase) {
				// Success! The expected output was found.
				return
			}
		case <-timeout:
			t.Fatalf("Timed out waiting for command output. Accumulated output: %q", receivedOutput.String())
		}
	}
}
