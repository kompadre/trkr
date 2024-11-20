package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"runtime"
	"time"
	. "trkr"
	ui "trkr/internal"
	"trkr/internal/events"
	"trkr/internal/player"
	track "trkr/internal/views"
)

func main() {
	rl.InitWindow(int32(ui.GetOptions().ScreenWidth), int32(ui.GetOptions().ScreenHeight), "internal v0.0")
	defer rl.CloseWindow()
	rl.SetTargetFPS(30)

	CurrentProject = demoProject()

	if runtime.GOARCH == "arm64" {
		opts := ui.GetOptions()
		display := rl.GetCurrentMonitor()
		opts.ScreenWidth, opts.ScreenHeight = rl.GetMonitorWidth(display), rl.GetMonitorHeight(display)
		rl.SetWindowSize(opts.ScreenWidth, opts.ScreenHeight)
		rl.ToggleFullscreen()
	}

	//events.RegisterCallback(events.EventKindTick, func(ctx events.EventContext) bool {
	//	fmt.Printf("Called tick callback. Payload: %d.\n", ctx.EventPayload.(int64))
	//	return true
	//})
	var nextInput = time.Now().Add(time.Millisecond * 250).UnixMilli()
	//var nextTick = time.Now().Add(time.Second).UnixMilli()

	track.Show()
	for !rl.WindowShouldClose() {
		ts := time.Now()
		if nextInput < ts.UnixMilli() {
			if events.Trigger(events.EventKindInput, events.EventContext{
				EventData:    nil,
				EventPayload: events.CalculateInputSnapshot(),
			}) {
				nextInput = ts.Add(time.Millisecond * 250).UnixMilli()
			}
		}

		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)
		if !player.IsPlaying || runtime.GOARCH != "arm64" {
			events.Trigger(events.EventKindGuiDraw, events.EventContext{
				EventData:    nil,
				EventPayload: ts.UnixMilli(),
			})
		}
		rl.EndDrawing()
	}
}

func demoProject() *Project {
	currentProject := Project{
		Tracks: []Track{{Phrases: []Phrase{{}, {}, {}}}, {}, {}}}
	currentTrack := &currentProject.Tracks[0]
	currentTrack.IsMultisample = true
	currentTrack.Samples = []Sample{
		{SampleFile: "./assets/music/kick.wav"},
		{SampleFile: "./assets/music/snare.wav"},
		{SampleFile: "./assets/music/hat.wav"},
	}
	currentPhrase := &currentTrack.Phrases[0]
	currentPhrase.Steps[0x00].Notes = [MaxNotesInStep]Note{1, 0}
	currentPhrase.Steps[0x04].Notes = [MaxNotesInStep]Note{2, 0}
	currentPhrase.Steps[0x06].Notes = [MaxNotesInStep]Note{1, 0}
	currentPhrase.Steps[0x08].Notes = [MaxNotesInStep]Note{1, 0}
	currentPhrase.Steps[0x0c].Notes = [MaxNotesInStep]Note{2, 0}
	currentPhrase.Steps[0x10].Notes = [MaxNotesInStep]Note{1, 0}
	currentPhrase.Steps[0x14].Notes = [MaxNotesInStep]Note{2, 0}
	currentPhrase.Steps[0x18].Notes = [MaxNotesInStep]Note{1, 0}
	for i := range currentPhrase.Steps {
		currentPhrase.Steps[i].Notes[1] = 2
	}
	currentPhrase = &currentTrack.Phrases[1]
	currentPhrase.Steps[0x00].Notes = [MaxNotesInStep]Note{1, 0}
	currentPhrase.Steps[0x04].Notes = [MaxNotesInStep]Note{2, 0}
	currentPhrase.Steps[0x06].Notes = [MaxNotesInStep]Note{1, 0}
	currentPhrase.Steps[0x08].Notes = [MaxNotesInStep]Note{1, 0}
	currentPhrase.Steps[0x0c].Notes = [MaxNotesInStep]Note{2, 0}
	currentPhrase.Steps[0x10].Notes = [MaxNotesInStep]Note{1, 0}
	currentPhrase.Steps[0x14].Notes = [MaxNotesInStep]Note{2, 0}
	currentPhrase.Steps[0x18].Notes = [MaxNotesInStep]Note{1, 0}
	for i := range currentPhrase.Steps {
		currentPhrase.Steps[i].Notes[1] = 2
	}

	currentPhrase = &currentTrack.Phrases[2]
	for i := range currentPhrase.Steps {
		currentPhrase.Steps[i].Notes[0] = 2
	}

	currentTrack = &currentProject.Tracks[1]
	currentTrack.Samples = []Sample{
		{SampleFile: "./assets/music/key.wav", RootNote: ParseNote("C 2")},
	}
	////currentPhrase = &currentTrack.Phrases[0]
	////currentPhrase.Steps[0].Notes = [MaxNotesInStep]Note{ParseNote("C 2"), 0, 0}
	////currentPhrase.Steps[4].Notes = [MaxNotesInStep]Note{0, ParseNote("D#2"), 0}
	////currentPhrase.Steps[8].Notes = [MaxNotesInStep]Note{0, 0, ParseNote("G 2")}
	////currentPhrase.Steps[12].Notes = [MaxNotesInStep]Note{ParseNote("D#2"), 0, 0}
	////currentPhrase.Steps[16].Notes = [MaxNotesInStep]Note{0, ParseNote("C 2"), 0}
	////currentPhrase.Steps[20].Notes = [MaxNotesInStep]Note{0, 0, ParseNote("D#2")}
	////currentPhrase.Steps[24].Notes = [MaxNotesInStep]Note{ParseNote("C 2"), 0, 0}
	//
	//currentPhrase = &Phrase{}
	//currentPhrase.Steps[0].Notes = [MaxNotesInStep]Note{ParseNote("C 2"), 0, 0}
	//currentPhrase.Steps[4].Notes = [MaxNotesInStep]Note{0, ParseNote("F 2"), 0}
	//currentPhrase.Steps[8].Notes = [MaxNotesInStep]Note{0, 0, ParseNote("G#2")}
	//currentPhrase.Steps[12].Notes = [MaxNotesInStep]Note{ParseNote("F 2"), 0, 0}
	//currentPhrase.Steps[16].Notes = [MaxNotesInStep]Note{0, ParseNote("C 2"), 0}
	//currentPhrase.Steps[20].Notes = [MaxNotesInStep]Note{0, 0, ParseNote("F 2")}
	//currentPhrase.Steps[24].Notes = [MaxNotesInStep]Note{ParseNote("C 2"), 0, 0}
	//currentTrack.Phrases = append(currentTrack.Phrases, *currentPhrase)

	currentPhrase = &Phrase{}
	currentPhrase.Steps[0].Notes = [MaxNotesInStep]Note{ParseNote("C 2"), ParseNote("D#2"), ParseNote("G 2"), ParseNote("A#2")}
	currentPhrase.Steps[16].Notes = [MaxNotesInStep]Note{ParseNote("C 2"), ParseNote("D#2"), ParseNote("G 2"), ParseNote("G#2")}
	currentTrack.Phrases = append(currentTrack.Phrases, *currentPhrase)

	currentPhrase = &Phrase{}
	currentPhrase.Steps[0].Notes = [MaxNotesInStep]Note{ParseNote("C 2"), ParseNote("D#2"), ParseNote("G 2"), ParseNote("A#2")}
	currentPhrase.Steps[16].Notes = [MaxNotesInStep]Note{ParseNote("C 2"), ParseNote("D#2"), ParseNote("G 2"), ParseNote("D 3")}
	currentTrack.Phrases = append(currentTrack.Phrases, *currentPhrase)

	currentPhrase = &Phrase{}
	currentPhrase.Steps[0].Notes = [MaxNotesInStep]Note{ParseNote("C 2"), ParseNote("D 2"), ParseNote("F 2"), ParseNote("A#2")}
	currentPhrase.Steps[16].Notes = [MaxNotesInStep]Note{ParseNote("C 2"), ParseNote("F 2"), ParseNote("G#2"), ParseNote("A#2")}
	currentTrack.Phrases = append(currentTrack.Phrases, *currentPhrase)

	currentPhrase = &Phrase{}
	currentPhrase.Steps[0].Notes = [MaxNotesInStep]Note{ParseNote("C 2"), ParseNote("D 2"), ParseNote("F 2"), ParseNote("A#2")}
	currentPhrase.Steps[16].Notes = [MaxNotesInStep]Note{ParseNote("D 2"), ParseNote("F 2"), ParseNote("G 2"), ParseNote("A#2")}
	currentTrack.Phrases = append(currentTrack.Phrases, *currentPhrase)

	currentTrack = &currentProject.Tracks[2]
	currentTrack.Samples = []Sample{
		{SampleFile: "./assets/music/key.wav", RootNote: ParseNote("C 2")},
	}
	currentPhrase = &Phrase{}
	currentPhrase.Steps[0].Notes = [MaxNotesInStep]Note{ParseNote("C 4")}
	currentPhrase.Steps[1].Notes = [MaxNotesInStep]Note{ParseNote("G 2")}
	currentPhrase.Steps[4].Notes = [MaxNotesInStep]Note{ParseNote("F 2")}
	currentPhrase.Steps[5].Notes = [MaxNotesInStep]Note{ParseNote("D 2")}
	currentPhrase.Steps[8].Notes = [MaxNotesInStep]Note{ParseNote("G#2")}
	currentPhrase.Steps[9].Notes = [MaxNotesInStep]Note{ParseNote("G 2")}
	currentTrack.Phrases = append(currentTrack.Phrases, *currentPhrase)

	currentPhrase = &Phrase{}
	currentPhrase.Steps[0].Notes = [MaxNotesInStep]Note{ParseNote("C 4")}
	currentPhrase.Steps[1].Notes = [MaxNotesInStep]Note{ParseNote("C 5")}
	currentPhrase.Steps[4].Notes = [MaxNotesInStep]Note{ParseNote("C 4")}
	currentPhrase.Steps[5].Notes = [MaxNotesInStep]Note{ParseNote("C 3")}
	currentPhrase.Steps[8].Notes = [MaxNotesInStep]Note{ParseNote("C 4")}
	currentPhrase.Steps[9].Notes = [MaxNotesInStep]Note{ParseNote("C 5")}
	currentTrack.Phrases = append(currentTrack.Phrases, *currentPhrase)

	//currentPhrase.Steps[24].Notes = [MaxNotesInStep]Note{0, 0, ParseNote("C 4")}
	////currentPhrase.Steps[6].Notes = [MaxNotesInStep]Note{0, ParseNote("C"), 0}
	////currentPhrase.Steps[8].Notes = [MaxNotesInStep]Note{ParseNote("D"), 0, 0}
	////currentPhrase.Steps[10].Notes = [MaxNotesInStep]Note{0, ParseNote("D#2"), 0}
	////currentPhrase.Steps[12].Notes = [MaxNotesInStep]Note{ParseNote("D#2"), 0, 0}
	//currentPhrase = &currentTrack.Phrases[1]
	//currentPhrase.Steps[17].Notes = [MaxNotesInStep]Note{22, 0}
	//currentPhrase.Steps[21].Notes = [MaxNotesInStep]Note{23, 0}
	//currentPhrase.Steps[23].Notes = [MaxNotesInStep]Note{12, 0, 17}
	//currentPhrase.Steps[24].Notes = [MaxNotesInStep]Note{0, 16, 0}
	//currentPhrase.Steps[25].Notes = [MaxNotesInStep]Note{24, 0, 20}
	//currentPhrase = &currentTrack.Phrases[3]
	//currentPhrase.Steps[0x01].Notes = [MaxNotesInStep]Note{14, 0}
	//currentPhrase = &currentTrack.Phrases[4]
	//currentPhrase.Steps[23].Notes = [MaxNotesInStep]Note{12, 0, 17}
	//currentPhrase.Steps[24].Notes = [MaxNotesInStep]Note{0, 14, 0}
	//currentPhrase.Steps[25].Notes = [MaxNotesInStep]Note{0, 0, 21}
	return &currentProject
}
