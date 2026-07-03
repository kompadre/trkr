package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"time"
	. "trkr"
	ui "trkr/internal/ui"

	"trkr/internal/audio"
	ev "trkr/internal/events"
	"trkr/internal/player"
	"trkr/internal/ui/view"

	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	rl.SetConfigFlags(rl.FlagVsyncHint)
	Logf("Initializing window of size %dx%d.\n", ui.GetOptions().ScreenWidth, ui.GetOptions().ScreenHeight)
	rl.InitWindow(int32(ui.GetOptions().ScreenWidth), int32(ui.GetOptions().ScreenHeight), "trkr v.0.0.1")
	// After InitWindow...
	defer rl.CloseWindow()
	defer func() {
		err := SaveProject()
		if err != nil {
			fmt.Printf("Error autosaving: %v.\n", err)
		}
	}()

	CurrentProject = &Project{}
	_, err := os.Stat("autosave.json")

	if err != nil {
		fmt.Printf("Loading demo project...\n")
		demoProject(CurrentProject)
	} else {
		LoadProject("autosave.json", CurrentProject)
		CurrentProject.Filename = "autosave.json"
		if len(CurrentProject.Tracks) == 0 {
			_ = NewTrack(CurrentProject)
		}
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
	Logf("[ZZZ] Configuring Tracks[0].\n")

	view.CreateTrack(ui.RootElement)
	view.CreateSongView(ui.RootElement)
	view.CreateSettings(ui.RootElement)
	player.CreatePlayer(ui.RootElement)
	defer (func() {
		if player.IsPlaying {
			player.StopChannel <- true
		}
	})()

	ev.RegisterCallback(ev.EventKindInput, func(ctx ev.EventContext) bool {
		return ui.RootElement.HandleInput(ctx.EventPayload.(*ev.InputSnapshot))
	}, ui.RootElement.ID)

	ev.RegisterCallback(ev.EventKindUpdate, func(ctx ev.EventContext) bool {
		return ui.RootElement.Draw(ctx)
	}, ui.RootElement.ID)

	image := rl.LoadImage("./assets/images/bg.png")
	//	bg := rl.LoadTextureFromImage(image)
	rl.UnloadImage(image)

	skipInputTriggers := 0

	filePath := "./assets/fonts/JetBrainsMono-SemiBold.ttf"

	// 1. Load the font directly using Raylib's built-in file loader.
	// This allocates the glyph and rectangle arrays safely inside C memory,
	// completely avoiding Go pointer pinning restrictions.
	ui.Font = rl.LoadFontEx(filePath, view.FontSize, nil, 95)

	// 2. Set the texture filter so your pixel-perfect text stays crisp on the H700 screen
	rl.SetTextureFilter(ui.Font.Texture, rl.FilterPoint)

	laid := ui.Laid{}
	laid.SetBreakpoint(rl.GetScreenWidth())
	drawPayload := ui.ElementDrawPayload{Laid: &laid}
	inputPayload := ev.InputSnapshot{}
	redrawFrames := 5
	var haveFocus bool

	// Pre-allocate event contexts outside the main loop
	inputCtx := ev.EventContext{EventData: nil, EventPayload: &inputPayload}
	updateCtx := ev.EventContext{EventData: nil, EventPayload: &drawPayload}
	postUpdateCtx := ev.EventContext{}
	audioUpdateCtx := ev.EventContext{}

	var profile bool
	flag.BoolVar(&profile, "profile", false, "Start profiling")
	flag.Parse()

	if profile {
		go func() {
			exePath, _ := os.Executable()
			profPath := filepath.Join(filepath.Dir(exePath), "cpu.prof")

			f, err := os.Create(profPath)
			if err != nil {
				return
			}
			defer f.Close()

			log.Println("Profiling started...")
			pprof.StartCPUProfile(f)

			// Let it run for 30 seconds while you play the game
			time.Sleep(30 * time.Second)

			pprof.StopCPUProfile()
			log.Println("Profiling stopped. Safe to copy cpu.prof!")
		}()
	}

	for !rl.WindowShouldClose() {
		if !rl.IsWindowFocused() {
			rl.ClearWindowState(rl.FlagVsyncHint)
			time.Sleep(33 * time.Millisecond)
			haveFocus = false
		} else {
			haveFocus = true
			if !rl.IsWindowState(rl.FlagVsyncHint) {
				rl.SetWindowState(rl.FlagVsyncHint)
			}
		}

		currentFPS := rl.GetFPS()
		if skipInputTriggers > 0 {
			skipInputTriggers--
		} else {
			ev.CalculateInput(inputCtx.EventPayload.(*ev.InputSnapshot))
			if ev.Trigger(ev.EventKindInput, inputCtx) {
				redrawFrames = 3
				if currentFPS > 0 {
					skipInputTriggers = int(0.2 * float32(currentFPS))
				} else {
					skipInputTriggers = 6
				}

			}
		}

		rl.BeginDrawing()
		if haveFocus || redrawFrames > 0 || player.IsPlaying {
			rl.ClearBackground(ui.WindowBg5)
			//			rl.DrawTexture(bg, 0, 0, rl.White)
			laid.PushContext(rl.NewRectangle(0, 0, float32(rl.GetScreenWidth()), float32(rl.GetScreenHeight())))

			if ev.Trigger(ev.EventKindUpdate, updateCtx) {
				redrawFrames += 10
			}
			laid.PopContext()
			ui.FlushText()
			if redrawFrames > 0 {
				redrawFrames--
			}
		}
		rl.EndDrawing()
		if ev.Trigger(ev.EventKindPostUpdate, postUpdateCtx) {
			redrawFrames = 3
		}

		ev.Trigger(ev.EventKindAudioUpdate, audioUpdateCtx)
		ui.CurrentFrame++
	}
}

func demoProject(currentProject *Project) {
	Logf("Creating a new demo project.\n")
	currentProject.Filename = "autosave.json"
	currentPhrase := NewPhrase(currentProject)
	instrument := NewInstrument(currentProject)
	instrument.SampleSourceType = SampleSourceTypeWavefile
	instrument.SampleSource = [3]string{"./assets/music/kick.wav", "./assets/music/snare.wav", "./assets/music/hat.wav"}
	instrument.LoadSamples()
	currentTrack := Track{}
	currentTrack.InstrumentId = instrument.Id
	currentTrack.Instrument = instrument
	currentTrack.Instrument.Program = 0

	currentTrack.Phrases = []*Phrase{currentPhrase}
	currentTrack.PhraseIds = []int32{currentPhrase.ID}
	currentTrack.Volume = 1.0
	currentProject.Tracks = append(currentProject.Tracks, currentTrack)
	Logf("Created demo. Track[0].Instrument.SampleSourceType is %s.\n", currentProject.Tracks[0].Instrument.SampleSourceType.UiString())

}
