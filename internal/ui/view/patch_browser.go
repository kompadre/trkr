package view

import (
	"fmt"
	"os"
	"path/filepath"
	. "trkr"
	"trkr/internal/audio"
	"trkr/internal/audio/fm"
	ev "trkr/internal/events"
	"trkr/internal/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	patchBrowserVisible bool
	lastFocus           *ui.Element

	libraryFiles    []string
	selectedFileIdx int

	sourceBank     *fm.Bank
	selectedSrcIdx int

	projectBank    *fm.Bank
	selectedPrjIdx int

	activePane int // 0: Library, 1: Source, 2: Project
)

func CreatePatchBrowser(parent *ui.Element) {
	core := ui.NewElementCoreInstance(showPatchBrowser, hidePatchBrowser, patchBrowserHandleInputs, drawPatchBrowser)
	uiElem := ui.NewElement(0, 0, int32(ui.GetOptions().ScreenWidth), int32(ui.GetOptions().ScreenHeight), core, parent)
	uiElem.Name = "Patch Browser"
	uiElem.Visible = false
	uiElem.IsAnchor = true
	uiElem.IsModal = true
	ui.PatchBrowserDialog = uiElem

	refreshLibrary()
}

func showPatchBrowser() {
	lastFocus = ui.PatchBrowserDialog.Parent.FocusedChild
	patchBrowserVisible = true
	ui.PatchBrowserDialog.Visible = true
	ui.PatchBrowserDialog.Parent.SetFocus(ui.PatchBrowserDialog)

	// Initialize project bank from synth
	projectBank = &fm.Bank{}
	data := make([]byte, 4096)
	audio.SynthInstance.GetBank(data)
	for i := 0; i < 32; i++ {
		copy(projectBank.Voices[i][:], data[i*128:(i+1)*128])
	}
}

func hidePatchBrowser() {
	patchBrowserVisible = false
	ui.PatchBrowserDialog.Visible = false
	if lastFocus != nil {
		ui.PatchBrowserDialog.Parent.SetFocus(lastFocus)
	} else {
		ui.PatchBrowserDialog.Parent.HighlightJump(0)
	}
}

func refreshLibrary() {
	libraryFiles = []string{}
	filepath.Walk("./assets/syx", func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && filepath.Ext(path) == ".syx" {
			libraryFiles = append(libraryFiles, path)
		}
		return nil
	})
	if len(libraryFiles) > 0 {
		loadLibraryFile(libraryFiles[0])
	}
}

func loadLibraryFile(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	bank, err := fm.ParseSysex(data)
	if err == nil {
		sourceBank = bank
	}
}

func drawPatchBrowser(ctx ev.EventContext, hasFocus bool, isHighlighted bool) bool {
	p := ctx.EventPayload.(*ui.ElementDrawPayload)
	laid := p.Laid

	rl.DrawRectangle(0, 0, int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()), ui.WindowBg5)

	// Column 1: Library (Files)
	if rec, v := laid.Col(4, 4, 4, 4); v {
		rec.Height = float32(rl.GetScreenHeight())
		bgColor := RGBA(30, 30, 30, 255)
		if activePane == 0 {
			bgColor = RGBA(60, 60, 60, 255)
			rl.DrawRectangleLinesEx(rec, 2, ui.GetOptions().ColorHighlight)
		}
		rl.DrawRectangleRec(rec, bgColor)
		ui.DrawText("LIBRARY", int32(rec.X)+5, int32(rec.Y)+5, 12, rl.LightGray)

		y := int32(rec.Y) + 25
		viewOffset := 0
		if selectedFileIdx > 15 {
			viewOffset = selectedFileIdx - 15
		}

		for i, f := range libraryFiles {
			if i < viewOffset || i > viewOffset+16 {
				continue
			}
			color := rl.White
			if i == selectedFileIdx {
				rl.DrawRectangle(int32(rec.X), y, int32(rec.Width), 16, RGBA(255, 255, 255, 40))
			}
			ui.DrawText(filepath.Base(f), int32(rec.X)+5, y, 14, color)
			y += 16
		}
	}

	// Column 2: Source Voices
	if rec, v := laid.Col(4, 4, 4, 4); v {
		rec.Height = float32(rl.GetScreenHeight())
		bgColor := RGBA(30, 30, 30, 255)
		if activePane == 1 {
			bgColor = RGBA(60, 60, 60, 255)
			rl.DrawRectangleLinesEx(rec, 2, ui.GetOptions().ColorHighlight)
		}
		rl.DrawRectangleRec(rec, bgColor)
		ui.DrawText("SOURCE", int32(rec.X)+5, int32(rec.Y)+5, 12, rl.LightGray)

		if sourceBank != nil {
			y := int32(rec.Y) + 22
			lineHeight := int32(16)
			viewOffset := 0
			if selectedSrcIdx > 15 {
				viewOffset = selectedSrcIdx - 15
			}

			for i := 0; i < 32; i++ {
				if i < viewOffset || i > viewOffset+16 {
					continue
				}

				color := rl.White
				if i == selectedSrcIdx {
					rl.DrawRectangle(int32(rec.X), y, int32(rec.Width), int32(lineHeight), RGBA(255, 255, 255, 40))
				}
				ui.DrawText(fmt.Sprintf("%02d:%s", i+1, sourceBank.Voices[i].Name()), int32(rec.X)+2, y, 14, color)
				y += lineHeight
			}
		}
	}

	// Column 3: Project Bank
	if rec, v := laid.Col(4, 4, 4, 4); v {
		rec.Height = float32(rl.GetScreenHeight())
		bgColor := RGBA(30, 30, 30, 255)
		if activePane == 2 {
			bgColor = RGBA(60, 60, 60, 255)
			rl.DrawRectangleLinesEx(rec, 2, ui.GetOptions().ColorHighlight)
		}
		rl.DrawRectangleRec(rec, bgColor)
		ui.DrawText("PROJECT", int32(rec.X)+5, int32(rec.Y)+5, 12, rl.LightGray)

		if projectBank != nil {
			y := int32(rec.Y) + 22
			lineHeight := int32(16)
			viewOffset := 0
			if selectedPrjIdx > 15 {
				viewOffset = selectedPrjIdx - 15
			}

			for i := 0; i < 32; i++ {
				if i < viewOffset || i > viewOffset+16 {
					continue
				}
				color := rl.White
				if i == selectedPrjIdx {
					rl.DrawRectangle(int32(rec.X), y, int32(rec.Width), int32(lineHeight), RGBA(255, 255, 255, 40))
				}
				ui.DrawText(fmt.Sprintf("%02d:%s", i+1, projectBank.Voices[i].Name()), int32(rec.X)+2, y, 14, color)
				y += lineHeight
			}
		}
	}

	return true
}

func patchBrowserHandleInputs(input *ev.InputSnapshot, el *ui.Element) bool {
	if input.Down(ev.InputKindA) && input.Down(ev.InputKindB) {
		if activePane == 2 && projectBank != nil {
			projectBank.Voices[selectedPrjIdx].Clear()
			audio.SynthInstance.SetVoice(selectedPrjIdx, projectBank.Voices[selectedPrjIdx][:])
			input.ClearHoldTimers(ev.InputKindA, ev.InputKindB)
			return true
		}
	}

	if input.Tick(ev.InputKindA) == 1 && !input.Down(ev.InputKindB) {
		if activePane == 1 && sourceBank != nil {
			// Find first empty slot in project bank
			targetIdx := -1
			for i := 0; i < 32; i++ {
				if projectBank.Voices[i].IsEmpty() {
					targetIdx = i
					break
				}
			}

			if targetIdx != -1 {
				// Copy source to project
				projectBank.Voices[targetIdx] = sourceBank.Voices[selectedSrcIdx]
				audio.SynthInstance.SetVoice(targetIdx, projectBank.Voices[targetIdx][:])
				// Move project selection to the newly filled slot for feedback
				selectedPrjIdx = targetIdx
				return true
			}
		}
	}

	if input.Tick(ev.InputKindB) == 1 && !input.Down(ev.InputKindA) {
		hidePatchBrowser()
		return true
	}

	if input.Tick(ev.InputKindLeft) == 1 || (input.Tick(ev.InputKindLeft) > 30 && input.Tick(ev.InputKindLeft)%10 == 0) {
		activePane = (activePane + 2) % 3
		return true
	}
	if input.Tick(ev.InputKindRight) == 1 || (input.Tick(ev.InputKindRight) > 30 && input.Tick(ev.InputKindRight)%10 == 0) {
		activePane = (activePane + 1) % 3
		return true
	}
	if input.Tick(ev.InputKindL) == 1 {
		activePane = 0
		return true
	}
	if input.Tick(ev.InputKindR) == 1 {
		activePane = 2
		return true
	}

	if input.Down(ev.InputKindUp) {
		switch activePane {
		case 0:
			if len(libraryFiles) > 0 {
				selectedFileIdx = (selectedFileIdx + len(libraryFiles) - 1) % len(libraryFiles)
				loadLibraryFile(libraryFiles[selectedFileIdx])
			}
		case 1:
			selectedSrcIdx = (selectedSrcIdx + 31) % 32
			previewVoice()
		case 2:
			selectedPrjIdx = (selectedPrjIdx + 31) % 32
		}
		return true
	}

	if input.Down(ev.InputKindDown) {
		switch activePane {
		case 0:
			if len(libraryFiles) > 0 {
				selectedFileIdx = (selectedFileIdx + 1) % len(libraryFiles)
				loadLibraryFile(libraryFiles[selectedFileIdx])
			}
		case 1:
			selectedSrcIdx = (selectedSrcIdx + 1) % 32
			previewVoice()
		case 2:
			selectedPrjIdx = (selectedPrjIdx + 1) % 32
		}
		return true
	}

	return true
}

func previewVoice() {
	if sourceBank != nil {
		// Temporarily set the voice to the current track's program slot for preview
		track := CurrentTrack()
		audio.SynthInstance.SetVoice(int(track.Instrument.Program), sourceBank.Voices[selectedSrcIdx][:])
	}
}
