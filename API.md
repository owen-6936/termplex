# API Reference (Module Tracker)

This document tracks the exported functions and methods for each core package, serving as a quick API reference.

## Generic Orchestration Engine

### `session` Package

- `NewSessionManager(maxWindows int) *SessionManager`: Creates a manager for all sessions.
- `(sm *SessionManager) CreateSession(name, tags) (id, error)`: Creates a new top-level orchestration session.
- `(sm *SessionManager) AddWindow(sessionID, name, tags) (id, error)`: Adds a window to a specific session.
- `(sm *SessionManager) GetSession(id) (*Session, bool)`: Retrieves a session by its ID.
- `(sm *SessionManager) TerminateSession(id) error`: Terminates a session and all its child windows, panes, and shells.
- `(sm *SessionManager) CreateSessionFromManifest(filePath) (id, error)`: Builds an entire session from a `.termplex.json` file.

### `window` Package

- `NewWindowManager(name, tags) *WindowManager`: Creates a manager for a single window.
- `(wm *WindowManager) AddPane() (id, error)`: Adds a new pane to the window.
- `(wm *WindowManager) GetPane(id) (*pane.PaneManager, bool)`: Retrieves a pane by its ID.
- `(wm *WindowManager) TerminateWindow()`: Terminates a window and all its panes.

### `pane` Package

- `NewPaneManager(id) *PaneManager`: Creates a manager for a single pane.
- `(pm *PaneManager) SpawnShell(interactive, command) (*shell.ShellSession, error)`: Spawns a new OS process within the pane.
- `(pm *PaneManager) TerminateShell(id, gracePeriod) (bool, error)`: Terminates a specific shell within the pane.
- `(pm *PaneManager) TerminatePane(gracePeriod)`: Terminates the pane and all shells running within it.
- `(pm *PaneManager) AddTag(key, value)`: Safely adds a tag to the pane to signal a milestone.
- `(pm *PaneManager) WaitForTag(key, value, timeout) error`: Blocks until a specific tag is set, or a timeout occurs.

### `shell` Package

- `(s *ShellSession) StartReading(stdoutHandler, stderrHandler)`: Starts goroutines to read from the shell's I/O pipes.
- `(s *ShellSession) SendCommand(command) error`: Sends a command to the shell's stdin (non-blocking).
- `(s *ShellSession) SendCommandAndWait(command) (output, error)`: Sends a command and blocks until it completes, returning its output.
- `(s *ShellSession) Close(gracePeriod) error`: Gracefully terminates the shell process with a force-kill fallback.

## `tmux` Backend

### `tmux` Package

- `NewSessionManager(sessionName) (*SessionManager, error)`: Creates a new, real, detached `tmux` session.
- `(sm *SessionManager) AddPane() (*Pane, error)`: Adds a new pane to the `tmux` window by splitting it.
- `(sm *SessionManager) KillSession() error`: Destroys the entire `tmux` session.
- `(p *Pane) SendKeys(command) error`: Sends keystrokes to a specific `tmux` pane.
- `(p *Pane) Capture() (output, error)`: Captures the visible text content of a `tmux` pane.
