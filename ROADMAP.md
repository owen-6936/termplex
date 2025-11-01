# 🗺️ Termplex Roadmap

This roadmap outlines the evolution of **Termplex** as a modular, clarity-first terminal orchestration engine. It emphasizes composable primitives, contributor-friendly workflows, and programmable session management.

---

## ✅ Phase 1: Core Orchestration Primitives

- [x] `RunTmux(args...)`: raw command executor
- [x] `NewSession(name)`: create detached `tmux` session
- [x] `NewWindow(session)`: spawn window in session (implicitly created)
- [x] `NewPane(window)`: create pane in window via `split-window`
- [x] `SendKeys(target, cmd)`: send keystrokes to pane
- [x] `StartShell(target, path)`: seed shell in pane
- [x] `GetPanePath(target)`: query working directory
- [x] `CapturePane(target)`: capture visible buffer contents

---

## 🧠 Phase 2: Contributor Ecosystem

- [x] Exported module tracker (API Reference in `README.md`)
- [ ] ~~Functional changelog engine~~ (Scrapped for now)
- [ ] Cognitive milestone tagging (e.g. “Shell seeded”, “TTY isolated”)
- [ ] Session manifest format (`.termplex.json`)
- [ ] Reproducible test harnesses for CI environments

---

## 🧪 Phase 3: Test & Debug Layer

- [x] CI-safe test suite with seeded shell flows
- [x] `testenv.go`: guards for CI environment and `tmux` availability
- [x] `debug.go`: trace `tmux` commands and stderr via `TERMPLEX_DEBUG=1`
- [x] `assert.go`: test helpers for common assertions (`Contains`, `NoError`, etc.)

---

## 🎨 Phase 4: UI & Visualization

- [ ] `termplex-ui`: visual overlays for session/window/pane hierarchy
- [ ] Live buffer viewer with syntax highlighting
- [ ] Pane process inspector
- [ ] Contributor onboarding visualizer

---

## 🌐 Phase 5: Language & Documentation

- [ ] Multilingual UI support (Chinese + 2–3 additional languages)
- [ ] Clarity-first documentation with onboarding flows
- [ ] Visual glossary for tmux hierarchy (session → window → pane → process)

---

## 📦 Phase 6: Packaging & Distribution

- [ ] CLI wrapper (`termplex run`, `termplex inspect`)
- [ ] Go module release (`github.com/nexicore/termplex`)
- [x] GitHub Actions integration
- [ ] Contributor badge system (optional)

---

## 🧭 Meta Goals

- Build teachable systems that respect contributor agency  
- Scaffold cognitive clarity through naming, structure, and feedback  
- Celebrate technical wins as cognitive milestones  
- Prioritize testability and reproducibility in all designs  
- Foster an open-source community around terminal orchestration  
