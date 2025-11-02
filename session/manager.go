package session

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/owen-6936/termplex/manifest"
	"github.com/owen-6936/termplex/window"
)

// SessionManager controls session lifecycle and enforces window limits.
type SessionManager struct {
	Sessions             map[string]*Session
	Windows              map[string]*window.WindowManager
	MaxWindowsPerSession int
}

// NewSessionManager initializes a new SessionManager with a window limit.
func NewSessionManager(maxWindows int) *SessionManager {
	return &SessionManager{
		Sessions:             make(map[string]*Session),
		Windows:              make(map[string]*window.WindowManager),
		MaxWindowsPerSession: maxWindows,
	}
}

// CreateSession registers a new shell session.
func (sm *SessionManager) CreateSession(name string, tags map[string]string) (string, error) {
	id := uuid.New().String()
	if _, exists := sm.Sessions[id]; exists {
		return "", errors.New("session ID collision")
	}

	sm.Sessions[id] = &Session{
		ID:         id,
		Name:       name,
		CreatedAt:  time.Now(),
		Tags:       tags,                  // A map is not a slice, so it remains.
		WindowRefs: make(map[string]bool), // A map is not a slice, so it remains.
	}

	fmt.Printf("🧠 Session created: %s (%s)\n", id, name)
	return id, nil
}

// AddWindow registers a new window in a session.
func (sm *SessionManager) AddWindow(sessionID string, name string, tags map[string]string) (string, error) {
	session, exists := sm.Sessions[sessionID]
	if !exists {
		return "", errors.New("session not found")
	}

	if len(session.WindowRefs) >= sm.MaxWindowsPerSession {
		return "", errors.New("session window limit reached")
	}

	windowID := uuid.New().String()
	if _, exists := sm.Windows[windowID]; exists {
		return "", errors.New("window ID collision")
	}

	sm.Windows[windowID] = window.NewWindowManager(name, tags)
	session.WindowRefs[windowID] = true
	return windowID, nil
}

// HasSession checks if a session exists.
func (sm *SessionManager) HasSession(id string) bool {
	_, exists := sm.Sessions[id]
	return exists
}

// GetSession retrieves a session by ID.
func (sm *SessionManager) GetSession(id string) (*Session, bool) {
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
		sm.Windows[windowID].TerminateWindow()
		delete(sm.Windows, windowID)
	}

	delete(sm.Sessions, id)
	fmt.Printf("🧹 Session terminated: %s\n", id)
	return nil
}

// CreateSessionFromManifest reads a manifest file, parses it, and builds the entire
// session, including all windows, panes, and startup shells/commands.
func (sm *SessionManager) CreateSessionFromManifest(filePath string) (string, error) {
	m, err := manifest.LoadFromFile(filePath)
	if err != nil {
		return "", fmt.Errorf("could not load session from manifest: %w", err)
	}

	// 1. Create the top-level session.
	sessionID, err := sm.CreateSession(m.SessionName, m.SessionTags)
	if err != nil {
		return "", err
	}

	// 2. Iterate over windows defined in the manifest.
	for _, winManifest := range m.Windows {
		windowID, err := sm.AddWindow(sessionID, winManifest.WindowName, winManifest.WindowTags)
		if err != nil {
			return "", err // Or handle error more gracefully
		}
		wm := sm.Windows[windowID]

		// 3. Iterate over panes for each window.
		for _, paneManifest := range winManifest.Panes {
			paneID, err := wm.AddPane(paneManifest.PaneName)
			if err != nil {
				return "", err
			}
			pane, _ := wm.GetPane(paneID)

			// 4. Spawn the startup shell for the pane.
			shell, err := pane.SpawnShell(paneManifest.StartupShell.Interactive, paneManifest.StartupShell.Command...)
			if err != nil {
				return "", err
			}

			// 5. Send any startup commands to the newly created shell.
			for _, cmd := range paneManifest.StartupCommands {
				_ = shell.SendCommand(cmd)
			}
		}
	}

	return sessionID, nil
}
