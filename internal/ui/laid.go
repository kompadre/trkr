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
	AvailableHeight  float32 // Track the remaining vertical budget

	// Width & Block Tracking
	BlockWidth     float32 // Total width of current block/container
	AllocatedWidth float32 // Cumulative width allocated in current row
	RemainingWidth float32 // Remaining horizontal budget in current row
	MaxRowWidth    float32 // Peak width used by content across rows
}

type Laid struct {
	ctxStack     []LaidContext
	ColBounds    rl.Rectangle
	Breakpoint   Breakpoint

	// Layout Caching & Element Tree Integration
	cachedBounds map[uint16]rl.Rectangle
	lastWidth    int
	lastHeight   int
	isDirty      bool
}

func (l *Laid) Invalidate() {
	l.isDirty = true
}

func (l *Laid) IsDirty() bool {
	return l.isDirty || l.cachedBounds == nil
}

func (l *Laid) GetCachedBounds(id uint16) (rl.Rectangle, bool) {
	if l.IsDirty() || id == 0 {
		return rl.Rectangle{}, false
	}
	rect, ok := l.cachedBounds[id]
	return rect, ok
}

func (l *Laid) SetCachedBounds(id uint16, rect rl.Rectangle) {
	if l.cachedBounds == nil {
		l.cachedBounds = make(map[uint16]rl.Rectangle)
	}
	l.cachedBounds[id] = rect
}

func (l *Laid) BeginFrame(width, height int) {
	bp := GetCurrentBreakpoint(width)
	if width != l.lastWidth || height != l.lastHeight || bp != l.Breakpoint {
		l.Breakpoint = bp
		l.lastWidth = width
		l.lastHeight = height
		l.isDirty = true
	}
	if l.isDirty {
		l.cachedBounds = make(map[uint16]rl.Rectangle)
		l.isDirty = false
	}
}

func (l *Laid) SetBreakpoint(width int) {
	bp := GetCurrentBreakpoint(width)
	if bp != l.Breakpoint {
		l.Breakpoint = bp
		l.Invalidate()
	}
}

func (l *Laid) Context() *LaidContext {
	if len(l.ctxStack) == 0 {
		w, h := float32(rl.GetScreenWidth()), float32(rl.GetScreenHeight())
		if w <= 0 {
			w = 480
		}
		if h <= 0 {
			h = 320
		}
		rect := rl.Rectangle{X: 0, Y: 0, Width: w, Height: h}
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

// BlockWidth returns total width of the current structural block context
func (l *Laid) BlockWidth() float32 {
	if len(l.ctxStack) == 0 {
		return float32(rl.GetScreenWidth())
	}
	return l.ctxStack[len(l.ctxStack)-1].BlockWidth
}

// RemainingWidth returns remaining horizontal space in the current row
func (l *Laid) RemainingWidth() float32 {
	if len(l.ctxStack) == 0 {
		return 0
	}
	return l.ctxStack[len(l.ctxStack)-1].RemainingWidth
}

// AllocatedWidth returns accumulated width in the current row
func (l *Laid) AllocatedWidth() float32 {
	if len(l.ctxStack) == 0 {
		return 0
	}
	return l.ctxStack[len(l.ctxStack)-1].AllocatedWidth
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
// Subsequent calls to col will subdivide current col
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
	ctx.BlockWidth = ctx.Bounds.Width
	ctx.RemainingWidth = max(0, ctx.Bounds.Width-ctx.AllocatedWidth)
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
		l.BreakRow()
	}

	// 3. Calculate width with exact block width tracking & subpixel overflow prevention
	colUnitWidth := ctx.Bounds.Width / 12.0
	allocatedWidth := colUnitWidth * float32(span)

	// Sub-pixel snapping for exact 12-column grid alignment
	if ctx.CurrentCol+span == 12 || ctx.CursorX+allocatedWidth > ctx.Bounds.X+ctx.Bounds.Width {
		allocatedWidth = (ctx.Bounds.X + ctx.Bounds.Width) - ctx.CursorX
	}

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

	// 4. Update horizontal state & block width trackers
	ctx.CursorX += allocatedWidth
	ctx.CurrentCol += span
	ctx.AllocatedWidth += allocatedWidth
	ctx.RemainingWidth = max(0, ctx.Bounds.Width-ctx.AllocatedWidth)
	if ctx.AllocatedWidth > ctx.MaxRowWidth {
		ctx.MaxRowWidth = ctx.AllocatedWidth
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
		BlockWidth:      bounds.Width,
		RemainingWidth:  bounds.Width,
		AllocatedWidth:  0,
		MaxRowWidth:     0,
		CurrentCol:      0,
		RowHeight:       0,
	}
	l.ctxStack = append(l.ctxStack, ctx)
}

// Add a clear, semantic wrapper for structural panels
func (l *Laid) PushGreedyContext(bounds rl.Rectangle) {
	if len(l.ctxStack) > 0 {
		parent := &l.ctxStack[len(l.ctxStack)-1]
		bounds.Height = parent.AvailableHeight
		if bounds.Width == 0 {
			bounds.Width = parent.RemainingWidth
		}
	}
	l.PushContext(bounds)
}

func (l *Laid) PopContext() {
	if len(l.ctxStack) < 2 {
		if len(l.ctxStack) == 1 {
			l.ctxStack = l.ctxStack[:0]
		}
		return
	}

	child := l.ctxStack[len(l.ctxStack)-1]
	l.ctxStack = l.ctxStack[:len(l.ctxStack)-1]

	parent := &l.ctxStack[len(l.ctxStack)-1]

	childHeightUsed := (child.CursorY + child.RowHeight) - child.Bounds.Y
	if childHeightUsed > parent.RowHeight {
		parent.RowHeight = childHeightUsed
	}

	if child.MaxRowWidth > parent.MaxRowWidth {
		parent.MaxRowWidth = child.MaxRowWidth
	}
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
	ctx.AllocatedWidth = 0
	ctx.RemainingWidth = ctx.Bounds.Width
}

func (l *Laid) Text(text string, size int32, color rl.Color) {
	b := l.Bounds()
	DrawText(text, int32(b.X), int32(b.Y-1), size, color)
	ctx := &l.ctxStack[len(l.ctxStack)-1]
	ctx.RowHeight = max(float32(size)+2, ctx.RowHeight)
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

// ComputeElementLayout recursively calculates and caches layout bounds for an Element tree
func (l *Laid) ComputeElementLayout(e *Element) rl.Rectangle {
	if e == nil || !e.Visible || e.Removed {
		return rl.Rectangle{}
	}

	l.Context() // Ensure active context on stack

	// Check if already cached and valid
	if cached, ok := l.GetCachedBounds(e.ID); ok {
		e.ComputedBounds = cached
		return cached
	}

	var bounds rl.Rectangle

	if e.IsAnchor {
		l.PushGreedyContext(e.Rectangle)
		l.SetBreakpoint(0)
		bounds = l.Bounds()
		defer l.PopContext()
	} else if e.Col != [4]int{} {
		var entered bool
		if bounds, entered = l.Col(e.Col[0], e.Col[1], e.Col[2], e.Col[3]); entered {
			l.PushContext(bounds)
			l.SetRowHeight(0)
			defer l.PopContext()
		}
	} else {
		bounds = l.Bounds()
	}

	if e.Rectangle.Height > 0 {
		l.SetRowHeight(e.Rectangle.Height)
	}
	l.Pad(float32(e.LeftPadding), float32(e.TopPadding))

	e.ComputedBounds = bounds
	l.SetCachedBounds(e.ID, bounds)

	for _, child := range e.Children {
		if child.Visible && !child.Removed {
			childBounds := l.ComputeElementLayout(child)
			if e.IsAnchor {
				l.Context().CursorX += childBounds.Width
			}
		}
	}

	return bounds
}

