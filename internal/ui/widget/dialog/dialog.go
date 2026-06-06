package dialog

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	ev "trkr/internal/events"
	"trkr/internal/ui"
)

type Dialog struct {
	Label  string
	Text   string
	Action ui.Action
	Ttl    int32
}

func NewDialog(label string, text string, relativeLeft int32, relativeTop int32, width int32, height int32, action ui.Action, parent *ui.Element) *ui.Element {
	d := &Dialog{Label: label, Action: action, Text: text}
	el := ui.NewElement(relativeLeft, relativeTop, width, height, d, parent)
	el.Width = width
	el.Height = height
	return el
}

func (d Dialog) Show() {}
func (d Dialog) Hide() {}

func (d *Dialog) HandleInput(input ev.InputSnapshot, el *ui.Element) bool { return false }

func (d *Dialog) Draw(ctx ev.EventContext, hasFocus bool) bool {
	bgcolor := ui.WindowBg2

	rl.DrawRectangle(0, 0, int32(ui.GetOptions().ScreenWidth), int32(ui.GetOptions().ScreenHeight), bgcolor)
	ui.DrawText(d.Label, 4, 0, 16, ui.WindowFg1)
	ui.DrawText(d.Text, 4, 20, 20, ui.WindowFg2)
	return true
}
