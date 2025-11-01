package tmux

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/owen-6936/termplex/debug"
)

// runTmux executes a tmux command and returns stdout or an error.
// It trims trailing newlines and preserves stderr for debugging.
func runTmux(args ...string) (string, error) {
	cmd := exec.Command("tmux", args...)
	debug.Log("Executing command: tmux %s", strings.Join(args, " "))

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		if stderr.Len() > 0 {
			debug.Log("Command failed with stderr: %s", stderr.String())
		}
		return "", fmt.Errorf("tmux error: %v\nstderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}
