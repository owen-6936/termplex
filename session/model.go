package session

import "time"

// ShellSession represents a top-level orchestration unit.
// It owns windows, tracks creation metadata, and supports tagging for contributor clarity.
type ShellSession struct {
	ID         string            // Unique session ID
	Name       string            // Human-readable name (e.g. "LLM Session")
	CreatedAt  time.Time         // Timestamp of session creation
	Tags       map[string]string // Optional metadata (e.g. project, owner, purpose)
	WindowRefs map[string]bool   // Map of window IDs owned by this session
}
