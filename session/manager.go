package session

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/owen-6936/termplex/window"
)

type WindowManager = window.WindowManager

// ShellSession represents a top-level orchestration unit.
// It owns windows, tracks creation metadata, and supports tagging for contributor clarity.

// SessionManager controls session lifecycle and enforces window limits.
type SessionManager struct {
	Sessions             map[string]*ShellSession
	Windows              map[string]*WindowManager
	MaxWindowsPerSession int
}

// NewSessionManager initializes a new SessionManager with a window limit.
func NewSessionManager(maxWindows int) *SessionManager {
	return &SessionManager{
		Sessions:             make(map[string]*ShellSession),
		Windows:              make(map[string]*WindowManager),
		MaxWindowsPerSession: maxWindows,
	}
}

// CreateSession registers a new shell session.
func (sm *SessionManager) CreateSession(name string, tags map[string]string) (string, error) {
	id := uuid.New().String()
	if _, exists := sm.Sessions[id]; exists {
		return "", errors.New("session ID collision")
	}

	sm.Sessions[id] = &ShellSession{
		ID:         id,
		Name:       name,
		CreatedAt:  time.Now(),
		Tags:       tags,
		WindowRefs: make(map[string]bool),
	}

	fmt.Printf("🧠 Session created: %s (%s)\n", id, name)
	return id, nil
}

// HasSession checks if a session exists.
func (sm *SessionManager) HasSession(id string) bool {
	_, exists := sm.Sessions[id]
	return exists
}

// GetSession retrieves a session by ID.
func (sm *SessionManager) GetSession(id string) (*ShellSession, bool) {
	session, exists := sm.Sessions[id]
	return session, exists
}

// TerminateSession removes a session and its windows.
func (sm *SessionManager) TerminateSession(id string) error {
	session, exists := sm.Sessions[id]
	if !exists {
		return errors.New("session not found")
	}

	for windowID := range session.WindowRefs {
		delete(sm.Windows, windowID)
	}

	delete(sm.Sessions, id)
	fmt.Printf("🧹 Session terminated: %s\n", id)
	return nil
}
