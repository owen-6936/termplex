package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/owen-6936/termplex/session"
)

func main() {
	fmt.Println("🚀 Starting Termplex Architecture Demo...")
	var wg sync.WaitGroup

	// 1. Initialize the top-level SessionManager
	sm := session.NewSessionManager(5) // Allow up to 5 windows per session

	// 2. Create a new orchestration Session
	sessionID, err := sm.CreateSession("DemoProject", map[string]string{"owner": "owen"})
	if err != nil {
		panic(fmt.Sprintf("Failed to create session: %v", err))
	}

	// 3. Add a Window to the Session
	windowID, err := sm.AddWindow(sessionID, "MainWindow", nil)
	if err != nil {
		panic(fmt.Sprintf("Failed to add window: %v", err))
	}
	wm := sm.Windows[windowID]

	// 4. Add a Pane to the Window
	paneID, err := wm.AddPane()
	if err != nil {
		panic(fmt.Sprintf("Failed to add pane: %v", err))
	}
	pane, _ := wm.GetPane(paneID)

	// 5. Start a goroutine to consume all multiplexed output from the pane
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("\n--- 🎧 Listening for all output from Pane ---")
		for output := range pane.OutputChan {
			streamType := "STDOUT"
			if output.IsStderr {
				streamType = "STDERR"
			}
			fmt.Printf("[Pane Output | Shell: ...%s | %s]: %s", output.ShellID[len(output.ShellID)-6:], streamType, string(output.Data))
		}
		fmt.Println("--- 🛑 Pane output stream closed ---")
	}()

	// 6. Spawn shells within the Pane
	// Spawn an interactive shell
	interactiveShell, err := pane.SpawnShell(true, "bash", "-i")
	if err != nil {
		panic(fmt.Sprintf("Failed to spawn interactive shell: %v", err))
	}

	// Spawn a non-interactive background task
	_, err = pane.SpawnShell(false, "bash", "-c", "echo 'Background task starting...'; sleep 1; echo 'Background task finished.'")
	if err != nil {
		panic(fmt.Sprintf("Failed to spawn background shell: %v", err))
	}

	// 7. Interact with the interactive shell specifically
	fmt.Println("\n--- ▶️ Sending command to interactive shell ---")
	output, err := interactiveShell.SendCommandAndWait("echo 'Hello from the interactive shell!'")
	if err != nil {
		fmt.Printf("Error sending command: %v\n", err)
	}
	fmt.Printf("--- ⏹️  Received specific response: %s\n", output)

	// Let background tasks run for a bit
	time.Sleep(2 * time.Second)

	// 8. Terminate the entire session, which will clean up everything
	fmt.Println("\n--- 🧹 Terminating entire session ---")
	sm.TerminateSession(sessionID)

	// Wait for the output consumer goroutine to finish
	wg.Wait()
	fmt.Println("\n✅ Demo finished successfully.")
}
