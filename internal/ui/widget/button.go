package widget

import (
	. "trkr"
	ev "trkr/internal/events"
	"trkr/internal/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Button struct {
	Label  string
	Action ui.Action
}

func NewButton(label string, col [4]int, height int32, action ui.Action, parent *ui.Element) *ui.Element {
	btn := Button{Label: label, Action: action}
	el := ui.NewElement(0, 0, 0, height, btn, parent)
	el.Col = col
	el.Core = &btn
	return el
}

func (b Button) Show() {}

func (b Button) Hide() {}

func (b Button) HandleInput(input *ev.InputSnapshot, el *ui.Element) bool {
	if b.Action == nil {
		return false
	}
	if input.Down(ev.InputKindEnter) {
		b.Action(el)
		return true
	}
	return false
}

func (b Button) Draw(ctx ev.EventContext, hasFocus bool) bool {
	laid := ctx.EventPayload.(*ui.ElementDrawPayload).Laid
	container := laid.Bounds()
	bgcolor := RGBA(30, 250, 30, 250)
	fgcolor := ui.WindowFg2
	labelPreffix := " "
	if hasFocus {
		bgcolor = ui.WindowBg3
		fgcolor = ui.WindowFg3
		labelPreffix = ">"
	}
	rl.DrawRectangleRec(container, bgcolor)
	// rl.DrawRectangleLinesEx(container, 3, rl.Red)
	rl.DrawText(labelPreffix+b.Label, int32(container.X+10), int32(container.Y+10), 20, fgcolor)
	laid.SetRowHeight(30)
	return false
}
