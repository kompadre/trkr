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
	Draw(ev.EventContext, bool) bool
}

type ElementCoreInstance struct {
	ShowCallback        func()
	HideCallback        func()
	HandleInputCallback func(*ev.InputSnapshot, *Element) bool
	DrawCallback        func(ev.EventContext, bool) bool
}

type ElementDrawPayload struct {
	Element *Element
	Laid    *Laid
	Left    int32
	Top     int32
}

func NewElementCoreInstance(show func(), hide func(), handleInput func(*ev.InputSnapshot, *Element) bool, draw func(ev.EventContext, bool) bool) *ElementCoreInstance {
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

func (ec *ElementCoreInstance) Draw(ctx ev.EventContext, hasFocus bool) bool {
	if ec.DrawCallback != nil {
		return ec.DrawCallback(ctx, hasFocus)
	}
	return false
}

type Element struct {
	ID           uint16
	Core         ElementCore
	Rectangle    rl.Rectangle
	IsAnchor     bool
	LeftPadding  int32
	TopPadding   int32
	Parent       *Element
	Children     []*Element
	FocusedChild *Element
	Visible      bool
	Removed      bool
}

var RootElement *Element
var TrackDialog *Element
var SongDialog *Element
var SettingsDialog *Element
var SettingsProject *Element

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

func (e *Element) HandleInput(input *ev.InputSnapshot) bool {
	if e.FocusedChild != nil {
		if e.FocusedChild.HandleInput(input) {
			return true
		}
	}
	if e.Core == nil {
		return false
	}
	return e.Core.HandleInput(input, e)
}

func (e *Element) FocusJump(offset int) {
	numChildren := len(e.Children)
	if numChildren == 0 {
		e.FocusedChild = nil
		return
	}

	currentIndex := -1
	hasVisibleChildren := false

	for i, child := range e.Children {
		if child == e.FocusedChild {
			currentIndex = i
		}
		if child.Visible {
			hasVisibleChildren = true
		}
		if hasVisibleChildren && currentIndex > -1 {
			break
		}
	}

	// If all children are invisible, clear focus immediately and bail
	if !hasVisibleChildren {
		fmt.Println("Container has no visible children. Clearing focus.")
		e.FocusedChild = nil
		return
	}

	if currentIndex == -1 {
		currentIndex = 0
	}

	newIndex := (currentIndex + offset) % numChildren
	if newIndex < 0 {
		newIndex += numChildren
	}

	for !e.Children[newIndex].Visible {
		newIndex = (newIndex + 1) % numChildren
	}

	fmt.Printf("Focusing on child %d (ID:%d).\n", newIndex, e.Children[newIndex].ID)
	e.FocusedChild = e.Children[newIndex]
}

func (e *Element) HasFocus() bool {
	return e.Parent != nil && e.Parent.FocusedChild == e
}

func (e *Element) SetFocus(focusedElem *Element) {
	e.FocusedChild = focusedElem
}

func (e *Element) Draw(ctx ev.EventContext) bool {
	if e.IsAnchor {
		ctx.EventPayload.(*ElementDrawPayload).Laid.PushContext(e.Rectangle)
		ctx.EventPayload.(*ElementDrawPayload).Laid.SetRowHeight(e.Rectangle.Height)
		defer func() {
			ctx.EventPayload.(*ElementDrawPayload).Laid.PopContext()
		}()
	}
	if e.Core != nil && e.Core.Draw(ctx, e.HasFocus()) {
		return true
	}

	for i := range e.Children {
		if e.Children[i].Visible {
			ctx.EventPayload.(*ElementDrawPayload).Element = e.Children[i]
			if e.Children[i].Draw(ctx) {
				return true
			}
		}
	}

	return false
}

func (el *Element) DrawContainer(ctx ev.EventContext) {
	ctxPay := ctx.EventPayload.(*ElementDrawPayload)
	// rect := rl.NewRectangle(float32(ctxPay.Left), float32(ctxPay.Top), float32(el.Width-ctxPay.Left*2), float32(el.Height))
	rl.DrawRectangleLinesEx(ctxPay.Laid.Bounds(), 2, rl.Black)
}

func (e *Element) Add(c *Element, SetFocus bool) {
	fmt.Printf("Adding %v to %v.\n", c, e)
	e.Children = append(e.Children, c)
	if SetFocus {
		e.FocusedChild = c
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
