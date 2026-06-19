package main

import (
	"fmt"
	"runtime"
	//"time"
	. "trkr"
	ui "trkr/internal/ui"

	"trkr/internal/audio"
	ev "trkr/internal/events"
	"trkr/internal/player"
	"trkr/internal/ui/view"

	rl "github.com/gen2brain/raylib-go/raylib"
	"os"
)

func main() {
	rl.SetConfigFlags(rl.FlagVsyncHint)
	rl.InitWindow(int32(ui.GetOptions().ScreenWidth), int32(ui.GetOptions().ScreenHeight), "trkr v.0.0.1")
	defer rl.CloseWindow()
	defer func() {
		err := SaveProject()
		if err != nil {
			fmt.Printf("Error autosaving: %v.\n", err)
		}
	}()
	_, err := os.Stat("autosave.json")
	if err != nil {
		fmt.Printf("Loading demo project...\n")
		CurrentProject = demoProject()
	} else {
		CurrentProject = &Project{}
		LoadProject("autosave.json", CurrentProject)
		CurrentProject.Filename = "autosave.json"
	}
	audio.Init()
	defer audio.Cleanup()
	fmt.Printf("Current project has %d tracks.\n", len(CurrentProject.Tracks))
	//fmt.Printf("CurrentProject: %v. UiTrackId: %d\n", CurrentProject, ui.TrackId)

	if runtime.GOARCH == "arm64" {
		opts := ui.GetOptions()
		display := rl.GetCurrentMonitor()
		opts.ScreenWidth, opts.ScreenHeight = rl.GetMonitorWidth(display), rl.GetMonitorHeight(display)
		rl.SetWindowSize(opts.ScreenWidth, opts.ScreenHeight)
		rl.ToggleFullscreen()
	}

	ui.RootElement = &ui.Element{ID: ui.GetSpareId()}

	view.CreateTrack(ui.RootElement)
	view.CreateSongView(ui.RootElement)
	view.CreateSettings(ui.RootElement)
	player.CreatePlayer(ui.RootElement)

	ev.RegisterCallback(ev.EventKindInput, func(ctx ev.EventContext) bool {
		return ui.RootElement.HandleInput(ctx.EventPayload.(ev.InputSnapshot))
	}, ui.RootElement.ID)

	ev.RegisterCallback(ev.EventKindUpdate, func(ctx ev.EventContext) bool {
		return ui.RootElement.Draw(ctx)
	}, ui.RootElement.ID)

	image := rl.LoadImage("./assets/images/bg.png")
	bg := rl.LoadTextureFromImage(image)
	rl.UnloadImage(image)

	skipInputTriggers := 0
	ui.Font = rl.LoadFont("./assets/fonts/JetBrainsMono-SemiBold.ttf")
	drawPayload := &ui.ElementDrawPayload{}
	redrawFrames := 5
	for !rl.WindowShouldClose() {
		currentFPS := rl.GetFPS()
		if skipInputTriggers > 0 {
			skipInputTriggers--
		} else {
			if ev.Trigger(ev.EventKindInput, ev.EventContext{
				EventData:    nil,
				EventPayload: ev.CalculateInput(),
			}) {
				redrawFrames = 3
				if currentFPS > 0 {
					skipInputTriggers = int(0.2 * float32(currentFPS))
				} else {
					skipInputTriggers = 6
				}
			}
		}
		rl.BeginDrawing()
		if redrawFrames > 0 || player.IsPlaying {
			rl.ClearBackground(ui.WindowBg5)
			rl.DrawTexture(bg, 0, 0, rl.White)
			if !player.IsPlaying || runtime.GOARCH != "arm64" {
				ev.Trigger(ev.EventKindUpdate, ev.EventContext{
					EventData:    nil,
					EventPayload: drawPayload,
				})
			}
			redrawFrames--
		}
		rl.EndDrawing()
		if ev.Trigger(ev.EventKindPostUpdate, ev.EventContext{}) {
			redrawFrames = 3
		}

		ev.Trigger(ev.EventKindAudioUpdate, ev.EventContext{})
		ui.CurrentFrame++
	}
}

func demoProject() *Project {
	currentProject := &Project{
		Filename: "autosave.json",
	}
	currentPhrase := NewPhrase(currentProject)

	currentTrack := Track{}
	currentTrack.IsMultisample = true
	currentTrack.Samples = []Sample{
		{SampleFile: "./assets/music/kick.wav"},
		{SampleFile: "./assets/music/snare.wav"},
		{SampleFile: "./assets/music/hat.wav"},
	}
	currentTrack.Phrases = []*Phrase{currentPhrase}
	currentTrack.PhraseIds = []int32{currentPhrase.ID}

	currentProject.Tracks = append(currentProject.Tracks, currentTrack)

	return currentProject
}
