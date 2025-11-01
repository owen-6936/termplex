# 📜 Termplex Functional Changelog

## 📄 Session Manifest Support

- **Declarative Sessions**: Implemented support for `.termplex.json` manifest files, allowing entire sessions, windows, panes, and startup commands to be defined declaratively.
- **`manifest` Package**: Created a new package to define the manifest structure (`model.go`) and parse the JSON files (`parser.go`).
- **`session.CreateSessionFromManifest(filePath)`**: Added a new method to the `SessionManager` that reads a manifest and orchestrates the creation of all defined resources.
- **Example Manifest**: Included `example.termplex.json` to demonstrate the format.
- **Testing**: Added a test suite for the `manifest` parser to ensure it correctly handles valid files, missing files, and invalid JSON.

---

## 🧠 Cognitive Milestone Tagging

- **`pane.AddTag(key, value)`**: Implemented a thread-safe method to add or update a tag on a `PaneManager`.
- **`pane.WaitForTag(key, value, timeout)`**: Implemented a blocking method that efficiently waits for a specific tag to be set, using `sync.Cond` to avoid busy-looping. This allows different parts of an orchestration to synchronize based on application-defined milestones (e.g., waiting for a service to be "ready").
- **Demo Update**: The `main.go` demo has been updated to showcase a real-world example of one shell waiting for a background service to signal its readiness via a tag.

---

## 🚀 `tmux` Backend Implementation

- **`tmux` Package**: Introduced a new `tmux` package to serve as a specific backend for the orchestration engine.
- **`tmux.SessionManager`**: Implemented a manager to create and terminate real, detached `tmux` sessions.
- **`tmux.Pane`**: Represents a `tmux` pane, with methods like `SendKeys` and `Capture` to interact with it via `tmux` commands.
- **Integration Test**: Added a `tmux`-specific integration test (`TestTmuxBackendLifecycle`) that validates the entire lifecycle against a real `tmux` server, ensuring correctness.

---

## ✨ Core Architecture Complete

- **`SessionManager`**: Implemented top-level manager for creating and terminating orchestration sessions.
- **`WindowManager`**: Manages collections of panes within a session.
- **`PaneManager`**: Manages shell processes within a single pane, enforcing the "one interactive shell" rule and providing a multiplexed `OutputChan` for all I/O.
- **End-to-End Testing**: All manager layers (`session`, `window`, `pane`) now have isolated unit tests to ensure correctness and stability.
- **Architectural Demo**: Added a `main.go` file that demonstrates the full, end-to-end flow of the architecture, from session creation to termination.

---

## 📦 Architecture & Core Packages

- **`shell` Package**:
  - **Introduced a new, stateless `shell` package** for low-level process management, decoupling it from pane/window abstractions.
  - **`ShellSession`**: A self-contained struct representing a managed OS process with its own `exec.Cmd`, I/O pipes, and output buffers.
  - **`StartReading`**: Automatically captures `stdout` and `stderr` into the session's internal buffers upon creation.
  - **`SendCommandAndWait`**: Provides reliable synchronous command execution with automatic UUID-based delimiter management.
  - **`Close`**: Implements a robust graceful shutdown with a configurable grace period and a force-kill fallback mechanism.
  - **Unit Tests**: Comprehensive, isolated test suite ensures reliability of the process lifecycle, I/O, and termination logic.

---

## 🧱 Core Primitives

- **`RunTmux(args...)`**: Raw command executor with stderr capture and error wrapping
- **`NewSession(name)`**: Detached session creation with idempotent guards
- **`KillSession(name)`**: Session cleanup for CI and test isolation
- **`HasSession(name)`**: Existence check for session lifecycle validation
- **`NewWindow(session)`**: Window creation with correct name parsing via `-F "#{window_name}"`
- **`SendKeys(target, cmd)`**: Keystroke injection for pane command execution
- **`NewPane(session, window)`**: Pane creation via `split-window` and index parsing from `list-panes`
- **`Pane.Target()`**: Target string builder for tmux pane addressing
- **`Pane.StartShell(path)`**: Shell seeding with `cd path && exec bash`
- **`Pane.GetCurrentPath()`**: Working directory query via `#{pane_current_path}`
- **`Pane.GetCurrentCommand()`**: Active process query via `#{pane_current_command}`

---

## 🧪 Test Suite

- **`window_test.go`**: Validates `NewWindow()` and `SendKeys()` with shell seeding and path resolution
- **`session_test.go`**: Validates session lifecycle: create, check, kill
- **`pane_test.go`**: Validates `NewPane()`, `StartShell()`, `GetCurrentPath()`, and `GetCurrentCommand()` with CI-safe assertions
- **GitHub Actions Integration**: `.github/workflows/test.yml` runs `go test -v ./...` on push and PR
- **CI Path Guarding**: Relaxed assertions for runner environments (`/home/runner/...`)

---

## 🧠 Cognitive Milestones

- **Shell seeded + CI verified**: Pane orchestration validated across local and CI environments
- **Window + Pane orchestration**: Split-window logic and targeting confirmed
- **Session lifecycle hardened**: Idempotent creation and cleanup for reproducible test runs

---

## 🗺️ Roadmap Progress

- ✅ Phase 1 primitives scaffolded and tested
- ✅ Phase 3 CI test suite integrated
- 🟡 Phase 2 changelog engine pending
- 🟡 Phase 4 process spawning next
