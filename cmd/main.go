package main

import (
	"fmt"
	"runtime"
	"time"
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
	rl.SetTargetFPS(30)
	rl.InitWindow(int32(ui.GetOptions().ScreenWidth), int32(ui.GetOptions().ScreenHeight), "trkr v.0.0.1")
	defer rl.CloseWindow()
	// Force SDL to negotiate an OpenGL ES 3.0 context instead of Desktop OpenGL

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
	}

	if runtime.GOARCH == "arm64" {
		opts := ui.GetOptions()
		display := rl.GetCurrentMonitor()
		opts.ScreenWidth, opts.ScreenHeight = rl.GetMonitorWidth(display), rl.GetMonitorHeight(display)
		rl.SetWindowSize(opts.ScreenWidth, opts.ScreenHeight)
		rl.ToggleFullscreen()
	}

	rootElement := &ui.Element{ID: ui.GetSpareId()}

	track.Create(rootElement)
	settings.Create(rootElement)

	ev.RegisterCallback(ev.EventKindInput, func(ctx ev.EventContext) bool {
		return rootElement.HandleInput(ctx.EventPayload.(ev.InputSnapshot))
	}, rootElement.ID)

	ev.RegisterCallback(ev.EventKindGuiDraw, func(ctx ev.EventContext) bool {
		return rootElement.Draw(ctx)
	}, rootElement.ID)

	var tickerPause time.Duration = 16 * time.Millisecond
	timer := time.NewTimer(tickerPause)
	defer timer.Stop() // Clean up when the playback loop terminates
	skipInputTriggers := 0
	ui.RootElement = rootElement
	ui.Font = rl.LoadFont("./assets/fonts/Montserrat-Bold.ttf")
	drawPayload := &ui.ElementDrawPayload{}
	for !rl.WindowShouldClose() {
		select {
		case <-timer.C:
			timer.Reset(tickerPause)
		}

		if skipInputTriggers > 0 {
			skipInputTriggers--
		} else {
			if ev.Trigger(ev.EventKindInput, ev.EventContext{
				EventData:    nil,
				EventPayload: ev.CalculateInputSnapshot(),
			}) {
				skipInputTriggers = 5
			}
		}
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)
		if !player.IsPlaying || runtime.GOARCH != "arm64" {
			ev.Trigger(ev.EventKindGuiDraw, ev.EventContext{
				EventData:    nil,
				EventPayload: drawPayload,
			})
		}
		rl.EndDrawing()
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
