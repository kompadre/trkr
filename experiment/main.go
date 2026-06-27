package main

import rl "github.com/gen2brain/raylib-go/raylib"

type Breakpoint int

const (
	SizeXS Breakpoint = iota // e.g., < 480px (RG351P width)
	SizeSM                   // e.g., 480px to 640px
	SizeMD                   // e.g., 640px to 800px (RG 40XX H sits here)
	SizeLG                   // e.g., > 800px
)

// GetCurrentBreakpoint evaluates the current window width from Raylib
func GetCurrentBreakpoint(width int) Breakpoint {
	if width < 480 {
		return SizeXS
	}
	if width < 640 {
		return SizeSM
	}
	if width < 800 {
		return SizeMD
	}
	return SizeLG
}

// LayoutContext keeps track of where the next widget should be drawn
// within the current structural boundary.
type LayoutContext struct {
	Bounds           rl.Rectangle
	CursorX, CursorY float32
	CurrentCol       int
	RowHeight        float32
	AvailableHeight  float32 // <--- Track the remaining vertical budget
}

type GuiSystem struct {
	ctxStack   []LayoutContext
	Breakpoint Breakpoint
}

var ui GuiSystem

func (g *GuiSystem) SetRowHeight(height float32) {
	if len(g.ctxStack) == 0 {
		return
	}
	ctx := &g.ctxStack[len(g.ctxStack)-1]

	// Force the active row to stretch to this height
	if height == 0 {
		ctx.RowHeight = ctx.AvailableHeight // Fill the remaining screen!
	} else {
		ctx.RowHeight = height
	}
}

func (g *GuiSystem) Col(xs, sm, md, lg int) (rl.Rectangle, bool) {
	if len(g.ctxStack) == 0 {
		return rl.Rectangle{}, false
	}

	ctx := &g.ctxStack[len(g.ctxStack)-1]

	// 1. Resolve span
	span := xs
	switch g.Breakpoint {
	case SizeSM:
		span = sm
	case SizeMD:
		span = md
	case SizeLG:
		span = lg
	}

	if span <= 0 {
		return rl.Rectangle{}, false
	}

	// 2. Wrap lines if we exceed 12 columns
	if ctx.CurrentCol+span > 12 {
		ctx.CursorX = ctx.Bounds.X
		ctx.CursorY += ctx.RowHeight

		// Reduce our available height budget by the row we just finished
		ctx.AvailableHeight -= ctx.RowHeight

		ctx.RowHeight = 0
		ctx.CurrentCol = 0
	}

	// 3. Calculate width and use a predictable row height for leaf blocks
	colUnitWidth := ctx.Bounds.Width / 12.0
	allocatedWidth := colUnitWidth * float32(span)

	// Setup a baseline row height if one isn't active
	if ctx.RowHeight < 24 {
		ctx.RowHeight = 24
	}

	// Sub-columns take the active row height, NOT the entire screen!
	allocatedHeight := ctx.RowHeight

	bounds := rl.Rectangle{
		X:      ctx.CursorX,
		Y:      ctx.CursorY,
		Width:  allocatedWidth,
		Height: allocatedHeight,
	}
	// 4. Update horizontal state for immediate siblings
	ctx.CursorX += allocatedWidth
	ctx.CurrentCol += span

	// Track the baseline row height (e.g., standard widgets might only use 24px)
	// We use 24 as a fallback baseline unless a child explicitly grows larger
	if ctx.RowHeight < 24 {
		ctx.RowHeight = 24
	}

	return bounds, true
}

func (g *GuiSystem) PushContext(bounds rl.Rectangle) {
	ctx := LayoutContext{
		Bounds:          bounds,
		CursorX:         bounds.X,
		CursorY:         bounds.Y,
		AvailableHeight: bounds.Height, // Tracks vertical real estate
		CurrentCol:      0,
		RowHeight:       0,
	}
	g.ctxStack = append(g.ctxStack, ctx)
}

// Add a clear, semantic wrapper for structural panels
func (g *GuiSystem) PushGreedyContext(bounds rl.Rectangle) {
	// Grab the parent context to see exactly what vertical budget is left
	parent := &g.ctxStack[len(g.ctxStack)-1]

	// Force the bounds height to fill the entire remaining vertical budget
	bounds.Height = parent.AvailableHeight

	g.PushContext(bounds)
}

func (g *GuiSystem) PopContext() {
	if len(g.ctxStack) < 2 {
		// Popping the root context, just remove it
		g.ctxStack = g.ctxStack[:len(g.ctxStack)-1]
		return
	}

	child := g.ctxStack[len(g.ctxStack)-1]
	g.ctxStack = g.ctxStack[:len(g.ctxStack)-1]

	parent := &g.ctxStack[len(g.ctxStack)-1]

	childHeightUsed := (child.CursorY + child.RowHeight) - child.Bounds.Y

	if childHeightUsed > parent.RowHeight {
		parent.RowHeight = childHeightUsed
	}

	parent.AvailableHeight -= childHeightUsed
}

func (g *GuiSystem) BreakRow() {
	if len(g.ctxStack) == 0 {
		return
	}
	ctx := &g.ctxStack[len(g.ctxStack)-1]

	// Move the cursor down past the current row's tallest element
	ctx.CursorX = ctx.Bounds.X
	ctx.CursorY += ctx.RowHeight

	// Deduct the completed row from the remaining height budget
	ctx.AvailableHeight -= ctx.RowHeight

	// Reset trackers for the fresh row
	ctx.RowHeight = 0
	ctx.CurrentCol = 0
}

func _main() {
	rl.InitWindow(1280, 1024, "bla")

	// 1. Root application layout (takes the whole screen)
	screenBounds := rl.Rectangle{X: 0, Y: 0, Width: float32(rl.GetScreenWidth()), Height: float32(rl.GetScreenHeight())}

	ui.Breakpoint = GetCurrentBreakpoint(rl.GetScreenWidth())

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		ui.PushContext(screenBounds)
		ui.SetRowHeight(0)

		// 1. Structural Columns use the greedy push to claim their vertical blocks
		if leftPane, v := ui.Col(0, 0, 4, 3); v {
			rl.DrawRectangleRec(leftPane, rl.Orange)
			// If leftPane needs an internal sub-grid, use PushGreedyContext(leftPane)
		}

		if mainPane, v := ui.Col(12, 12, 8, 9); v {
			rl.DrawRectangleRec(mainPane, rl.Red)

			// We want the main area to be a greedy container down the screen
			ui.PushGreedyContext(mainPane)

			// 2. Inner rows now automatically slice up neatly at 24px per row!
			for range 18 {
				if subCol, v := ui.Col(12, 6, 3, 1); v {
					rl.DrawRectangleRec(subCol, rl.Blue)

					// Clean outlines that wrap perfectly
					rl.DrawRectangleLinesEx(subCol, 1, rl.White)
				}
			}
			ui.BreakRow()
			// 2. Inner rows now automatically slice up neatly at 24px per row!
			for range 18 {
				if subCol, v := ui.Col(12, 6, 3, 1); v {
					rl.DrawRectangleRec(subCol, rl.Blue)

					// Clean outlines that wrap perfectly
					rl.DrawRectangleLinesEx(subCol, 1, rl.White)
				}
			}
			ui.PopContext()

		}

		ui.PopContext()
		rl.EndDrawing()
	}
}

func DrawSlider(bounds rl.Rectangle, text string) {
	rl.DrawRectangleRec(bounds, rl.Green)
	rl.DrawText(text+" [--|--]", bounds.ToInt32().X, bounds.ToInt32().Y, 20, rl.Yellow)
}

func DrawButton(bounds rl.Rectangle, text string) {
	rl.DrawRectangleRec(bounds, rl.Blue)
	rl.DrawText(text, bounds.ToInt32().X, bounds.ToInt32().Y, 20, rl.White)
}
