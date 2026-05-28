package settings

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	. "trkr"
	ev "trkr/internal/events"
	"trkr/internal/player"
	ui "trkr/internal/ui"
)

var Showing bool

func Show() {
	ev.RegisterCallback(ev.EventKindGuiDraw, func(ctx ev.EventContext) bool {
		return DrawSettings()
	}, "Draw Settings")
	ev.RegisterCallback(ev.EventKindInput, func(ctx ev.EventContext) bool {
		return handleInputs(ctx.EventPayload.(ev.InputSnapshot))
	}, "Input Settings")

	Showing = true
}
func Hide() {
	ev.PopCallback(ev.EventKindGuiDraw)
	ev.PopCallback(ev.EventKindInput)
}

func DrawSettings() bool {
	rl.DrawRectangle(5, int32(ui.GetOptions().VerticalPadding), int32(ui.GetOptions().ScreenWidth-10), int32(ui.GetOptions().ScreenHeight-(ui.GetOptions().VerticalPadding*2)), rl.Maroon)
	return true
}

func handleInputs(input ev.InputSnapshot) bool {
	var result bool
	if input[ev.InputKindPressedEnter] {
		Hide()
		result = true
	} else if input[ev.InputKindPressedA] {
		err := SaveProject("autosave.json")
		if err != nil {
			fmt.Println(err)
		}
		result = true
	} else if input[ev.InputKindPressedB] {
		player.Stop()
		ResetHead()
		err := LoadProject("autosave.json")
		if err != nil {
			fmt.Println(err)
		}
		result = true
	}

	return result
}
