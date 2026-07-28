package ui

import (
	"fmt"
	"slices"
	"sync"
	ev "trkr/internal/events"

	rl "github.com/gen2brain/raylib-go/raylib"

	. "trkr"
)

const (
	OptionsScreenWidth  = 480
	OptionsScreenHeight = 320
)

type Options struct {
	ColorHighlight  rl.Color
	RowHeight       int
	VerticalPadding int
	ScreenWidth     int
	ScreenHeight    int
}

var ElementTreeMutex sync.Mutex

var CurrentFrame uint32

var TrackId int
var SectionId int
var PhraseId int

type ElementCore interface {
	Show()
	Hide()
	HandleInput(*ev.InputSnapshot, *Element) bool
	Draw(ev.EventContext, bool, bool) bool
}

type ElementCoreInstance struct {
	ShowCallback        func()
	HideCallback        func()
	HandleInputCallback func(*ev.InputSnapshot, *Element) bool
	DrawCallback        func(ev.EventContext, bool, bool) bool
}

type ElementDrawPayload struct {
	Element *Element
	Laid    *Laid
	Left    int32
	Top     int32
}

func NewElementCoreInstance(show func(), hide func(), handleInput func(*ev.InputSnapshot, *Element) bool, draw func(ev.EventContext, bool, bool) bool) *ElementCoreInstance {
	return &ElementCoreInstance{ShowCallback: show, HideCallback: hide, HandleInputCallback: handleInput, DrawCallback: draw}
}

func (ec *ElementCoreInstance) Show() {
	if ec.ShowCallback != nil {
		ec.ShowCallback()
	}
}

func (ec *ElementCoreInstance) Hide() {
	if ec.HideCallback != nil {
		ec.HideCallback()
	}
}

func (ec *ElementCoreInstance) HandleInput(input *ev.InputSnapshot, el *Element) bool {
	if ec.HandleInputCallback != nil {
		return ec.HandleInputCallback(input, el)
	}
	return false
}

func (ec *ElementCoreInstance) Draw(ctx ev.EventContext, hasFocus bool, isHighlighted bool) bool {
	if ec.DrawCallback != nil {
		return ec.DrawCallback(ctx, hasFocus, isHighlighted)
	}
	return false
}

type Element struct {
	ID                uint16
	Name              string
	Core              ElementCore
	Rectangle         rl.Rectangle
	ComputedBounds    rl.Rectangle
	IsAnchor          bool
	LeftPadding       int32
	TopPadding        int32
	Parent            *Element
	Children          []*Element
	FocusedChild      *Element
	HighlightedChild  *Element
	FocusOutAfterLast bool
	Col               [4]int
	Visible           bool
	IsModal           bool
	Removed           bool
}

func (e *Element) Bounds() rl.Rectangle {
	if e.ComputedBounds.Width > 0 || e.ComputedBounds.Height > 0 {
		return e.ComputedBounds
	}
	return e.Rectangle
}

var RootElement *Element
var TrackDialog *Element
var SongDialog *Element
var SettingsDialog *Element
var InstrumentDialog *Element
var PatchBrowserDialog *Element

func NewElement(left int32, top int32, width int32, height int32, core ElementCore, parent *Element) *Element {
	rect := rl.Rectangle{X: float32(left), Y: float32(top), Width: float32(width), Height: float32(height)}
	elem := &Element{ID: nextUniqueId(), Rectangle: rect, Visible: true}
	if core != nil {
		elem.Core = core
	}
	if parent != nil {
		elem.Parent = parent
		parent.Add(elem, parent.FocusedChild == nil)
	}
	return elem
}

func NewRow(height int32, parent *Element) *Element {
	e := NewElement(0, 0, 0, height, nil, parent)
	e.Col = [4]int{12, 12, 12, 12}
	e.FocusOutAfterLast = true
	return e
}

func (e *Element) HandleInput(input *ev.InputSnapshot) bool {
	// 0. Priority: Modal children capture all input for their parent's entire subtree
	for _, child := range e.Children {
		if child.Visible && child.IsModal {
			return child.HandleInput(input)
		}
	}

	// 1. If we have an active (focused) child, it gets priority
	if e.FocusedChild != nil && e.FocusedChild.Visible {
		if e.FocusedChild.HandleInput(input) {
			return true
		}
		// If child returned false, it means it wants to be deactivated.
		// Top-level views focused on RootElement should never be automatically deactivated.
		if e != RootElement {
			e.FocusedChild = nil
			return true
		}
		return false
	}

	// 2. Leaf Core Elements: Widgets and monolithic views that handle everything in their Core.
	// They have no children, so they do not participate in container-level navigation.
	if e.Core != nil && len(e.Children) == 0 {
		return e.Core.HandleInput(input, e)
	}

	// 3. Container Elements: They have children, so their primary duty is navigation and focusing.
	// If a custom Core handler is present on a container, let it intercept global events first.
	if e.Core != nil {
		if e.Core.HandleInput(input, e) {
			return true
		}
	}

	// 3.5. Delegate input down to the highlighted child if it is also a container.
	// This enables seamless navigation inside nested layout rows/containers.
	if e.HighlightedChild != nil && e.HighlightedChild.Visible && len(e.HighlightedChild.Children) > 0 {
		if e.HighlightedChild.HandleInput(input) {
			return true
		}
	}

	// 4. Container-level navigation & activation fallback
	if input.Tick(ev.InputKindA) == 1 {
		if e.HighlightedChild != nil && e.HighlightedChild.Visible {
			e.FocusedChild = e.HighlightedChild
			if e.Parent != nil {
				e.Parent.FocusedChild = e
			}
			return true
		}
	}

	if input.Tick(ev.InputKindB) == 1 {
		return false // Signals parent to deactivate us
	}

	if input.Down(ev.InputKindDown) || input.Down(ev.InputKindRight) {
		e.HighlightJump(1)
		return true
	} else if input.Down(ev.InputKindUp) || input.Down(ev.InputKindLeft) {
		e.HighlightJump(-1)
		return true
	}

	return false
}

func (e *Element) HighlightJump(offset int) {
	numChildren := len(e.Children)
	if numChildren == 0 {
		e.HighlightedChild = nil
		return
	}
	if !e.Visible && e.IsHighlighted() && e.Parent != nil {
		e.Parent.HighlightJump(0)
		return
	}

	currentIndex := -1
	lastVisibleChildIdx := -1
	numVisibleChildren := 0
	var firstVisibileChild *Element
	for i, child := range e.Children {
		if child == e.HighlightedChild {
			currentIndex = i
		}
		if child.Visible {
			if firstVisibileChild == nil {
				firstVisibileChild = child
			}
			numVisibleChildren++
			lastVisibleChildIdx = i
		}
	}

	// If all children are invisible, clear highlight immediately and bail
	if numVisibleChildren == 0 {
		fmt.Println("Container has no visible children. Clearing highlight.")
		e.HighlightedChild = nil
		return
	}

	if offset == 1 && lastVisibleChildIdx == currentIndex && e.FocusOutAfterLast && e.Parent != nil {
		// If we ask for the next child but we already at the last visible one.
		e.HighlightedChild = firstVisibileChild
		e.Parent.HighlightJump(1)
		return
	}

	if offset == -1 && (currentIndex == -1 || e.Children[currentIndex] == firstVisibileChild) && e.FocusOutAfterLast && e.Parent != nil {
		// If we ask for the previous child but we already at the first visible one.
		e.Parent.HighlightJump(-1)
		return
	}

	if numVisibleChildren == 1 { // Is the only visible child
		e.HighlightedChild = e.Children[lastVisibleChildIdx]
		return
	}

	if currentIndex == -1 {
		currentIndex = 0
	}

	newIndex := (currentIndex + offset) % numChildren
	if newIndex < 0 {
		newIndex += numChildren
	}

	searchDir := offset
	if searchDir == 0 {
		searchDir = 1
	}

	for !e.Children[newIndex].Visible {
		newIndex = (newIndex + searchDir) % numChildren
		if newIndex < 0 {
			newIndex += numChildren
		}
	}

	fmt.Printf("Highlighting child %d (ID:%d).\n", newIndex, e.Children[newIndex].ID)
	e.HighlightedChild = e.Children[newIndex]
}

func (e *Element) IsHighlighted() bool {
	return e.Parent != nil && e.Parent.HighlightedChild == e
}

func (e *Element) HasFocus() bool {
	return e.Parent != nil && e.Parent.FocusedChild == e
}

func (e *Element) SetFocus(focusedElem *Element) {
	e.FocusedChild = focusedElem
	e.HighlightedChild = focusedElem
}

func (e *Element) Draw(ctx ev.EventContext, hasFocus bool, isHighlighted bool) bool {
	laid := ctx.EventPayload.(*ElementDrawPayload).Laid
	if e.IsAnchor {
		laid.PushGreedyContext(e.Rectangle)
		laid.SetBreakpoint(0)
		defer func() {
			laid.PopContext()
			// rl.DrawRectangleLinesEx(laid.Bounds(), 3, rl.Blue)
		}()
	} else if e.Col != [4]int{} {
		if !laid.EnterCol(e.Col[0], e.Col[1], e.Col[2], e.Col[3]) {
			return true
		}
		defer func() {
			laid.ExitCol()
		}()
	}

	e.ComputedBounds = laid.Bounds()
	if e.ID > 0 {
		laid.SetCachedBounds(e.ID, e.ComputedBounds)
	}

	if !e.IsAnchor {
		laid.SetRowHeight(e.Rectangle.Height)
	}
	laid.Pad(float32(e.LeftPadding), float32(e.TopPadding))

	// Priority: If any child is modal and visible, it's the ONLY thing we draw
	for _, child := range e.Children {
		if child.Visible && child.IsModal {
			ctx.EventPayload.(*ElementDrawPayload).Element = child
			return child.Draw(ctx, hasFocus && child == e.FocusedChild, isHighlighted && child == e.HighlightedChild)
		}
	}

	if e.Core != nil && e.Core.Draw(ctx, hasFocus, isHighlighted) {
		return true
	}
	//	laid.Context().CursorY += e.Rectangle.Height

	for i := range e.Children {
		if e.Children[i].Visible {
			ctx.EventPayload.(*ElementDrawPayload).Element = e.Children[i]
			if e.Children[i].Draw(ctx, hasFocus && e.Children[i] == e.FocusedChild, isHighlighted && e.Children[i] == e.HighlightedChild) {
				return true
			}
			if e.IsAnchor {
				laid.Context().CursorX += e.Children[i].Rectangle.Width
			}
		}
	}

	return false
}

func (el *Element) DrawContainer(ctx ev.EventContext) {
	ctxPay := ctx.EventPayload.(*ElementDrawPayload)
	rec := ctxPay.Laid.Bounds()
	if rec.Height == 0 {
		rec.Height = 20
	}
	Logf("Paiting element's container: %v.\n", rec)
	// rect := rl.NewRectangle(float32(ctxPay.Left), float32(ctxPay.Top), float32(el.Width-ctxPay.Left*2), float32(el.Height))
	rl.DrawRectangleLinesEx(rec, 2, rl.Black)
}

func (e *Element) Add(c *Element, setInitialFocus bool) {
	fmt.Printf("Adding %v to %v.\n", c, e)
	e.Children = append(e.Children, c)
	if setInitialFocus {
		e.HighlightedChild = c
	}
}

func (e *Element) Remove() {
	ev.RegisterCallback(ev.EventKindPostUpdate, func(ctx ev.EventContext) bool {
		if e.HasFocus() {
			e.Parent.FocusedChild = nil
		}
		e.Parent.Children = slices.DeleteFunc(e.Parent.Children, func(c *Element) bool {
			return c == e
		})

		// 2. Define the local recursive helper function
		var burnSubtree func(target *Element) // Declared first so it can self-reference
		burnSubtree = func(target *Element) {
			if target == nil || target.Removed {
				return
			}

			// Trigger custom component cleanups (Cgo, etc.)
			if cleanable, ok := any(target).(Cleanable); ok {
				cleanable.Cleanup()
			}

			// Recursively cascade down the children
			for _, c := range target.Children {
				burnSubtree(c)
			}

			target.Removed = true
		}

		// 3. Kick off the localized destruction
		burnSubtree(e)

		return true
	}, e.ID)
}

type Action func(any) bool

var lastUniqueId uint16

func nextUniqueId() uint16 {
	lastUniqueId = lastUniqueId + 1
	return lastUniqueId
}

var mapElementToId map[uint16]*Element

var options = Options{
	ColorHighlight:  TrackerLine,
	RowHeight:       20,
	VerticalPadding: 10,
	ScreenWidth:     Clamp(OptionsScreenWidth, 480, 1920),
	ScreenHeight:    Clamp(OptionsScreenHeight, 320, 1280),
}

var (
	WindowBg1 = RGBA(190, 190, 190, 255)
	WindowBg2 = RGBA(30, 190, 90, 255)
	WindowBg3 = RGBA(90, 190, 30, 255)
	WindowBg4 = RGBA(250, 250, 250, 255)
	WindowBg5 = RGBA(30, 30, 30, 255)

	InputBg1 = RGBA(255, 255, 255, 255)

	WindowFg1   = RGBA(30, 30, 30, 255)
	WindowFg2   = RGBA(220, 230, 210, 255)
	WindowFg3   = RGBA(250, 250, 250, 255)
	TrackerLine = RGBA(128, 128, 170, 60)

	Font rl.Font
)

var textBatcher = TextBatcher{}

func DrawText(text string, left int32, top int32, size int32, color rl.Color) {
	// rl.DrawTextEx(Font, text, rl.NewVector2(float32(left), float32(top)), float32(size), 1, color)
	textBatcher.Add(text, float32(left), float32(top), float32(size), color)
}

func FlushText() {
	textBatcher.Flush(Font)
}

func GetSpareId() uint16 {
	lastUniqueId = lastUniqueId + 1
	return lastUniqueId
}

func GetOptions() *Options {
	return &options
}

func SetOptions(opts Options) {
	options = opts
}
