# 📜 Termplex Functional Changelog

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
