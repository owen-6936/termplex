package tmux

import (
	"fmt"
	"strconv"
	"strings"
)

// SessionManager manages the lifecycle of a real tmux session and its windows/panes.
type SessionManager struct {
	SessionName string
	Panes       []*Pane
}

// NewSessionManager creates a new, detached tmux session.
func NewSessionManager(sessionName string) (*SessionManager, error) {
	// Create a detached session with a single window.
	_, err := runTmux("new-session", "-d", "-s", sessionName)
	if err != nil {
		// It's okay if the session already exists.
		if !strings.Contains(err.Error(), "duplicate session") {
			return nil, fmt.Errorf("failed to create tmux session: %w", err)
		}
	}

	// The first pane is always at window 0, pane 0.
	initialPane := &Pane{
		SessionName: sessionName,
		WindowIndex: 0,
		PaneIndex:   0,
	}

	return &SessionManager{
		SessionName: sessionName,
		Panes:       []*Pane{initialPane},
	}, nil
}

// AddPane splits the last created pane to create a new one.
func (sm *SessionManager) AddPane() (*Pane, error) {
	if len(sm.Panes) == 0 {
		return nil, fmt.Errorf("no panes exist to split")
	}
	// Split the window of the most recently added pane.
	lastPane := sm.Panes[len(sm.Panes)-1]
	targetWindow := fmt.Sprintf("%s:%d", lastPane.SessionName, lastPane.WindowIndex)

	// Create a new pane and get its index.
	// The -P flag prints the new pane's index.
	out, err := runTmux("split-window", "-t", targetWindow, "-P", "-F", "#{pane_index}")
	if err != nil {
		return nil, fmt.Errorf("failed to split window: %w", err)
	}

	paneIndex, err := strconv.Atoi(strings.TrimSpace(out))
	if err != nil {
		return nil, fmt.Errorf("failed to parse new pane index: %w", err)
	}

	newPane := &Pane{
		SessionName: sm.SessionName,
		WindowIndex: lastPane.WindowIndex, // For simplicity, stay in the same window.
		PaneIndex:   paneIndex,
	}
	sm.Panes = append(sm.Panes, newPane)
	return newPane, nil
}

// KillSession destroys the entire tmux session.
func (sm *SessionManager) KillSession() error {
	_, err := runTmux("kill-session", "-t", sm.SessionName)
	return err
}
