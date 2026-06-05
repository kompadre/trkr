package button

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	ev "trkr/internal/events"
	"trkr/internal/ui"
)

type Button struct {
	Label  string
	Action ui.Action
}

func NewButton(label string, relativeLeft int32, relativeTop int32, width int32, height int32, action ui.Action, parent *ui.Element) *Button {
	btn := Button{Label: label, Action: action}
	el := ui.NewElement(relativeLeft, relativeTop, width, height, btn, parent)
	el.Width = width
	el.Height = height
	return &btn
}

func (b Button) Show() {}

func (b Button) Hide() {}

func (b Button) HandleInput(input ev.InputSnapshot, el *ui.Element) bool {
	if b.Action == nil {
		return false
	}
	if input[ev.InputKindPressedEnter] {
		b.Action(b.Label)
		return true
	}
	return false
}

func (b Button) Draw(ctx ev.EventContext, hasFocus bool) bool {
	var left, top, width, height int32
	if ctx.EventPayload != nil {
		left = ctx.EventPayload.(*ui.ElementDrawPayload).Left +
			ctx.EventPayload.(*ui.ElementDrawPayload).Element.Left
		top = ctx.EventPayload.(*ui.ElementDrawPayload).Top +
			ctx.EventPayload.(*ui.ElementDrawPayload).Element.Top
		width = ctx.EventPayload.(*ui.ElementDrawPayload).Element.Width
		height = ctx.EventPayload.(*ui.ElementDrawPayload).Element.Height
	}
	bgcolor := ui.WindowBg2
	fgcolor := ui.WindowFg2
	labelPreffix := " "

	if hasFocus {
		bgcolor = ui.WindowBg3
		fgcolor = ui.WindowFg3
		labelPreffix = ">"
	}
	rl.DrawRectangle(int32(left), int32(top), width, height, bgcolor)
	rl.DrawText(labelPreffix+b.Label, int32(left+5), int32(top+5), 20, fgcolor)
	return false
}
