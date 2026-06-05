package dialog

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	ev "trkr/internal/events"
	"trkr/internal/ui"
)

type Dialog struct {
	Title  string
	Label  string
	Action ui.Action
	Ttl    int32
}

func NewDialog(label string, relativeLeft int32, relativeTop int32, width int32, height int32, action ui.Action, parent *ui.Element) *Dialog {
	d := &Dialog{Label: label, Action: action}
	el := ui.NewElement(relativeLeft, relativeTop, width, height, d, parent)
	el.Width = width
	el.Height = height
	ev.RegisterCallback(ev.EventKindGuiDraw, func(ctx ev.EventContext) bool {
		d.Draw(ctx, true)
		return true
	}, el.ID)
	return d
}

func (d Dialog) Show() {}

func (d Dialog) Hide() {}

func (d Dialog) HandleInput(input ev.InputSnapshot, el *ui.Element) bool { return false }

func (d *Dialog) Draw(ctx ev.EventContext, hasFocus bool) bool {
	bgcolor := ui.WindowBg2
	fgcolor := ui.WindowFg2
	labelPreffix := " "

	rl.DrawRectangle(0, 0, int32(ui.GetOptions().ScreenWidth), int32(ui.GetOptions().ScreenHeight), bgcolor)
	rl.DrawText(d.Label, 10, 10, 20, fgcolor)
	return true
}
