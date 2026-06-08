package settings

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	ev "trkr/internal/events"
	"trkr/internal/ui"
	"trkr/internal/ui/widget"
)

var Showing bool
var Logo rl.Texture2D

func Create(parent *ui.Element) {
	var uiElem *ui.Element
	core := ui.NewElementCoreInstance(Show, Hide, handleInputs, DrawSettings)
	uiElem = ui.NewElement(0, 0, 0, 0, core, parent)
	uiElem.Width = int32(ui.GetOptions().ScreenWidth)
	uiElem.Height = 110
	fmt.Printf("Creating settings: uiElem.ID is %d.\n", uiElem.ID)
	uiElem.TopPadding = 30
	uiElem.LeftPadding = 10
	uiElem.Visible = false
	fmt.Printf("Settings ID is %v.\n", uiElem.ID)
	_ = widget.NewInput("Name: ", 0, 0, 150, 26, nil, uiElem)
	_ = widget.NewButton("PROJ", 0, 20, 80, 30, (func(payload any) bool {

		return true
	}), uiElem)
	_ = widget.NewButton("TRAK", 45, 20, 80, 30, nil, uiElem)
	_ = widget.NewButton("SAVE", 90, 20, 80, 30, nil, uiElem)
	_ = widget.NewButton("PHRA", 135, 20, 80, 30, nil, uiElem)
	_ = widget.NewButton("[x]", 180, 20, 50, 30, nil, uiElem)

	ui.SettingsDialog = uiElem

	image := rl.LoadImage("./assets/images/logo.png")
	Logo = rl.LoadTextureFromImage(image)
	rl.UnloadImage(image)
}

func Hide() {
	ui.SettingsDialog.Visible = false
}

func Show() {
	if ui.SettingsDialog.Visible == false {
		fmt.Printf("Setting uiElem.Visible to true. ID is %d.\n", ui.SettingsDialog.ID)
		ui.SettingsDialog.Visible = true
		ui.SettingsDialog.Parent.SetFocus(ui.SettingsDialog)
	}
}

func DrawSettings(ctx ev.EventContext, hasFocus bool) bool {
	p := ctx.EventPayload.(*ui.ElementDrawPayload)
	rl.DrawRectangle(p.Left, p.Top, p.Element.Width, p.Element.Height, ui.WindowBg1)
	ui.DrawText("Project", p.Left+10, p.Top, 20, ui.WindowFg1)
	//	DrawTextureRec(Texture2D texture, Rectangle source, Vector2 position, Color tint)
	rl.DrawTextureRec(Logo, rl.Rectangle{0, 0, 159, 69}, rl.Vector2{float32(ui.GetOptions().ScreenWidth - 159), 0}, rl.White)
	return false
}

func handleInputs(input ev.InputSnapshot, el *ui.Element) bool {
	var result bool
	if input[ev.InputKindPressedEnter] {
		fmt.Printf("Hiding...")
		ui.SettingsDialog.Parent.FocusJump(1)
		Hide()
		result = true
	} else if input[ev.InputKindPressedRight] {
		ui.SettingsDialog.FocusJump(1)
		result = true
	} else if input[ev.InputKindPressedLeft] {
		ui.SettingsDialog.FocusJump(-1)
		result = true
	}
	return result
}
