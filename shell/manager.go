package shell

import (
	"errors"
	"fmt"
	"os/exec"
	"syscall"
	"time"

	"github.com/google/uuid"
)

// ShellManager controls shell lifecycle and enforces environment rules.
type ShellManager struct {
	EnvRegistry map[string]bool // Supported environments
	Shells      map[string]*ShellSession
}

// NewShellManager initializes a shell manager with known environments.
func NewShellManager(supportedEnvs []string) *ShellManager {
	registry := make(map[string]bool)
	for _, env := range supportedEnvs {
		registry[env] = true
	}
	return &ShellManager{
		EnvRegistry: registry,
		Shells:      make(map[string]*ShellSession),
	}
}

// IsEnvSupported checks if an environment is valid.
func (sm *ShellManager) IsEnvSupported(env string) bool {
	return sm.EnvRegistry[env]
}

// SpawnShell creates a new shell session.
func (sm *ShellManager) SpawnShell(interactive bool, command ...string) (*ShellSession, error) {
	if len(command) == 0 {
		return nil, errors.New("SpawnShell requires a command to execute")
	}
	cmd := exec.Command(command[0], command[1:]...)

	// Isolate the process in its own group to prevent interference with the test runner.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	shellID := uuid.New().String()
	newShell := &ShellSession{
		ID:          shellID,
		Cmd:         cmd,
		Stdin:       stdin,
		Stdout:      stdout,
		Stderr:      stderr,
		StartedAt:   time.Now(),
		Interactive: interactive,
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command %v: %w", command, err)
	}

	sm.Shells[shellID] = newShell
	fmt.Printf("🐚 Shell spawned: %s (%v)\n", shellID, command)
	return newShell, nil
}

// TerminateAllShells iterates through all managed shells and terminates them.
func (sm *ShellManager) TerminateAllShells() {
	if len(sm.Shells) == 0 {
		return
	}
	fmt.Printf("🧹 Terminating all %d active shells...\n", len(sm.Shells))
	// Note: We iterate over a copy of the keys because TerminateShell modifies the map.
	shellIDs := make([]string, 0, len(sm.Shells))
	for id := range sm.Shells {
		shellIDs = append(shellIDs, id)
	}
	for _, id := range shellIDs {
		_ = sm.TerminateShell(id)
	}
}

// SendCommand simulates sending a command to a shell.
func (sm *ShellManager) SendCommand(shellID, command string) (string, error) {
	_, exists := sm.Shells[shellID]
	if !exists {
		return "", errors.New("shell not found")
	}
	fmt.Printf("🔄 Command sent to shell %s: %s\n", shellID, command)
	return "Command acknowledged", nil
}

// TerminateShell removes a shell session.
func (sm *ShellManager) TerminateShell(shellID string) error {
	shell, exists := sm.Shells[shellID]
	if !exists {
		return errors.New("shell not found")
	}

	// Close the underlying process with a 2-second grace period.
	_ = shell.Close(2 * time.Second)

	delete(sm.Shells, shellID)
	fmt.Printf("🧹 Shell terminated: %s\n", shellID)
	return nil
}
