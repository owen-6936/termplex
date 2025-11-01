# Termplex: tmux Orchestration Core

**Termplex’s `tmux` package** offers direct, process-aware control over tmux sessions, windows, and panes. It powers reproducible terminal environments, contributor overlays, and programmable CLI workflows for multitasking-aware shell orchestration.

---

## 🧠 Philosophy

- **Clarity-first abstraction**: Wraps raw `tmux` commands with minimal, intention-revealing helpers
- **Process introspection**: Query pane state, working directories, and buffer contents with precision
- **Composable primitives**: Built for integration with Termplex’s session manager, changelog engine, and contributor overlays
- **CI-safe and testable**: Supports detached sessions, shell seeding, and robust guards for non-interactive environments

---

## 🚀 Getting Started

Install the package:

```bash
go get github.com/nexicore/termplex/tmux
```

Import it in your Go project:

```go
import "github.com/nexicore/termplex/tmux"
```

---

## 🔧 Core Primitives

| Function                      | Purpose                                      |
|------------------------------|----------------------------------------------|
| `RunTmux(args ...string)`    | Execute raw tmux commands                    |
| `NewSession(name string)`    | Create a detached tmux session               |
| `NewWindow(session string)`  | Add a window to an existing session          |
| `SendKeys(target string, cmd string)` | Send keystrokes to a target pane     |
| `GetPanePath(target string)` | Query the working directory of a pane        |
| `CapturePane(target string)` | Capture the visible buffer of a pane         |
| `StartShell(target string, path string)` | Seed a shell in a target pane     |

---

## 🧪 Test Strategy

- Detached sessions with seeded shells for reproducibility
- CI-safe guards for interactive and non-interactive workflows
- Functional changelogs for exported primitives and orchestration events

---

## 📚 License

MIT © 2025 Georgiy Komarov & Nexicore Digitals
