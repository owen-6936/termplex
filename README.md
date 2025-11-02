# Termplex: A Go Engine for Terminal Orchestration

[![Build Status](https://github.com/owen-6936/termplex/actions/workflows/ci.yml/badge.svg)](https://github.com/owen-6936/termplex/actions/workflows/ci.yml)
![Go Report Card](https://goreportcard.com/badge/github.com/owen-6936/termplex?t=1)
[![GoDoc](https://godoc.org/github.com/owen-6936/termplex?status.svg)](https://godoc.org/github.com/owen-6936/termplex)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Termplex** is a Go library for building process-aware, multiplexed terminal applications. It provides two main components:

1. A **Generic Orchestration Engine** for managing hierarchical workspaces of windows, panes, and shell processes.
2. A **`tmux` Backend** that uses the engine's principles to automate a real `tmux` server.

---

## 🧠 Philosophy

- **Layered Architecture**: A generic orchestration engine (`session`, `window`, `pane`, `shell`) provides the core logic, while specific backends (like `tmux`) implement the details.
- **Clarity-first abstraction**: Wraps raw `tmux` commands with minimal, intention-revealing helpers
- **Process Introspection**: Provides direct access to shell I/O streams and output buffers.
- **Robust I/O**: Uses pseudo-terminals (PTY) for interactive shells to ensure correct TTY behavior, while being fully thread-safe for concurrent operations.
- **CI-safe and testable**: Supports detached sessions, shell seeding, and robust guards for non-interactive environments

---

## 🏛️ Core Concepts

| Package         | Role                                                                                             |
|-----------------|--------------------------------------------------------------------------------------------------|
| `window`        | Manages a collection of panes, representing a logical workspace.                                 |
| `pane`          | Manages a single view that can contain one interactive shell and multiple background processes.    |
| `shell`         | Provides low-level, stateless utilities for spawning, interacting with, and terminating OS processes (`exec.Cmd`). |
| `tmux`          | Provides a specific backend for orchestrating shells within a real `tmux` server environment.      |

view the full symbol documentation at [API Reference](API.md)

This design separates the "what" (the state of windows and panes) from the "how" (the underlying process management), allowing for flexible and testable orchestration.

![Termplex Go Architecture](termplex-design.svg)

For a conceptual overview of the `tmux` object model, see the Visual Glossary for the [`tmux` Hierarchy](tmux_hierarchy.md).

---

## 🚀 Getting Started

### Installation

```bash
go get github.com/owen-6936/termplex
```

Import it in your Go project:

```go
import "github.com/owen-6936/termplex/tmux"
```

---

## 📄 Declarative Sessions with Manifests

Termplex supports declarative session creation using a `.termplex.json` manifest file. This allows you to define an entire workspace—including windows, panes, and startup commands—in a single, version-controllable file.

You can create a session from a manifest like this:

```go
sm := session.NewSessionManager(5)
sessionID, err := sm.CreateSessionFromManifest("path/to/your/manifest.json")
```

### Example Manifest (`example.termplex.json`)

```json
{
  "sessionName": "WebAppDev",
  "sessionTags": {
    "project": "termplex-demo"
  },
  "windows": [
    {
      "windowName": "Backend",
      "panes": [
        {
          "paneTags": { "role": "api-server" },
          "startupShell": {
            "interactive": true,
            "command": ["bash", "-i"]
          },
          "startupCommands": [
            "echo 'Starting API server... (simulated)'",
            "go --version"
          ]
        },
        {
          "paneTags": { "role": "database-logs" },
          "startupShell": {
            "interactive": false,
            "command": ["bash", "-c", "echo 'Tailing database logs...'; sleep 2"]
          }
        }
      ]
    }
  ]
}
```

---

## 🧪 Test Strategy

- Detached sessions with seeded shells for reproducibility
- CI-safe guards for interactive and non-interactive workflows
- Functional changelogs for exported primitives and orchestration events

---

## 📚 License

MIT © 2025 Owen Erhabor
