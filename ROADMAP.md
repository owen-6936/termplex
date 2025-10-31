# 🗺️ Termplex Roadmap

This roadmap outlines the planned evolution of Termplex as a modular terminal orchestration engine. It focuses on clarity-first primitives, contributor-friendly workflows, and programmable session management.

---

## ✅ Phase 1: Core Orchestration Primitives

- [x] `RunTmux(args...)`: raw command executor
- [ ] `NewSession(name)`: create detached session
- [ ] `NewWindow(session)`: spawn window in session
- [ ] `NewPane(window)`: create pane in window
- [ ] `SendKeys(target, cmd)`: send keystrokes to pane
- [ ] `StartShell(target, path)`: seed shell in pane
- [ ] `GetPanePath(target)`: query working directory
- [ ] `CapturePane(target)`: capture visible buffer
- [ ] `PipePane(target, path)`: stream output to file

---

## 🧠 Phase 2: Contributor Ecosystem

- [ ] Functional changelog engine (replace badge attribution)
- [ ] Exported module tracker (name, params, usage)
- [ ] Cognitive milestone tagging (e.g. “Shell seeded”, “TTY isolated”)
- [ ] Session manifest format (`.termplex.json`)
- [ ] Reproducible test harnesses for CI

---

## 🧪 Phase 3: Test & Debug Layer

- [ ] `testenv.go`: guards for detached sessions, shell readiness
- [ ] `debug.go`: trace tmux stderr, command logs
- [ ] `assert.go`: test helpers for path, buffer, process
- [ ] CI-safe test suite with seeded shell flows

---

## 🎨 Phase 4: UI & Visualization

- [ ] `termplex-ui`: visual overlays for session/pane hierarchy
- [ ] Live buffer viewer with syntax highlighting
- [ ] Pane process inspector
- [ ] Contributor onboarding visualizer

---

## 🌐 Phase 5: Language & Docs

- [ ] Multilingual UI support (Chinese + 2–3 others)
- [ ] Clarity-first documentation with onboarding flows
- [ ] Visual glossary for tmux hierarchy (session → window → pane → process)

---

## 📦 Phase 6: Packaging & Distribution

- [ ] CLI wrapper (`termplex run`, `termplex inspect`)
- [ ] Go module release (`github.com/nexicore/termplex`)
- [ ] GitHub Actions integration
- [ ] Contributor badge system (optional)

---

## 🧭 Meta Goals

- Build teachable systems that respect contributor agency
- Scaffold cognitive clarity through naming, structure, and feedback
- Celebrate technical wins as cognitive milestones
- Prioritize testability and reproducibility in all designs
- Foster an open-source community around terminal orchestration
