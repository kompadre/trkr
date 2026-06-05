package settings

import (
	"fmt"
	ev "trkr/internal/events"
	"trkr/internal/ui"
	"trkr/internal/ui/widget/button"
	"trkr/internal/ui/widget/input"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var Showing bool

var btnTrackerTexture rl.Texture2D

var uiElem *ui.Element

func Create(parent *ui.Element) {
	core := ui.NewElementCoreInstance(Show, Hide, handleInputs, DrawSettings)
	uiElem = ui.NewElement(0, 0, 0, 0, core, parent)
	uiElem.Width = int32(ui.GetOptions().ScreenWidth)
	uiElem.Height = 110
	fmt.Printf("Creating settings: uiElem.ID is %d.\n", uiElem.ID)
	uiElem.TopPadding = 30
	uiElem.LeftPadding = 10
	uiElem.Visible = false
	fmt.Printf("Settings ID is %v.\n", uiElem.ID)
	_ = input.NewInput("Name: ", 0, 0, 200, 26, nil, uiElem)
	_ = button.NewButton("SONG", 0, 20, 80, 30, (func(any) bool {
		return true
	}), uiElem)
	_ = button.NewButton("TRAK", 45, 20, 80, 30, nil, uiElem)
	_ = button.NewButton("SAVE", 90, 20, 80, 30, nil, uiElem)
	_ = button.NewButton("PHRA", 135, 20, 80, 30, nil, uiElem)
	_ = button.NewButton("[x]", 180, 20, 50, 30, nil, uiElem)
}

func Hide() {
	uiElem.Visible = false
}

func Show() {
	fmt.Printf("Setting uiElem.Visible to true. ID is %d.\n", uiElem.ID)
	uiElem.Visible = true
	uiElem.Parent.SetFocus(uiElem)
}

func DrawSettings(ctx ev.EventContext, hasFocus bool) bool {
	p := ctx.EventPayload.(*ui.ElementDrawPayload)
	rl.DrawRectangle(p.Left, p.Top, p.Element.Width, p.Element.Height, ui.WindowBg1)
	ui.DrawText("Settings", p.Left, p.Top, 20, ui.WindowFg1)
	return false
}

func handleInputs(input ev.InputSnapshot, el *ui.Element) bool {
	var result bool
	if input[ev.InputKindPressedEnter] {
		uiElem.Parent.FocusJump(1)
		Hide()
		result = true
	} else if input[ev.InputKindPressedRight] {
		uiElem.FocusJump(1)
		result = true
	} else if input[ev.InputKindPressedLeft] {
		uiElem.FocusJump(-1)
		result = true
	}
	return result
}
