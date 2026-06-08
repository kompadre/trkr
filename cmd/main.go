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
	"trkr/internal/ui/view/settings"
	"trkr/internal/ui/view/track"

	rl "github.com/gen2brain/raylib-go/raylib"
	"os"
)

func main() {
	audio.Init()
	defer audio.CleanUp()
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
		CurrentProject = demoProject()
	} else {
		project := Project{}
		CurrentProject = &project
		LoadProject("autosave.json")
		project.Filename = "autosave.json"
	}

	if runtime.GOARCH == "arm64" {
		opts := ui.GetOptions()
		display := rl.GetCurrentMonitor()
		opts.ScreenWidth, opts.ScreenHeight = rl.GetMonitorWidth(display), rl.GetMonitorHeight(display)
		rl.SetWindowSize(opts.ScreenWidth, opts.ScreenHeight)
		rl.ToggleFullscreen()
	}

	ui.RootElement = &ui.Element{ID: ui.GetSpareId()}

	track.Create(ui.RootElement)
	settings.Create(ui.RootElement)

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
				EventPayload: ev.CalculateInputSnapshot(),
			}) {
				redrawFrames = 3
				if currentFPS > 0 {
					skipInputTriggers = int(0.1 * float32(currentFPS))
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
		ui.CurrentFrame++
	}
}

func demoProject() *Project {
	currentProject := Project{
		Filename: "autosave.json",
		Tracks: []Track{
			{Phrases: []Phrase{{}}},
			{Phrases: []Phrase{{}}},
			{Phrases: []Phrase{{}}},
		},
	}

	currentTrack := &currentProject.Tracks[0]
	currentTrack.IsMultisample = true
	currentTrack.Samples = []Sample{
		{SampleFile: "./assets/music/kick.wav"},
		{SampleFile: "./assets/music/snare.wav"},
		{SampleFile: "./assets/music/hat.wav"},
	}
	currentPhrase := &currentTrack.Phrases[0]
	for i := range currentPhrase.Steps {
		if i%4 == 0 {
			currentPhrase.Steps[i].Notes[0] = 1
		}
		if i%2 == 0 {
			currentPhrase.Steps[i].Notes[1] = 2
		}
	}

	currentTrack = &currentProject.Tracks[1]
	currentTrack.IsMultisample = false
	currentTrack.Sample = Sample{
		SampleFile: "./assets/music/key.wav", RootNote: ParseNote("C 2"),
	}

	currentTrack = &currentProject.Tracks[2]
	currentTrack.IsMultisample = false
	currentTrack.Sample = Sample{
		SampleFile: "./assets/music/key.wav", RootNote: ParseNote("C 2"),
	}
	return &currentProject
}
