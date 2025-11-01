package shell

import (
	"time"
)

// ShellSession represents a shell process (bash, python, deepseek, etc.).
// It carries metadata for modality, environment, and lifecycle tracking.
type ShellSession struct {
	ID          string            // Unique shell ID
	Env         string            // Environment name (e.g. "bash", "deepseek-r1")
	Interactive bool              // Whether shell is interactive
	StartedAt   time.Time         // Timestamp of shell start
	Tags        map[string]string // Optional metadata (e.g. model, task, owner)
}
