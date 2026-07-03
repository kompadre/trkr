package ui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	. "trkr"
)

type Breakpoint int

const (
	SizeXS Breakpoint = iota // e.g., < 480px (RG351P width)
	SizeSM                   // e.g., 480px to 640px
	SizeMD                   // e.g., 640px to 800px (RG 40XX H sits here)
	SizeLG                   // e.g., > 800px
)

// GetCurrentBreakpoint evaluates the current window width from Raylib
func GetCurrentBreakpoint(width int) Breakpoint {
	if width < 400 {
		return SizeXS
	}
	if width < 800 {
		return SizeSM
	}
	if width < 1280 {
		return SizeMD
	}
	return SizeLG
}

// LaidContext keeps track of where the next widget should be drawn
// within the current structural boundary.
type LaidContext struct {
	Bounds           rl.Rectangle
	CursorX, CursorY float32
	CurrentCol       int
	RowHeight        float32
	AvailableHeight  float32 // <--- Track the remaining vertical budget
}

type Laid struct {
	ctxStack   []LaidContext
	ColBounds  rl.Rectangle
	Breakpoint Breakpoint
}

func (l *Laid) SetBreakpoint(width int) {
	l.Breakpoint = GetCurrentBreakpoint(width)
	Logf("Breakpoint detected: %d.\n", l.Breakpoint)
}

func (l *Laid) Context() *LaidContext {
	if len(l.ctxStack) == 0 {
		rect := rl.Rectangle{X: 0, Y: 0, Width: float32(rl.GetScreenWidth()), Height: float32(rl.GetScreenHeight())}
		l.PushContext(rect)
		l.SetRowHeight(0)
	}
	return &l.ctxStack[len(l.ctxStack)-1]
}

func (l *Laid) Bounds() rl.Rectangle {
	if l.ColBounds.Width > 0 {
		return l.ColBounds
	} else if len(l.ctxStack) > 0 {
		return l.ctxStack[len(l.ctxStack)-1].Bounds
	}
	return rl.Rectangle{}
}

func (l *Laid) SetRowHeight(height float32) {
	if len(l.ctxStack) == 0 {
		return
	}
	ctx := &l.ctxStack[len(l.ctxStack)-1]

	// Force the active row to stretch to this height
	if height == 0 {
		ctx.RowHeight = ctx.AvailableHeight // Fill the remaining screen!
	} else {
		ctx.RowHeight = height
	}
}

// Treat current col bounds as section
// Subsequent calls to col  will subdivide current col
func (l *Laid) EnterCol(xs, sm, md, lg int) bool {
	if b, v := l.Col(xs, sm, md, lg); v {
		l.PushContext(b)
		l.SetRowHeight(0)
		return true
	}
	return false
}

func (l *Laid) ExitCol() {
	l.PopContext()
}

func (l *Laid) Pad(x, y float32) {
	ctx := l.Context()
	ctx.CursorX += x
	ctx.CursorY += y
	ctx.AvailableHeight -= y
	ctx.Bounds.X += x
	ctx.Bounds.Width -= x
	ctx.Bounds.Y += y
	ctx.Bounds.Height -= y
}

func (l *Laid) Col(xs, sm, md, lg int) (rl.Rectangle, bool) {
	l.ColBounds = rl.Rectangle{}
	if len(l.ctxStack) == 0 {
		return rl.Rectangle{}, false
	}

	ctx := &l.ctxStack[len(l.ctxStack)-1]

	// 1. Resolve span
	span := xs
	switch l.Breakpoint {
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
	if ctx.RowHeight < 16 && ctx.RowHeight != 0 {
		ctx.RowHeight = 16
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

	if ctx.RowHeight < 16 {
		ctx.RowHeight = 16
	}
	l.ColBounds = bounds
	return bounds, true
}

func (l *Laid) PushContext(bounds rl.Rectangle) {
	ctx := LaidContext{
		Bounds:          bounds,
		CursorX:         bounds.X,
		CursorY:         bounds.Y,
		AvailableHeight: bounds.Height, // Tracks vertical real estate
		CurrentCol:      0,
		RowHeight:       0,
	}
	l.ctxStack = append(l.ctxStack, ctx)
}

// Add a clear, semantic wrapper for structural panels
func (l *Laid) PushGreedyContext(bounds rl.Rectangle) {
	// Grab the parent context to see exactly what vertical budget is left
	parent := &l.ctxStack[len(l.ctxStack)-1]

	// Force the bounds height to fill the entire remaining vertical budget
	bounds.Height = parent.AvailableHeight

	l.PushContext(bounds)
}

func (l *Laid) PopContext() {
	if len(l.ctxStack) < 2 {
		// Popping the root context, just remove it
		l.ctxStack = l.ctxStack[:len(l.ctxStack)-1]
		return
	}

	child := l.ctxStack[len(l.ctxStack)-1]
	l.ctxStack = l.ctxStack[:len(l.ctxStack)-1]

	parent := &l.ctxStack[len(l.ctxStack)-1]

	childHeightUsed := (child.CursorY + child.RowHeight) - child.Bounds.Y

	if childHeightUsed > parent.RowHeight {
		parent.RowHeight = childHeightUsed
	}

	parent.AvailableHeight -= childHeightUsed
}

func (l *Laid) BreakRow() {
	if len(l.ctxStack) == 0 {
		return
	}
	ctx := &l.ctxStack[len(l.ctxStack)-1]

	// Move the cursor down past the current row's tallest element
	ctx.CursorX = ctx.Bounds.X
	ctx.CursorY += ctx.RowHeight

	// Deduct the completed row from the remaining height budget
	ctx.AvailableHeight -= ctx.RowHeight

	// Reset trackers for the fresh row
	ctx.RowHeight = 0
	ctx.CurrentCol = 0
}

func (l *Laid) Text(text string, size int32, color rl.Color) {
	b := l.Bounds()
	DrawText(text, int32(b.X), int32(b.Y), size, color)
	ctx := &l.ctxStack[len(l.ctxStack)-1]
	ctx.RowHeight = max(float32(size), ctx.RowHeight)
}

func (l *Laid) Pixel(x, y int32, color rl.Color) {
	b := l.Bounds()
	Logf("Bounds: %v.\n", b)
	rl.DrawPixel(int32(b.X)+x, int32(b.Y)+y, color)
}

func (l *Laid) TextBlock(text string, size int32, color rl.Color) {
	l.Col(12, 12, 12, 12)
	l.Text(text, size, color)
}
