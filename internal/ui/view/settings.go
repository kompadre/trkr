package view

import (
	"fmt"
	"strconv"
	. "trkr"
	ev "trkr/internal/events"
	"trkr/internal/ui"
	// "trkr/internal/ui/view/settings"
	"trkr/internal/ui/widget"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var Showing bool
var Logo rl.Texture2D

func CreateSettings(parent *ui.Element) {
	var uiElem *ui.Element
	core := ui.NewElementCoreInstance(showSettings, hideSettings, settingsHandleInputs, drawSettings)
	uiElem = ui.NewElement(0, 0, int32(ui.GetOptions().ScreenWidth), 300, core, parent)
	uiElem.TopPadding = 30
	uiElem.LeftPadding = 10
	uiElem.Visible = false
	uiElem.IsAnchor = true
	projectName := widget.NewInput("Name: ", 0, 0, 150, 26, func(a any) bool {
		Logf("Setting project filename to %s.\n", a)
		CurrentProject.Filename = a.(string) + ".json"
		return true
	}, uiElem)
	projectName.SetValue(CurrentProject.Filename)
	bpm := widget.NewInput("BPM: ", 80, 0, 120, 26, nil, uiElem)
	bpm.WidgetType = widget.WidgetInputTypeNumber
	bpm.SetValue(strconv.Itoa(BeatsPerMinute))

	//	sampleName := widget.NewInput("Sample: ", 0, 20, 150, 26, nil, uiElem)
	//	sampleName.SetValue(CurrentProject.Current().Instrument.SampleSourceType.UiString())
	trackSkips := widget.NewInput("T.Skips: ", 80, 0, 120, 26, nil, uiElem)
	trackSkips.WidgetType = widget.WidgetInputTypeNumber
	_ = widget.NewButton("PROJ", 0, 40, 80, 30, (func(payload any) bool {

		return true
	}), uiElem)
	_ = widget.NewButton("TRAK", 45, 40, 80, 30, nil, uiElem)
	_ = widget.NewButton("SAVE", 90, 40, 80, 30, nil, uiElem)
	_ = widget.NewButton("PHRA", 135, 40, 80, 30, nil, uiElem)
	_ = widget.NewButton("[x]", 180, 40, 50, 30, nil, uiElem)

	//	ui.SettingsProject = settings.CreateProjectDialog(uiElem)
	ui.SettingsDialog = uiElem

	image := rl.LoadImage("./assets/images/logo.png")
	Logo = rl.LoadTextureFromImage(image)
	rl.UnloadImage(image)
}

func hideSettings() {
	ui.SettingsDialog.Visible = false
}

func showSettings() {
	if ui.SettingsDialog.Visible == false {
		fmt.Printf("Setting uiElem.Visible to true. ID is %d.\n", ui.SettingsDialog.ID)
		ui.SettingsDialog.Visible = true
		ui.SettingsDialog.Parent.SetFocus(ui.SettingsDialog)
		ui.SettingsDialog.SetFocus(ui.SettingsProject)
	}
}

func drawSettings(ctx ev.EventContext, hasFocus bool) bool {
	p := ctx.EventPayload.(*ui.ElementDrawPayload)
	if rec, v := p.Laid.Col(12, 12, 6, 6); v {
		rl.DrawRectangleRec(rec, RGBA(0xff, 0xff, 0xff, 0xa0))
		ui.DrawText("Project", int32(rec.X)+10, int32(rec.Y), 20, ui.WindowFg1)
		rec.X += 10
		rec.Y += 20
		p.Laid.PushContext(rec)

		// rl.DrawTextureRec(Logo, rl.Rectangle{0, 0, 159, 69}, rl.Vector2{float32(ui.GetOptions().ScreenWidth - 159), 0}, rl.White)
	}
	// DrawTextureRec(Texture2D texture, Rectangle source, Vector2 position, Color tint)
	return false
}

func settingsHandleInputs(input *ev.InputSnapshot, el *ui.Element) bool {
	var result bool
	if input.Down(ev.InputKindEnter) || input.Tick(ev.InputKindB) > 60 {
		fmt.Printf("Hiding...")
		ui.SettingsDialog.Parent.FocusJump(1)
		hideSettings()
		result = true
	} else if input.Down(ev.InputKindRight) || input.Down(ev.InputKindDown) {
		ui.SettingsDialog.FocusJump(1)
		result = true
	} else if input.Down(ev.InputKindLeft) || input.Down(ev.InputKindUp) {
		ui.SettingsDialog.FocusJump(-1)
		result = true
	}
	return result
}
