package view

import (
	"fmt"
	"os"
	"strconv"
	. "trkr"
	ev "trkr/internal/events"
	"trkr/internal/player"
	"trkr/internal/ui"
	"trkr/internal/ui/widget"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var Showing bool
var Logo rl.Texture2D

func CreateSettings(parent *ui.Element) {
	var uiElem *ui.Element
	core := ui.NewElementCoreInstance(showSettings, hideSettings, settingsHandleInputs, drawSettings)
	uiElem = ui.NewElement(0, 0, int32(ui.GetOptions().ScreenWidth), 300, core, parent)
	uiElem.Name = "Settings"
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

	buttonsRow := ui.NewElement(0, 0, int32(rl.GetScreenWidth()), 40, nil, uiElem)
	buttonsRow.Col = [4]int{10, 10, 10, 10}
	buttonsRow.TopPadding = 10
	buttonsRow.FocusOutAfterLast = true
	buttonCol := [4]int{3, 3, 3, 3}
	widget.NewButton("BACK", buttonCol, 30, func(_ any) bool {
		uiElem.Visible = false
		return true
	}, buttonsRow)
	widget.NewButton("WAV", buttonCol, 30, func(_ any) bool {
		go func() {
			err := player.ExportWav("export.wav")
			if err != nil {
				fmt.Printf("Export error: %v\n", err)
			}
		}()
		return true
	}, buttonsRow)
	widget.NewButton("SAVE", buttonCol, 30, func(_ any) bool {
		err := SaveProject()
		if err != nil {
			fmt.Printf("Save error: %v\n", err)
		}
		return true
	}, buttonsRow)
	widget.NewButton("QUIT", buttonCol, 30, func(_ any) bool {
		os.Exit(0)
		return true
	}, buttonsRow)

	ui.SettingsDialog = uiElem

	image := rl.LoadImage("./assets/images/logo.png")
	Logo = rl.LoadTextureFromImage(image)
	rl.UnloadImage(image)
}

func hideSettings() {
	ui.SettingsDialog.Visible = false
	ui.TrackDialog.Visible = true
	ui.SettingsDialog.Parent.SetFocus(ui.TrackDialog)
}

func showSettings() {
	if ui.SettingsDialog.Visible == false {
		fmt.Printf("Setting uiElem.Visible to true. ID is %d.\n", ui.SettingsDialog.ID)
		ui.SettingsDialog.Visible = true
		ui.SettingsDialog.Parent.SetFocus(ui.SettingsDialog)
	}
}

func drawSettings(ctx ev.EventContext, hasFocus bool, isHighlighted bool) bool {
	p := ctx.EventPayload.(*ui.ElementDrawPayload)
	if rec, v := p.Laid.Col(12, 12, 6, 6); v {
		rec.Width = float32(rl.GetScreenWidth())
		rec.Height = float32(rl.GetScreenHeight())
		Logf("Settings rect: %v\n", rec)
		rl.DrawRectangleRec(rec, ui.WindowBg1)
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
	if input.Tick(ev.InputKindB) == 1 {
		hideSettings()
		return true
	}
	return false
}
