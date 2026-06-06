package ui

import rl "github.com/gen2brain/raylib-go/raylib"
import c "image/color"
import ev "trkr/internal/events"
import "fmt"
import "sync"
import "slices"

type Options struct {
	ColorHighlight  rl.Color
	RowHeight       int
	VerticalPadding int
	ScreenWidth     int
	ScreenHeight    int
}

var ElementTreeMutex sync.Mutex

var CurrentFrame uint32

type ElementCore interface {
	Show()
	Hide()
	HandleInput(ev.InputSnapshot, *Element) bool
	Draw(ev.EventContext, bool) bool
}

type ElementCoreInstance struct {
	ShowCallback        func()
	HideCallback        func()
	HandleInputCallback func(ev.InputSnapshot, *Element) bool
	DrawCallback        func(ev.EventContext, bool) bool
}

type ElementDrawPayload struct {
	Element *Element
	Left    int32
	Top     int32
}

func NewElementCoreInstance(show func(), hide func(), handleInput func(ev.InputSnapshot, *Element) bool, draw func(ev.EventContext, bool) bool) ElementCoreInstance {
	return ElementCoreInstance{ShowCallback: show, HideCallback: hide, HandleInputCallback: handleInput, DrawCallback: draw}
}

func (ec ElementCoreInstance) Show() {
	if ec.ShowCallback != nil {
		ec.ShowCallback()
	}
}

func (ec ElementCoreInstance) Hide() {
	if ec.HideCallback != nil {
		ec.HideCallback()
	}
}

func (ec ElementCoreInstance) HandleInput(input ev.InputSnapshot, el *Element) bool {
	if ec.HandleInputCallback != nil {
		return ec.HandleInputCallback(input, el)
	}
	return false
}

func (ec ElementCoreInstance) Draw(ctx ev.EventContext, hasFocus bool) bool {
	if ec.DrawCallback != nil {
		return ec.DrawCallback(ctx, hasFocus)
	}
	return false
}

type Element struct {
	ID           uint16
	Core         ElementCore
	Left         int32
	Top          int32
	LeftPadding  int32
	TopPadding   int32
	Width        int32
	Height       int32
	Parent       *Element
	Children     []*Element
	FocusedChild *Element
	Visible      bool
	Removed      bool
}

var RootElement *Element

func NewElement(left int32, top int32, width int32, height int32, core ElementCore, parent *Element) *Element {
	elem := &Element{ID: nextUniqueId(), Left: left, Top: top, Width: width, Height: height, Visible: true}
	if core != nil {
		elem.Core = core
	}
	if parent != nil {
		elem.Parent = parent
		parent.Add(elem, parent.FocusedChild == nil)
	}
	return elem
}

func (e *Element) HandleInput(input ev.InputSnapshot) bool {
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
		return // Nothing to focus
	}

	currentIndex := -1
	for i, child := range e.Children {
		if child == e.FocusedChild {
			currentIndex = i
			break
		}
	}
	if currentIndex == -1 {
		e.FocusedChild = e.Children[0]
		return
	}

	// Adding numChildren before modulo handles negative offsets safely
	newIndex := (currentIndex + offset) % numChildren
	if newIndex < 0 {
		newIndex += numChildren
	}

	// 3. Apply the focus change safely
	e.FocusedChild = e.Children[newIndex]
}

func (e *Element) HasFocus() bool {
	return e.Parent != nil && e.Parent.FocusedChild == e
}

func (e *Element) SetFocus(focusedElem *Element) {
	e.FocusedChild = focusedElem
}

func (e *Element) Draw(ctx ev.EventContext) bool {
	oldPayload := ElementDrawPayload{}
	if ctx.EventPayload != nil {
		oldPayload = *ctx.EventPayload.(*ElementDrawPayload)
		ctx.EventPayload.(*ElementDrawPayload).Element = e
		ctx.EventPayload.(*ElementDrawPayload).Top += e.Top
		ctx.EventPayload.(*ElementDrawPayload).Left += e.Left
		defer (func() {
			ctx.EventPayload.(*ElementDrawPayload).Left = oldPayload.Left
			ctx.EventPayload.(*ElementDrawPayload).Top = oldPayload.Top
			ctx.EventPayload.(*ElementDrawPayload).Element = oldPayload.Element
		})()
	}
	if e.Core != nil && e.Core.Draw(ctx, e.HasFocus()) {
		return true
	}
	if ctx.EventPayload != nil {
		ctx.EventPayload.(*ElementDrawPayload).Top += e.TopPadding
		ctx.EventPayload.(*ElementDrawPayload).Left += e.LeftPadding
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

func (e *Element) Add(c *Element, SetFocus bool) {
	fmt.Printf("Adding %v to %v.\n", c, e)
	e.Children = append(e.Children, c)
	if SetFocus {
		e.FocusedChild = c
	}
}

func (e *Element) Remove() {
	ev.RegisterCallback(ev.EventKindPostUpdate, func(ctx ev.EventContext) bool {
		if e.Removed {
			return true
		}

		for _, c := range e.Children {
			c.Remove()
		}
		if e.HasFocus() {
			e.Parent.FocusedChild = nil
		}
		e.Parent.Children = slices.DeleteFunc(e.Parent.Children, func(c *Element) bool {
			return c == e
		})
		e.Removed = true
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

var options *Options

var (
	WindowBg1 = c.RGBA{190, 190, 190, 255}
	WindowBg2 = c.RGBA{30, 190, 90, 255}
	WindowBg3 = c.RGBA{90, 190, 30, 255}
	WindowBg4 = c.RGBA{250, 250, 250, 255}

	InputBg1 = c.RGBA{255, 255, 255, 255}

	WindowFg1   = c.RGBA{30, 30, 30, 255}
	WindowFg2   = c.RGBA{220, 230, 210, 255}
	WindowFg3   = c.RGBA{250, 250, 250, 255}
	TrackerLine = c.RGBA{250, 190, 190, 100}

	Font rl.Font
)

func DrawText(text string, left int32, top int32, size int32, color c.RGBA) {
	rl.DrawTextEx(Font, text, rl.NewVector2(float32(left), float32(top)), float32(size), 1, color)
}

func GetSpareId() uint16 {
	lastUniqueId = lastUniqueId + 1
	return lastUniqueId
}

func GetOptions() *Options {
	if options == nil {
		options = &Options{
			ColorHighlight:  TrackerLine,
			RowHeight:       20,
			VerticalPadding: 10,
			ScreenWidth:     480,
			ScreenHeight:    320,
		}
	}
	return options
}

func SetOptions(opts Options) {
	options = &opts
}
