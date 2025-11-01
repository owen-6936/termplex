package tmux

import (
	"fmt"
	"strings"
)

// Pane represents a single pane within a tmux session.
// It holds the necessary identifiers to target it with tmux commands.
type Pane struct {
	SessionName string
	WindowIndex int
	PaneIndex   int
}

// Target returns the string used by tmux to identify this specific pane.
// Example: "my-session:1.0"
func (p *Pane) Target() string {
	return fmt.Sprintf("%s:%d.%d", p.SessionName, p.WindowIndex, p.PaneIndex)
}

// SendKeys sends a command to the pane as if it were typed.
// It automatically appends a newline to execute the command.
func (p *Pane) SendKeys(command string) error {
	// The `C-m` is equivalent to pressing Enter.
	_, err := runTmux("send-keys", "-t", p.Target(), command, "C-m")
	return err
}

// Capture captures the visible text content of the pane.
func (p *Pane) Capture() (string, error) {
	// The -p flag prints the output to stdout.
	out, err := runTmux("capture-pane", "-p", "-t", p.Target())
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}
