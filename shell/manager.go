package shell

import (
	"errors"
	"fmt"
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
func (sm *ShellManager) SpawnShell(env string, interactive bool, tags map[string]string) (*ShellSession, error) {
	if !sm.IsEnvSupported(env) {
		return nil, errors.New("unsupported environment: " + env)
	}

	shellID := uuid.New().String()
	newShell := &ShellSession{
		ID:          shellID,
		Env:         env,
		Interactive: interactive,
		StartedAt:   time.Now(),
		Tags:        tags,
	}

	sm.Shells[shellID] = newShell
	fmt.Printf("🐚 Shell spawned: %s (%s)\n", shellID, env)
	return newShell, nil
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
	if _, exists := sm.Shells[shellID]; !exists {
		return errors.New("shell not found")
	}
	delete(sm.Shells, shellID)
	fmt.Printf("🧹 Shell terminated: %s\n", shellID)
	return nil
}
