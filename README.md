# Termplex: tmux Orchestration Core

[![Build Status](https://github.com/owen-6936/termplex/actions/workflows/ci.yml/badge.svg)](https://github.com/owen-6936/termplex/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/owen-6936/termplex)](https://goreportcard.com/report/github.com/owen-6936/termplex)
[![GoDoc](https://godoc.org/github.com/owen-6936/termplex?status.svg)](https://godoc.org/github.com/owen-6936/termplex)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Termplex** is a Go-based terminal orchestration engine for managing multiplexed, process-aware terminal environments. It provides clean, testable primitives for creating, interacting with, and terminating shell sessions.

---

## 🧠 Philosophy

- **Layered Architecture**: A generic orchestration engine (`session`, `window`, `pane`, `shell`) provides the core logic, while specific backends (like `tmux`) implement the details.
- **Clarity-first abstraction**: Wraps raw `tmux` commands with minimal, intention-revealing helpers
- **Process introspection**: Query pane state, working directories, and buffer contents with precision
- **Composable primitives**: Built for integration with Termplex’s session manager, changelog engine, and contributor overlays
- **CI-safe and testable**: Supports detached sessions, shell seeding, and robust guards for non-interactive environments

---

## 🏛️ Core Architecture

| Package         | Role                                                                                             |
|-----------------|--------------------------------------------------------------------------------------------------|
| `window`        | Manages a collection of panes, representing a logical workspace.                                 |
| `pane`          | Manages a single view that can contain one interactive shell and multiple background processes.    |
| `shell`         | Provides low-level, stateless utilities for spawning, interacting with, and terminating OS processes (`exec.Cmd`). |
| `tmux`          | Provides a specific backend for orchestrating shells within a real `tmux` server environment.      |

This design separates the "what" (the state of windows and panes) from the "how" (the underlying process management), allowing for flexible and testable orchestration.

![Termplex Architecture Diagram](termplex-design.svg)

---

## 🚀 Getting Started

To see the generic orchestration engine in action, you can run the main demonstration file:

```bash
go run ./main.go
```

Import it in your Go project:

```go
import "github.com/nexicore/termplex/tmux"
```

---

## 🔧 API Reference (Module Tracker)

This section tracks the exported functions and methods for each core package, serving as a quick API reference.

### Generic Orchestration Engine

#### `session` Package

- `NewSessionManager(maxWindows int) *SessionManager`: Creates a manager for all sessions.
- `(sm *SessionManager) CreateSession(name, tags) (id, error)`: Creates a new top-level orchestration session.
- `(sm *SessionManager) AddWindow(sessionID, name, tags) (id, error)`: Adds a window to a specific session.
- `(sm *SessionManager) GetSession(id) (*Session, bool)`: Retrieves a session by its ID.
- `(sm *SessionManager) TerminateSession(id) error`: Terminates a session and all its child windows, panes, and shells.

#### `window` Package

- `NewWindowManager(name, tags) *WindowManager`: Creates a manager for a single window.
- `(wm *WindowManager) AddPane() (id, error)`: Adds a new pane to the window.
- `(wm *WindowManager) GetPane(id) (*pane.PaneManager, bool)`: Retrieves a pane by its ID.
- `(wm *WindowManager) TerminateWindow()`: Terminates a window and all its panes.

#### `pane` Package

- `NewPaneManager(id) *PaneManager`: Creates a manager for a single pane.
- `(pm *PaneManager) SpawnShell(interactive, command) (*shell.ShellSession, error)`: Spawns a new OS process within the pane.
- `(pm *PaneManager) TerminateShell(id, gracePeriod) (bool, error)`: Terminates a specific shell within the pane.
- `(pm *PaneManager) TerminatePane(gracePeriod)`: Terminates the pane and all shells running within it.

#### `shell` Package

- `(s *ShellSession) StartReading(stdoutHandler, stderrHandler)`: Starts goroutines to read from the shell's I/O pipes.
- `(s *ShellSession) SendCommand(command) error`: Sends a command to the shell's stdin (non-blocking).
- `(s *ShellSession) SendCommandAndWait(command) (output, error)`: Sends a command and blocks until it completes, returning its output.
- `(s *ShellSession) Close(gracePeriod) error`: Gracefully terminates the shell process with a force-kill fallback.

### `tmux` Backend

#### `tmux` Package

- `NewSessionManager(sessionName) (*SessionManager, error)`: Creates a new, real, detached `tmux` session.
- `(sm *SessionManager) AddPane() (*Pane, error)`: Adds a new pane to the `tmux` window by splitting it.
- `(sm *SessionManager) KillSession() error`: Destroys the entire `tmux` session.
- `(p *Pane) SendKeys(command) error`: Sends keystrokes to a specific `tmux` pane.
- `(p *Pane) Capture() (output, error)`: Captures the visible text content of a `tmux` pane.

---

## 🧪 Test Strategy

- Detached sessions with seeded shells for reproducibility
- CI-safe guards for interactive and non-interactive workflows
- Functional changelogs for exported primitives and orchestration events

---

## 📚 License

MIT © 2025 Georgiy Komarov & Nexicore Digitals
