# Termplex: tmux Orchestration Core

**Termplex's `tmux` package** provides direct, process-aware control over tmux sessions, windows, and panes. It powers reproducible terminal environments, contributor overlays, and programmable CLI workflows.

---

## 🧠 Philosophy

- **Minimal abstraction**: Wraps raw `tmux` commands with clarity-first helpers
- **Process visibility**: Query pane state, working directories, and buffer contents
- **Composable primitives**: Designed for integration with Termplex's session manager, changelog engine, and contributor overlays
- **Test-friendly**: Detached session support, shell seeding, and CI-safe guards

---

## 🚀 Getting Started

```bash
go get github.com/nexicore/termplex/tmux
```

```go
import "github.com/nexicore/termplex/tmux"
```

---

## 🔧 Core Primitives

| Function | Purpose |
|----------|---------|
| `RunTmux(args ...string)` | Execute raw tmux commands |
| `NewSession(name string)` | Create a detached session |
| `NewWindow(session string)` | Add a window to a session |
| `SendKeys(target string, cmd string)` | Send keystrokes to a pane |
| `GetPanePath(target string)` | Query working directory of a pane |
| `CapturePane(target string)` | Capture visible buffer |
| `StartShell(target string, path string)` | Seed a shell in a pane |

---

## 🧪 Test Strategy

- Detached sessions with seeded shells
- CI-safe guards for interactive commands
- Functional changelogs for exported primitives

---

## 📚 License

MIT © 2025 Georgiy Komarov & Nexicore Digitals
