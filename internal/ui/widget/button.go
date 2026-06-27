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

func NewButton(label string, relativeLeft int32, relativeTop int32, width int32, height int32, action ui.Action, parent *ui.Element) *Button {
	btn := Button{Label: label, Action: action}
	el := ui.NewElement(relativeLeft, relativeTop, width, height, btn, parent)
	el.Core = &btn
	return &btn
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
	switch t := ctx.EventPayload.(type) {
	case *ui.ElementDrawPayload:
		if container, v := t.Laid.Col(12, 12, 1, 1); v {
			bgcolor := RGBA(30, 250, 30, 250)
			// bgcolor := ui.WindowBg2
			fgcolor := ui.WindowFg2
			labelPreffix := " "

			if hasFocus {
				bgcolor = ui.WindowBg3
				fgcolor = ui.WindowFg3
				labelPreffix = ">"
			}
			rl.DrawRectangleRec(container, bgcolor)
			rl.DrawText(labelPreffix+b.Label, int32(container.X), int32(container.Y), 20, fgcolor)
		}
	}
	return false
}
