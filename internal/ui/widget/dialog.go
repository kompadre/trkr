package widget

import (
	"fmt"
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
	return el
}

func (d Dialog) Show() {}
func (d Dialog) Hide() {}

func (d *Dialog) HandleInput(input *ev.InputSnapshot, el *ui.Element) bool {
	fmt.Printf("Receiving input from Dialog.\n")
	if input.Down(ev.InputKindEnter) {
		el.Remove()
		return true
	}
	return false
}

func (d *Dialog) Draw(ctx ev.EventContext, hasFocus bool, isHighlighted bool) bool {
	ww, wh := int32(ui.GetOptions().ScreenWidth), int32(ui.GetOptions().ScreenHeight)
	rl.DrawRectangle(0, 0, ww, 20, ui.WindowBg4)
	rl.DrawRectangle(0, 20, ww, wh-20, ui.WindowBg5)
	ui.DrawText(d.Label, 20, 0, 20, ui.WindowFg1)
	ui.DrawText(d.Text, 20, 30, 16, ui.WindowFg2)
	return true
}
