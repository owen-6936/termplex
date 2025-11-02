package shell

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/google/uuid"
)

// ShellManager controls shell lifecycle and enforces environment rules.
type ShellManager struct {
	mu         sync.Mutex
	Shells     map[string]*ShellSession
	OutputChan chan PaneOutput // A multiplexed stream of output from all managed shells.
	closeChan  chan struct{}
}

// NewShellManager initializes a shell manager with known environments.
func NewShellManager(supportedEnvs []string) *ShellManager {
	return &ShellManager{
		Shells:     make(map[string]*ShellSession),
		OutputChan: make(chan PaneOutput, 100),
		closeChan:  make(chan struct{}),
	}
}

// SpawnShell creates a new shell session.
func (sm *ShellManager) SpawnShell(interactive bool, command ...string) (*ShellSession, error) {
	if len(command) == 0 {
		return nil, errors.New("SpawnShell requires a command to execute")
	}
	cmd := exec.Command(command[0], command[1:]...)

	// For interactive shells, we MUST use a PTY to make the shell behave correctly.
	// For non-interactive, simple pipes are sufficient and more lightweight.
	var ptmx io.ReadWriteCloser
	var stderrPipe io.ReadCloser

	if interactive {
		// The PTY acts as both stdin and stdout for the shell process.
		var err error
		ptmx, err = pty.Start(cmd)
		if err != nil {
			return nil, fmt.Errorf("failed to start pty: %w", err)
		}
		stderrPipe = ptmx // In a PTY, stderr is merged with stdout.
	} else {
		// Use standard pipes for non-interactive shells.
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
		}
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
		}
		stderrPipe, err = cmd.StderrPipe()
		if err != nil {
			return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
		}

		if err := cmd.Start(); err != nil {
			return nil, fmt.Errorf("failed to start command %v: %w", command, err)
		}
		// For non-interactive shells, we use the raw pipes.
		ptmx = &pipeReadWriteCloser{r: stdout, w: stdin}
	}

	shellID := uuid.New().String()
	newShell := &ShellSession{
		ID:          shellID,
		Cmd:         cmd,
		Stdin:       ptmx,
		Stdout:      ptmx,
		Stderr:      stderrPipe,
		StartedAt:   time.Now(),
		Interactive: interactive,
	}

	sm.mu.Lock()
	sm.Shells[shellID] = newShell
	sm.mu.Unlock()

	// Define handlers that forward output to the ShellManager's own channel.
	stdoutHandler := func(output []byte) {
		// Also call the shell's default handler to populate its internal buffer.
		newShell.OutputHandler(output)

		select {
		case sm.OutputChan <- PaneOutput{ShellID: newShell.ID, Timestamp: time.Now(), Data: bytes.Clone(output)}:
		case <-sm.closeChan:
		}
	}

	stderrHandler := func(output []byte) {
		// Also call the shell's default handler to populate its internal buffer.
		newShell.ErrorOutputHandler(output)

		select {
		case sm.OutputChan <- PaneOutput{ShellID: newShell.ID, Timestamp: time.Now(), Data: bytes.Clone(output), IsStderr: true}:
		case <-sm.closeChan:
		}
	}

	// The ShellManager is now responsible for starting the I/O readers.
	// This happens immediately, preventing any race conditions.
	newShell.StartReading(stdoutHandler, stderrHandler)

	fmt.Printf("🐚 Shell spawned: %s (%v)\n", shellID, command)
	return newShell, nil
}

// TerminateAllShells iterates through all managed shells and terminates them.
func (sm *ShellManager) TerminateAllShells() {
	close(sm.closeChan)

	sm.mu.Lock()
	shellIDs := make([]string, 0, len(sm.Shells))
	for id := range sm.Shells {
		shellIDs = append(shellIDs, id)
	}
	sm.mu.Unlock()

	if len(shellIDs) == 0 {
		return
	}
	fmt.Printf("🧹 Terminating all %d active shells...\n", len(shellIDs))

	for _, id := range shellIDs {
		_ = sm.TerminateShell(id)
	}
}

// SendCommand simulates sending a command to a shell.
func (sm *ShellManager) SendCommand(shellID, command string) (string, error) {
	sm.mu.Lock()
	shell, exists := sm.Shells[shellID]
	sm.mu.Unlock()
	if !exists {
		return "", errors.New("shell not found")
	}
	err := shell.SendCommand(command)
	if err != nil {
		return "", fmt.Errorf("failed to send command: %w", err)
	}
	fmt.Printf("🔄 Command sent to shell %s: %s\n", shellID, command)
	return "Command acknowledged", nil
}

// TerminateShell removes a shell session.
func (sm *ShellManager) TerminateShell(shellID string) error {
	sm.mu.Lock()
	shell, exists := sm.Shells[shellID]
	sm.mu.Unlock()
	if !exists {
		return errors.New("shell not found")
	}

	// Close the underlying process with a 2-second grace period.
	_ = shell.Close(2 * time.Second)

	sm.mu.Lock()
	delete(sm.Shells, shellID)
	sm.mu.Unlock()
	fmt.Printf("🧹 Shell terminated: %s\n", shellID)
	return nil
}

// pipeReadWriteCloser is a helper to adapt separate Read and Write closers
// into a single io.ReadWriteCloser for non-PTY shells, simplifying the interface.
type pipeReadWriteCloser struct {
	r io.ReadCloser
	w io.WriteCloser
}

func (prwc *pipeReadWriteCloser) Read(p []byte) (n int, err error)  { return prwc.r.Read(p) }
func (prwc *pipeReadWriteCloser) Write(p []byte) (n int, err error) { return prwc.w.Write(p) }
func (prwc *pipeReadWriteCloser) Close() error {
	_ = prwc.r.Close()
	_ = prwc.w.Close()
	return nil
}
