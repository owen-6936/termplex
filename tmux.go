package termplex

import (
	"bytes"
	"fmt"
	"os/exec"
)

// runTmux executes a tmux command and returns stdout or an error.
// It trims trailing newlines and preserves stderr for debugging.
func runTmux(args ...string) (string, error) {
    cmd := exec.Command("tmux", args...)

    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    err := cmd.Run()
    if err != nil {
        return "", fmt.Errorf("tmux error: %v\nstderr: %s", err, stderr.String())
    }

    return stdout.String(), nil
}
