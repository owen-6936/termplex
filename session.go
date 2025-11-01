package termplex

import (
	"fmt"
	"time"
)

type SessionManager struct {
	Sessions             map[string]*ShellSession
	Windows              map[string]*WindowManager
	MaxWindowsPerSession int
}

type Session struct {
	ID        string
	WindowIDs []string
	CreatedAt time.Time
	Tags      map[string]string
}

// NewSession creates a detached tmux session, replacing any existing one with the same name.
func NewSession(name string) error {
	// Kill existing session if it exists
	_, err := runTmux("has-session", "-t", name)
	if err == nil {
		_, _ = runTmux("kill-session", "-t", name)
	}

	_, err = runTmux("new-session", "-d", "-s", name)
	if err != nil {
		return fmt.Errorf("failed to create session %q: %w", name, err)
	}
	return nil
}

// KillSession terminates a tmux session.
func KillSession(name string) error {
	_, err := runTmux("kill-session", "-t", name)
	return err
}

// HasSession checks if a session exists.
func HasSession(name string) bool {
	_, err := runTmux("has-session", "-t", name)
	return err == nil
}
