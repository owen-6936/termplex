# 🗺️ Termplex Roadmap

This roadmap outlines the evolution of **Termplex** as a modular, clarity-first terminal orchestration engine. It emphasizes composable primitives, contributor-friendly workflows, and programmable session management.

---

## ✅ Phase 1: Core Orchestration Primitives

- [x] `RunTmux(args...)`: raw command executor
- [x] `NewSession(name)`: create detached tmux session
- [x] `NewWindow(session)`: spawn window in session
- [x] `NewPane(window)`: create pane in window
- [x] `SendKeys(target, cmd)`: send keystrokes to pane
- [x] `StartShell(target, path)`: seed shell in pane
- [x] `GetPanePath(target)`: query working directory
- [ ] `CapturePane(target)`: capture visible buffer contents
- [ ] `PipePane(target, path)`: stream pane output to file

---

## 🧠 Phase 2: Contributor Ecosystem

- [ ] Functional changelog engine (replaces badge-based attribution)
- [ ] Exported module tracker (name, params, usage)
- [ ] Cognitive milestone tagging (e.g. “Shell seeded”, “TTY isolated”)
- [ ] Session manifest format (`.termplex.json`)
- [ ] Reproducible test harnesses for CI environments

---

## 🧪 Phase 3: Test & Debug Layer

- [x] CI-safe test suite with seeded shell flows
- [ ] `testenv.go`: guards for detached sessions and shell readiness
- [ ] `debug.go`: trace tmux stderr and command logs
- [ ] `assert.go`: test helpers for path, buffer, and process state

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
