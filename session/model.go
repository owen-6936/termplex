package session

import "time"

// Session represents a top-level orchestration unit, like a workspace or project.
// It owns windows, tracks creation metadata, and supports tagging for contributor clarity.
type Session struct {
	ID         string            // Unique session ID
	Name       string            // Human-readable name (e.g. "LLM Session")
	CreatedAt  time.Time         // Timestamp of session creation
	Tags       map[string]string // Optional metadata (e.g. project, owner, purpose)
	WindowRefs map[string]bool   // Map of window IDs owned by this session
}
