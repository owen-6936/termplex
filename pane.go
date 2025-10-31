package termplex

import (
	"fmt"
	"strings"
)

// Pane represents a tmux pane within a window.
type Pane struct {
    SessionName string
    WindowName  string
    PaneIndex   int
}

// NewPane creates a new pane in the given window and returns its metadata.
// It assumes the new pane is created by splitting the active one.
func NewPane(sessionName, windowName string) (*Pane, error) {
    // Split the active pane in the window
    target := fmt.Sprintf("%s:%s", sessionName, windowName)
    _, err := runTmux("split-window", "-t", target, "-P")
    if err != nil {
        return nil, fmt.Errorf("failed to split pane: %w", err)
    }

    // List panes to find the highest index (assumes new pane is last)
    out, err := runTmux("list-panes", "-t", target, "-F", "#{pane_index}")
    if err != nil {
        return nil, fmt.Errorf("failed to list panes: %w", err)
    }

    // Parse pane indices and find the max
    lines := strings.Split(strings.TrimSpace(out), "\n")
    maxIndex := 0
    for _, line := range lines {
        var idx int
        fmt.Sscanf(line, "%d", &idx)
        if idx > maxIndex {
            maxIndex = idx
        }
    }

    return &Pane{
        SessionName: sessionName,
        WindowName:  windowName,
        PaneIndex:   maxIndex,
    }, nil
}

// Target returns the tmux target string for this pane.
func (p *Pane) Target() string {
    return fmt.Sprintf("%s:%s.%d", p.SessionName, p.WindowName, p.PaneIndex)
}

// StartShell seeds a shell in the pane at the given path.
func (p *Pane) StartShell(path string) error {
    cmd := fmt.Sprintf("cd %s && exec bash", path)
    return p.SendKeys(cmd)
}

// SendKeys sends a command to the pane.
func (p *Pane) SendKeys(cmd string) error {
    args := []string{"send-keys", "-t", p.Target(), cmd, "Enter"}
    _, err := runTmux(args...)
    return err
}

// GetCurrentPath returns the working directory of the pane.
func (p *Pane) GetCurrentPath() (string, error) {
    args := []string{"display-message", "-p", "-F", "#{pane_current_path}", "-t", p.Target()}
    out, err := runTmux(args...)
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(out), nil
}

// GetCurrentCommand returns the active command running in the pane.
func (p *Pane) GetCurrentCommand() (string, error) {
    args := []string{"display-message", "-p", "-F", "#{pane_current_command}", "-t", p.Target()}
    out, err := runTmux(args...)
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(out), nil
}

// CaptureBuffer returns the visible buffer of the pane.
func (p *Pane) CaptureBuffer() (string, error) {
    args := []string{"capture-pane", "-p", "-t", p.Target()}
    out, err := runTmux(args...)
    if err != nil {
        return "", err
    }
    return out, nil
}
