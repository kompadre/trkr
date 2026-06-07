package settings

import (
	"fmt"
	"time"
	ev "trkr/internal/events"
	"trkr/internal/ui"
	"trkr/internal/ui/widget/button"
	"trkr/internal/ui/widget/dialog"
	"trkr/internal/ui/widget/input"

	rl "github.com/gen2brain/raylib-go/raylib"
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
	_ = input.NewInput("Name: ", 0, 0, 150, 26, nil, uiElem)
	_ = button.NewButton("SONG", 0, 20, 80, 30, (func(payload any) bool {
		bel := payload.(*ui.Element)
		d := dialog.NewDialog("Title", "Some text here won't harm right?", 50, 20, 150, 50, nil, parent)
		d.Visible = true
		oldRoot := ui.RootElement
		ui.RootElement = d
		time.AfterFunc(5*time.Second, func() {
			fmt.Printf("Removing dialog...\n")
			bel.Parent.FocusedChild = nil
			ui.RootElement = oldRoot
			d.Remove()
		})
		return true
	}), uiElem)
	_ = button.NewButton("TRAK", 45, 20, 80, 30, nil, uiElem)
	_ = button.NewButton("SAVE", 90, 20, 80, 30, nil, uiElem)
	_ = button.NewButton("PHRA", 135, 20, 80, 30, nil, uiElem)
	_ = button.NewButton("[x]", 180, 20, 50, 30, nil, uiElem)

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
	ui.DrawText("Settings", p.Left+10, p.Top, 20, ui.WindowFg1)
	//	DrawTextureRec(Texture2D texture, Rectangle source, Vector2 position, Color tint)
	rl.DrawTextureRec(Logo, rl.Rectangle{0, 0, 160, 72}, rl.Vector2{float32(ui.GetOptions().ScreenWidth - 160), 0}, rl.White)
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
