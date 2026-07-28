# AGENTS.md

Lightweight handheld music tracker in Go using raylib, C++ DX7 MSFA synthesis, and miniaudio.

## Build & Test Commands

- **Build native binary**: `go build ./cmd/trkr` or `make linux-wayland`
- **Build targets**:
  - `make linux-amd64` (X11 build -> `trkr-x11`)
  - `make linux-arm64` (ARM64 handheld build via Docker container `Dockerfile.xcompile-arm64`)
  - `make win64` (Windows 64-bit build -> `trkr.exe`, requires `x86_64-w64-mingw32-gcc`/`g++`)
  - `make release` (Release build -> `trkr-release`)
- **Run all tests**: `go test ./...`
- **Run package tests**: `go test ./internal/player`, `go test ./internal/events`, `go test ./internal/audio/effects`, `go test ./internal/ui/widget`
- **Run single test**: `go test ./internal/events -run TestInputSnapshot`
- **Lint / Static check**: `go vet ./...`

## Key Architecture & Boundaries

- **`cmd/trkr/main.go`**: App entry point. Manages raylib window initialization, font loading, event loop execution, and project autosaving (`autosave.json`).
- **`trkr` (root package, `trkr.go`)**: Core data models (`Project`, `Track`, `Phrase`, `Step`, `Note`, `Instrument`, `Section`).
- **`internal/audio/`**: Audio pipeline.
  - `external/msfa/`: C++ Yamaha DX7 synthesis engine bridged via CGO (`msfa.go`). Supports per-slot voice loading and project-linked bank persistence.
  - `internal/audio/miniaudio/`: CGO wrapper around miniaudio audio backend.
  - `internal/audio/perc/`: Percussion synthesizer.
  - `internal/audio/effects/`: DSP audio effects (`delay.go`, `filter.go`).
  - `internal/audio/fm/`: DX7 bank utilities (`bank.go`) for parsing and serializing `.syx` files with Yamaha checksums.
- **`internal/events/`**: Central event pipeline (`Input`, `Update`, `PostUpdate`, `AudioUpdate`) and `InputSnapshot`.
- **`internal/player/`**: Transport control, section playhead engine, timing loops, and sample-accurate offline WAV export with look-ahead gain smoothing.

## Functional Safeguards & Tests

- **Integration Tests**: `internal/player/sequencer_integration_test.go` verifies note-accurate triggering, phrase repeats, and section transitions.
- **Audio Tests**: `internal/audio/mixer_test.go` verifies signal path and look-ahead saturator effectiveness.
- **Persistence Tests**: `trkr_test.go` verifies project save/load cycles for both JSON and project-linked `.syx` banks.

## Feature Workflows

- **FM Patch Browser**:
  - Located in `internal/ui/view/patch_browser.go`.
  - Scans `assets/syx/` recursively.
  - Allows auditioning voices in real-time.
  - Supports "Frankensteining" a project bank by copying voices into the first available empty slot.
  - Modal interaction: Always returns `true` in `HandleInput` to capture all events.
- **WAV Export**:
  - Implements an overflow-buffer system to maintain fixed block sizes for DSP effects (limiter/smoothing) during non-integer tick durations.
  - Synchronizes commands at the start of every row for perfect alignment.

## UI System & Layout


- **`ui.Element`**: Every UI component is an `Element`. They form a tree starting from `ui.RootElement`.
- **Layout Styles**:
  - **Grid (Columns)**: Set `el.Col = [4]int{xs, sm, md, lg}` (1-12 span) to participate in the `Laid` 12-column flex-grid.
  - **Anchors**: Set `el.IsAnchor = true` to create an independent coordinate system (useful for dialogs/popups).
  - **Rows**: Use `ui.NewRow(height, parent)` to create a full-width vertical block.
- **`Laid` Layout Engine**:
  - Uses a context stack (`PushContext`/`PopContext`).
  - `EnterCol()` creates a sub-grid; `ExitCol()` returns to parent.
  - **Gotcha**: Do NOT manually `SetRowHeight()` on Anchor elements; it will cause their children to inherit the container's full height as a line-height.
- **Widgets**:
  - **Knob**: `widget.NewKnob(label, col, min, max, initial, action, parent)`
  - **Input**: `widget.NewInput(label, x, y, w, h, action, parent)` (supports `WidgetInputTypeNumber`)
  - **Button**: `widget.NewButton(label, col, height, action, parent)`
- **Input & Focus**:
  - `el.HandleInput(snapshot)` dispatches input to the tree.
  - **Leaf Core vs. Container Elements**:
    - **Leaf Core Elements** (`e.Core != nil && len(e.Children) == 0`): Includes widgets (knobs, inputs, buttons) and monolithic views (`TrackDialog`, `SongDialog`, `PatchBrowserDialog`). They handle input exclusively via their custom core. Fallback container navigation is disabled.
    - **Container Elements** (`len(e.Children) > 0`): Includes dialog managers and layout rows. Their primary duty is navigating and focusing their children. They support optional custom core interceptors for global actions (like 'B' to save/close).
  - **Pre-focus System**: Navigation is separated from interaction.
    - **Highlighting**: D-pad navigation uses `el.HighlightJump(1/-1)` to cycle through visible children and highlight them. The active highlight state is passed to elements during `Draw` as the `isHighlighted` boolean.
    - **Seamless Nested Navigation**: If a parent container has `FocusedChild = nil` but its highlighted child is also a layout container (has children), input is automatically delegated to the highlighted child, allowing Left/Right/Up/Down to traverse seamlessly across nested grid layout rows.
    - **Interaction**: Pressing **'A'** focuses/activates the highlighted child (`FocusedChild = HighlightedChild`). This automatically propagates up the tree (`e.Parent.FocusedChild = e`) to establish a synchronized focus path from the root to the active leaf widget.
    - **Deactivation**: Pressing **'B'** (or **'A'** depending on the widget) unfocuses/deactivates the widget, returning navigation to the container level (`e.FocusedChild = nil`).
    - **Initial State**: All containers start in pure "Navigation Mode" by default (`FocusedChild = nil` and `HighlightedChild` pointing to the first child) to prevent widgets from auto-activating on focus gain.
  - Set `el.FocusOutAfterLast = true` to allow focus/highlight to escape a container back to its parent.
  - **Modal**: Set `el.IsModal = true` to make the element capture all input for its parent's subtree and be the only child drawn by that parent.
- **Rendering**:
  - Use `ui.DrawText()` for batched, pixel-perfect font rendering.
  - Call `ui.FlushText()` once at the end of the frame.

## Toolchain & Runtime Quirks

- **CGO Required**: `CGO_ENABLED=1` is required due to C/C++ code in `external/msfa`, `internal/audio/miniaudio`, and `internal/ui/drawtext.c`.
- **Raylib Fork**: `go.mod` uses a local replace directive mapping `github.com/gen2brain/raylib-go/raylib` to `./external/raylib-go/raylib`.
- **Working Directory**: The binary expects to be executed with the workspace root as working directory to load relative paths (`./assets/`, `./autosave.json`).
