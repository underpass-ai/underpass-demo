# USS Underpass Demo — VHS Recording Tape
# Produces a ~2 minute video showing the full spaceship demo.
#
# Requirements: vhs (go install github.com/charmbracelet/vhs@latest)
# Usage:        vhs demo.tape
# Output:       demo.gif + demo.mp4

Require tlctl

Output demo.gif
Output demo.mp4

Set Shell "bash"
Set FontSize 15
Set Width 1400
Set Height 850
Set Padding 10
Set Margin 20
Set MarginFill "#1e1e2e"
Set WindowBar Colorful
Set BorderRadius 8
Set Framerate 30
Set TypingSpeed 50ms
Set Theme "Catppuccin Mocha"
Set CursorBlink false

# ─── Launch ──────────────────────────────────────────────────

Type "./bin/tlctl --embedded"
Sleep 500ms
Enter
Sleep 2s

# ─── Mission: Phase 0 — NOMINAL ─────────────────────────────
# The TUI opens on the Mission view. All systems green.
Sleep 3s

# ─── Phase 1 — ANOMALY ──────────────────────────────────────
Space
Sleep 3s

# ─── Phase 2 — ENGINE CORE FAILURE ──────────────────────────
Space
Sleep 3s

# ─── Phase 3 — CASCADE ──────────────────────────────────────
Space
Sleep 2500ms

# ─── Phase 4 — THOMPSON SAMPLING ADAPTS ─────────────────────
Space
Sleep 3s

# ─── Quick detour: Agent Dispatch view ───────────────────────
# Show event-driven agents firing in real time
Type "a"
Sleep 6s

# ─── Back to Mission ────────────────────────────────────────
Type "m"
Sleep 1s

# ─── Phase 5 — WRONG REPAIR STRATEGY ────────────────────────
Space
Sleep 3s

# ─── Phase 6 — ESCALATION (Claude Opus) ─────────────────────
Space
Sleep 3500ms

# ─── Phase 7 — REHYDRATION (the money shot) ─────────────────
Space
Sleep 5s

# ─── Phase 8 — NEW BRANCH (hull-first working) ──────────────
Space
Sleep 3s

# ─── Phase 9 — RESOLUTION ───────────────────────────────────
Space
Sleep 4s

# ─── Thompson Sampling live draws ───────────────────────────
Type "t"
Sleep 5s

# ─── Bridge overview ────────────────────────────────────────
Type "b"
Sleep 3s

# ─── Systems table ──────────────────────────────────────────
Type "s"
Sleep 2s

# ─── Exit ────────────────────────────────────────────────────
Type "q"
Sleep 1s
