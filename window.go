package termplex

import (
	"fmt"
	"strings"
)

type WindowManager struct {
	Windows map[string]*Window
}

// Window represents a tmux window within a session.
type Window struct {
	SessionName string
	WindowIndex int
	WindowName  string
}

// NewWindow creates a new window in the given session.
// If name is empty, tmux assigns a default name.
func NewWindow(sessionName string, name string) (*Window, error) {
	args := []string{"new-window", "-t", sessionName, "-P", "-F", "#{window_name}"}
	if name != "" {
		args = append(args, "-n", name)
	}

	out, err := runTmux(args...)
	if err != nil {
		return nil, fmt.Errorf("failed to create window: %w", err)
	}

	// tmux returns the window name or index depending on flags
	windowName := strings.TrimSpace(out)

	return &Window{
		SessionName: sessionName,
		WindowName:  windowName,
		WindowIndex: 0, // optional: parse from output if needed
	}, nil
}

// SendKeys sends a command to the active pane in this window.
func (w *Window) SendKeys(cmd string) error {
	target := fmt.Sprintf("%s:%s", w.SessionName, w.WindowName)
	args := []string{"send-keys", "-t", target, cmd, "Enter"}
	_, err := runTmux(args...)
	return err
}

// Select activates this window in the session.
func (w *Window) Select() error {
	target := fmt.Sprintf("%s:%s", w.SessionName, w.WindowName)
	args := []string{"select-window", "-t", target}
	_, err := runTmux(args...)
	return err
}
