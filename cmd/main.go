package main

import (
	"fmt"
	"runtime"
	"time"
	. "trkr"
	ui "trkr/internal/ui"

	"trkr/internal/audio"
	"trkr/internal/events"
	"trkr/internal/player"
	"trkr/internal/views/track"

	rl "github.com/gen2brain/raylib-go/raylib"
	"os"
)

func main() {

	audio.Init()
	defer audio.CleanUp()
	rl.InitWindow(int32(ui.GetOptions().ScreenWidth), int32(ui.GetOptions().ScreenHeight), "internal v0.0")
	defer rl.CloseWindow()
	rl.SetTargetFPS(30)
	defer func() {
		err := SaveProject("autosave.json")
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

	track.Show()
	var tickerPause time.Duration = 16 * time.Millisecond
	timer := time.NewTimer(tickerPause)
	defer timer.Stop() // Clean up when the playback loop terminates
	skipInputTriggers := 0
	for !rl.WindowShouldClose() {
		select {
		case <-timer.C:
			timer.Reset(tickerPause)
		}

		if skipInputTriggers > 0 {
			skipInputTriggers--
		} else {
			if events.Trigger(events.EventKindInput, events.EventContext{
				EventData:    nil,
				EventPayload: events.CalculateInputSnapshot(),
			}) {
				skipInputTriggers = 5
			}
		}

		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)
		if !player.IsPlaying || runtime.GOARCH != "arm64" {
			events.Trigger(events.EventKindGuiDraw, events.EventContext{
				EventData:    nil,
				EventPayload: nil,
			})
		}
		rl.EndDrawing()
	}
}

func demoProject() *Project {
	currentProject := Project{
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
