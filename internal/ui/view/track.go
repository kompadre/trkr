package view

import (
	"fmt"
	"math"
	"path"
	"path/filepath"
	. "trkr"
	"trkr/external/msfa"
	"trkr/internal/audio"
	"trkr/internal/events"
	ev "trkr/internal/events"
	"trkr/internal/player"
	ui "trkr/internal/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var columnStarts = []int{0, 160, 200, 240, 280, 320, 360, 400, 440, 480}
var columnWidths = []int{160, 40, 40, 40, 40, 40, 40, 40, 40, 40}
var currentCol int
var currentRow int
var SyncToPlay bool
var LastNote Note = Note(37)
var uiElem *ui.Element
var leftPaneKnobs map[uint8](func(delta int))
var showVelocities bool

const (
	RowHeight = 16
	FontSize  = 16

	DefaultVelocity = 100
)

var RowsInScreen = int((math.Floor(float64(ui.GetOptions().ScreenHeight-40) / RowHeight)))

func trackShow() {
	uiElem.Visible = true
}

func trackHide() {
	uiElem.Visible = false
}

func GetTrackElement() *ui.Element {
	return uiElem
}

func CurrentTrack() *Track {
	return &CurrentProject.Tracks[Clamp(ui.TrackId, 0, len(CurrentProject.Tracks)-1)]
}

func CurrentPhrase() *Phrase {
	currentTrack := CurrentTrack()
	if len(currentTrack.PhraseIds[ui.SectionId]) > 0 {
		idx := Clamp(ui.PhraseId, 0, len(currentTrack.PhraseIds[ui.SectionId])-1)
		phraseId := currentTrack.PhraseIds[ui.SectionId][idx]
		if int(phraseId) < len(CurrentProject.Phrases) {
			return &CurrentProject.Phrases[phraseId]
		}
	}
	if len(currentTrack.Phrases[ui.SectionId]) > 0 {
		return currentTrack.Phrases[ui.SectionId][Clamp(ui.PhraseId, 0, len(currentTrack.Phrases[ui.SectionId])-1)]
	}
	return &CurrentProject.Phrases[0]
}

func _CurrentPhrase() *Phrase {
	currentTrack := CurrentTrack()
	return currentTrack.Phrases[ui.SectionId][ui.PhraseId]
}

func CreateTrack(rootElement *ui.Element) {
	core := ui.NewElementCoreInstance(trackShow, trackHide, handleInputTrack, drawTrack)
	uiElem = ui.NewElement(10, 10, int32(ui.GetOptions().ScreenWidth), int32(ui.GetOptions().ScreenHeight), core, rootElement)
	ui.TrackDialog = uiElem
	leftPaneKnobs = make(map[uint8]func(delta int))

	leftPaneKnobs[0] = func(delta int) {
		ui.TrackId = Clamp(ui.TrackId+delta, 0, len(CurrentProject.Tracks)-1)
	}
	leftPaneKnobs[1] = func(delta int) {
		currentTrack := CurrentTrack()
		currentTrack.Volume = Clamp(currentTrack.Volume+(float64(delta)*0.1), 0.0, 1.0)
		msfa.ChangeVolume(currentTrack.Id, int32(Clamp(256*currentTrack.Volume, 0, 256)))
	}
	leftPaneKnobs[2] = func(delta int) {
		currentTrack := CurrentTrack()
		currentTrack.Skips = Clamp(uint8(int(currentTrack.Skips)+delta), 0, 16)
	}
	leftPaneKnobs[3] = func(delta int) {
		currentTrack := CurrentProject.Tracks[ui.TrackId]
		currentPhrase := currentTrack.Phrases[ui.SectionId][ui.PhraseId]
		newId := Clamp(currentPhrase.ID+int32(delta), 0, int32(len(CurrentProject.Phrases)-1))
		if newId != currentPhrase.ID {
			Logf("Setting new phraseId %d for %d slot.\n", newId, ui.PhraseId)
			currentTrack.PhraseIds[ui.SectionId][ui.PhraseId] = newId
			currentTrack.Phrases[ui.SectionId][ui.PhraseId] = &CurrentProject.Phrases[newId]
		}
	}
	leftPaneKnobs[4] = func(delta int) {
		currentPhrase := CurrentPhrase()
		currentPhrase.Repeats = Clamp(uint8(int(currentPhrase.Repeats)+delta), 0, 16)
		currentPhrase.CurrentRepeat = 0
	}

	leftPaneKnobs[5] = func(delta int) {
		ins := CurrentProject.Tracks[Clamp(ui.TrackId, 0, len(CurrentProject.Tracks)-1)].Instrument
		ins.SampleSourceType = Clamp(ins.SampleSourceType+SampleSourceType(delta), 0, SampleSourceTypePerc)
		if ins.SampleSourceType == SampleSourceTypeWavefile {
			ins.SampleSource = [3]string{"./assets/music/kick.wav", "./assets/music/snare.wav", "./assets/music/hat.wav"}
			ins.RootNote = 0
			ins.LoopEnd = 0
			ins.LoadSamples()
		}
	}
	leftPaneKnobs[6] = func(delta int) {
		globPattern := "./assets/syx/dexed/*.syx"
		banks := []string{"./assets/syx/Dexed_01.syx"}
		matches, _ := filepath.Glob(globPattern)
		banks = append(banks, matches...)
		nextPatchFile := banks[0]
		currentPatchName := CurrentProject.FmPatchName

		if currentPatchName != "" {
			for i := range banks {
				if path.Base(banks[i]) == path.Base(currentPatchName) {
					nextPatchFile = banks[Clamp(i+delta, 0, len(banks)-1)]
					break
				}
			}
		}
		Logf("Current Patch name: %s. Loading %s.\n", currentPatchName, nextPatchFile)
		audio.SynthInstance.LoadPatch(nextPatchFile)
		CurrentProject.FmPatchName = nextPatchFile

	}

	leftPaneKnobs[7] = func(delta int) {
		currentTrack := &CurrentProject.Tracks[Clamp(ui.TrackId, 0, len(CurrentProject.Tracks)-1)]
		currentTrack.CurrentProgram = Clamp(int(currentTrack.CurrentProgram)+delta, 0, 31)
		currentTrack.Instrument.Program = uint8(currentTrack.CurrentProgram)
		audio.SynthInstance.ChangeProgram(currentTrack.Id, uint8(currentTrack.CurrentProgram))
	}

	switch ui.GetCurrentBreakpoint(rl.GetScreenWidth()) {
	case ui.SizeLG, ui.SizeMD:
		columnStarts = []int{0, 220, 260, 300, 340, 380, 420, 460}
		columnWidths = []int{180, 20, 40, 40, 40, 40, 40, 40, 40}
	}

}

func drawTrack(ctx events.EventContext, hasFocus bool) bool {
	if !uiElem.Visible || len(CurrentProject.Tracks) < 1 || ui.TrackId > len(CurrentProject.Tracks)-1 {
		return false
	}
	currentTrack := CurrentTrack() // CurrentProject.Tracks[Clamp(ui.TrackId, 0, len(CurrentProject.Tracks)-1)]
	currentPhrase := CurrentPhrase()

	firstVerticalItem := max(0, currentRow-RowsInScreen)
	laid := ctx.EventPayload.(*ui.ElementDrawPayload).Laid

	//if v := laid.EnterCol(4, 4, 4, 2); v {
	rect := rl.NewRectangle(0, 0, float32(columnWidths[0]), float32(rl.GetScreenHeight()))
	laid.PushGreedyContext(rect)
	laid.Pad(10, 10)
	laid.TextBlock(fmt.Sprintf("Track %d", ui.TrackId), FontSize, rl.White)
	if CurrentProject.Tracks[ui.TrackId].Instrument != nil {
		laid.Pad(10, 0)
		laid.TextBlock(fmt.Sprintf("Volume: %0.2f", currentTrack.Volume*100.0), FontSize, rl.White)
		if currentTrack.Skips > 0 {
			laid.TextBlock(fmt.Sprintf("Speed: 1/%d", currentTrack.Skips), FontSize, rl.White)
		} else {
			laid.TextBlock("Speed: Normal", FontSize, rl.White)
		}
		laid.TextBlock(fmt.Sprintf("Phrase: %d", currentPhrase.ID), FontSize, rl.White)
		laid.Pad(10, 0)
		laid.TextBlock(fmt.Sprintf("Repeats: %d", currentPhrase.Repeats), FontSize, rl.White)
		laid.Pad(-10, 0)
		laid.TextBlock(CurrentProject.Tracks[ui.TrackId].Instrument.SampleSourceType.UiString(), FontSize, rl.White)
		switch CurrentProject.Tracks[ui.TrackId].Instrument.SampleSourceType {
		case SampleSourceTypeFm:
			laid.TextBlock(path.Base(CurrentProject.FmPatchName), FontSize, rl.White)
			laid.TextBlock(fmt.Sprintf("%d: %s", currentTrack.Instrument.Program, audio.SynthProgramName(currentTrack.Id)), FontSize, rl.White)
		}
		// laid.TextBlock(fmt.Sprintf("BPM: %02.5f", player.EffectiveBpm), rowHeight, rl.White)
	}
	laid.PopContext()

	laid.SetRowHeight(0)
	bottomOffset := int32(ui.GetOptions().ScreenHeight - 10)
	topOffset := 9
	var x, y int32
	var lastX int32 = int32(columnStarts[len(columnStarts)-1] + columnWidths[len(columnWidths)-1])
	for i, step := range currentPhrase.Steps {
		if i < firstVerticalItem {
			continue
		} else if i > firstVerticalItem+RowsInScreen {
			break
		}

		y = int32(topOffset + i*(RowHeight+2))
		x = int32(columnStarts[1])

		rl.DrawLine(int32(x), int32(y-3), lastX, int32(y-3), rl.White)
		rl.DrawLine(int32(x), int32(y-3), lastX, int32(y-3), rl.Gray)
		lineNumFormat := " %02d"
		if i == currentPhrase.CurrentStep {
			lineNumFormat = ">%02d"
		}
		if i%4 == 0 {
			ui.DrawText(fmt.Sprintf(lineNumFormat, int(i)), int32(x-25), int32(y), RowHeight, rl.Lime)
		} else {
			ui.DrawText(fmt.Sprintf(lineNumFormat, int(i)), int32(x-25), int32(y), RowHeight, rl.Red)
		}
		for j := 0; j < len(step.Notes); j++ {
			if j >= len(columnStarts) {
				break
			}
			if step.Notes[j] != NoteNone {
				if !showVelocities {
					ui.DrawText(step.Notes[j].ToString(), int32(8+columnStarts[j+1]), int32(y), RowHeight, NoteColor[step.Notes[j]%12])
				} else {
					ui.DrawText(fmt.Sprintf("%03d", step.Velocities[j]-1), int32(8+columnStarts[j+1]), int32(y), RowHeight, NoteColor[step.Notes[j]%12])
				}
			}
		}
	}

	y += RowHeight + 2
	x = int32(columnStarts[1])
	rl.DrawLine(x, y, int32(columnStarts[len(columnStarts)-1]+columnWidths[len(columnWidths)-1]), y, rl.White)
	rl.DrawLine(x, y+1, int32(columnStarts[len(columnStarts)-1]+columnWidths[len(columnWidths)-1]), y+1, rl.Gray)
	for i := range columnStarts {
		rl.DrawLine(int32(columnStarts[i]+columnWidths[i]), y, int32(columnStarts[i]+columnWidths[i]), int32(topOffset-2), rl.White)
	}

	rl.DrawRectangle(0,
		int32((topOffset-2)+currentRow*(RowHeight+2)),
		int32(ui.GetOptions().ScreenWidth), RowHeight+2,
		RGBA(255, 255, 255, 5))

	rl.DrawRectangle(
		int32(columnStarts[Clamp(currentCol, 0, len(columnStarts)-1)]),
		int32((topOffset-2)+currentRow*(RowHeight+2)),
		int32(columnWidths[Clamp(currentCol, 0, len(columnWidths)-1)]),
		RowHeight+2, ui.GetOptions().ColorHighlight)

	rl.DrawText(fmt.Sprintf("TRK: %02d/%02d", ui.TrackId, len(CurrentProject.Tracks)-1), 10, bottomOffset, 10, rl.Maroon)
	rl.DrawText(fmt.Sprintf("PHR: %02d/%02d", ui.PhraseId, len(currentTrack.Phrases[currentTrack.CurrentSection])-1), 80, bottomOffset, 10, rl.Maroon)
	rl.DrawText(fmt.Sprintf("ROW: %02d", currentTrack.Current().CurrentStep), 200, bottomOffset, 10, rl.Maroon)
	ui.FlushText()

	return false
}

var lastSnapshot ev.InputSnapshot

func handleInputTrack(input *ev.InputSnapshot, el *ui.Element) bool {
	if !uiElem.Visible || len(CurrentProject.Tracks) < 1 {
		return false
	}

	result := false
	currentTrack := CurrentTrack() //&CurrentProject.Tracks[ui.TrackId]

	if input.Down(ev.InputKindDir) && input.Down(ev.InputKindA) && input.Down(ev.InputKindR) {
		switch true {
		case input.Down(ev.InputKindDown):
			fmt.Printf("Adding a new phrase!\n")
			clone := CurrentProject.Phrases[CurrentPhrase().ID].Clone()
			currentTrack.Phrases[ui.SectionId] = append(currentTrack.Phrases[ui.SectionId], clone)
			currentTrack.PhraseIds[ui.SectionId] = append(currentTrack.PhraseIds[ui.SectionId], clone.ID)
			ui.PhraseId = len(currentTrack.Phrases[ui.SectionId]) - 1

		case input.Down(ev.InputKindUp):
			fmt.Printf("Removing a phrase!\n")
			currentTrack.Phrases[ui.SectionId] = currentTrack.Phrases[ui.SectionId][:len(currentTrack.Phrases[ui.SectionId])-1]
			currentTrack.PhraseIds[ui.SectionId] = currentTrack.PhraseIds[ui.SectionId][:len(currentTrack.PhraseIds[ui.SectionId])-1]
			ResetHead()
			if ui.PhraseId == 0 {
				newPhrase := NewPhrase(CurrentProject)
				currentTrack.Phrases[ui.SectionId] = append(currentTrack.Phrases[ui.SectionId], newPhrase)
				currentTrack.PhraseIds[ui.SectionId] = append(currentTrack.PhraseIds[ui.SectionId], newPhrase.ID)
			} else {
				ui.PhraseId--
			}

		case input.Down(ev.InputKindRight):
			NewTrack(CurrentProject)

		case input.Down(ev.InputKindLeft):
			if len(CurrentProject.Tracks) < 2 {
				return false
			}
			fmt.Printf("Removing a track!\n")
			ui.TrackId = Clamp(ui.TrackId-1, 0, len(CurrentProject.Tracks)-1)
			CurrentProject.Tracks[len(CurrentProject.Tracks)-1].Cleanup()
			CurrentProject.Tracks = CurrentProject.Tracks[:len(CurrentProject.Tracks)-1]
		}
		// Returning early because removing track or phrase on the fly has grave consecuences.
		return true
	}

	if input.Down(ev.InputKindR) && input.Down(ev.InputKindDown) {
		Logf("PhraseId: %d. len(Phrases): %d.\n", ui.PhraseId, len(currentTrack.PhraseIds[ui.SectionId]))
		ui.PhraseId = Clamp(ui.PhraseId+1, 0, len(currentTrack.Phrases[ui.SectionId])-1)
		result = true
	} else if input.Down(ev.InputKindR) && input.Down(ev.InputKindUp) {
		ui.PhraseId = Clamp(ui.PhraseId-1, 0, len(currentTrack.Phrases[ui.SectionId])-1)
		result = true
	} else if input.Down(ev.InputKindR) && input.Down(ev.InputKindDir) {
		tv := rl.NewVector2(0, 0)
		oldTrackId, oldPhraseId := ui.TrackId, ui.PhraseId
		if input.Down(ev.InputKindLeft) {
			newTrackId := Clamp(ui.TrackId-1, 0, len(CurrentProject.Tracks)-1)
			if newTrackId != oldTrackId {
				tv.X = RowHeight
				_ = ui.NewTransition(el, tv)
				ui.TrackId = Clamp(ui.TrackId-1, 0, len(CurrentProject.Tracks)-1)
				ui.PhraseId = 0
			}
		} else if input.Down(ev.InputKindRight) {
			newTrackId := Clamp(ui.TrackId+1, 0, len(CurrentProject.Tracks)-1)
			if newTrackId != oldTrackId {
				tv.X = -RowHeight
				_ = ui.NewTransition(el, tv)
				ui.TrackId = Clamp(ui.TrackId+1, 0, len(CurrentProject.Tracks)-1)
				ui.PhraseId = 0
			}
		} else if input.Down(ev.InputKindUp) {
			newPhraseId := Clamp(ui.PhraseId-1, 0, len(CurrentProject.Tracks[ui.TrackId].Phrases[ui.SectionId])-1)
			if newPhraseId != oldPhraseId {
				tv.Y = RowHeight
				_ = ui.NewTransition(el, tv)
				ui.PhraseId = newPhraseId
			}
		} else if input.Down(ev.InputKindDown) {
			newPhraseId := Clamp(ui.PhraseId+1, 0, len(CurrentProject.Tracks[ui.TrackId].Phrases[ui.SectionId])-1)
			if newPhraseId != oldPhraseId {
				tv.Y = -RowHeight
				_ = ui.NewTransition(el, tv)
				ui.PhraseId = newPhraseId
			}
		}
		currentRow = 0
		currentCol = 0
		result = true
	} else if input.Down(ev.InputKindA) && input.Down(ev.InputKindDir) {
		if currentCol > 0 {
			noteSlot := &CurrentPhrase().Steps[Clamp(currentRow, 0, MaxStepsInPhrase)].Notes[Clamp(currentCol-1, 0, MaxNotesInStep)]
			velSlot := &CurrentPhrase().Steps[Clamp(currentRow, 0, MaxStepsInPhrase)].Velocities[Clamp(currentCol-1, 0, MaxNotesInStep)]

			oldNote := *noteSlot
			if input.Down(ev.InputKindRight) {
				*noteSlot++
			} else if input.Down(ev.InputKindLeft) {
				*noteSlot--
			} else if input.Down(ev.InputKindUp) {
				*noteSlot += 12
			} else {
				*noteSlot -= 12
			}

			if *noteSlot == NoteSkip || oldNote == NoteSkip {
				CurrentPhrase().EffectiveRows = 0
			}
			if *noteSlot == NoteNone || *noteSlot == NoteSkip || *noteSlot == NoteOff || *noteSlot == NoteCut {
				*velSlot = 0
			} else if *velSlot == 0 {
				// Default velocity for a new note.
				*velSlot = 100
			}
			result = true
		} else {
			// We are on the first column
			if knob, ok := leftPaneKnobs[uint8(currentRow)]; ok {
				if input.Down(ev.InputKindLeft) {
					knob(-1)
					result = true
				} else if input.Down(ev.InputKindRight) {
					knob(1)
					result = true
				}
				return true
			}
		}
	} else if input.Down(ev.InputKindB) && input.Down(ev.InputKindDir) {
		velSlot := &CurrentPhrase().Steps[Clamp(currentRow, 0, MaxStepsInPhrase)].Velocities[Clamp(currentCol-1, 0, MaxNotesInStep)]
		// oldVel := *velSlot
		if input.Down(ev.InputKindRight) {
			*velSlot++
		} else if input.Down(ev.InputKindLeft) {
			*velSlot--
		}
	} else if input.Down(ev.InputKindL) && input.Down(ev.InputKindDown) {
		_ = ui.NewTransition(ui.SongDialog, rl.NewVector2(0, 0))
		ev.RegisterCallback(ev.EventKindPostUpdate, func(ctx ev.EventContext) bool {
			el.Visible = false
			ui.SongDialog.Visible = true
			ui.RootElement.SetFocus(ui.SongDialog)
			input.ClearHoldTimers(ev.InputKindB)
			return true
		}, el.ID)
	} else if input.Down(ev.InputKindDir) {
		movementMultiplier := int(input.Tick(ev.InputKindDir))
		if input.Down(ev.InputKindDown) {
			currentRow = Clamp(currentRow+movementMultiplier, 0, MaxStepsInPhrase-1)
		} else if input.Down(ev.InputKindUp) {
			currentRow = Clamp(currentRow-movementMultiplier, 0, MaxStepsInPhrase-1)
		} else if input.Down(ev.InputKindLeft) {
			currentCol = Clamp(currentCol-movementMultiplier, 0, len(columnWidths)-1)
		} else if input.Down(ev.InputKindRight) {
			currentCol = Clamp(currentCol+movementMultiplier, 0, len(columnWidths)-1)
		}
		result = true
	} else if currentCol > 0 && input.Tick(ev.InputKindA) > 30 && input.Tick(ev.InputKindB) > 30 {
		noteSlot := &CurrentPhrase().Steps[Clamp(currentRow, 0, MaxStepsInPhrase)].Notes[Clamp(currentCol-1, 0, MaxNotesInStep)]
		if *noteSlot == NoteSkip {
			CurrentPhrase().EffectiveRows = 0
		}
		*noteSlot = NoteNone
		input.ClearHoldTimers(ev.InputKindA, ev.InputKindB)
		result = true
	} else if input.Down(ev.InputKindSpace) {
		if player.IsPlaying {
			currentRow = 0
			ResetHead()
			player.Stop()
		} else {
			ResetHead()
			currentRow = 0
			go player.Play()
		}
		result = true
	} else if input.Down(ev.InputKindEnter) {
		if !ui.SettingsDialog.Visible {
			showSettings()
			result = true
		}
	}

	if input.Down(ev.InputKindB) {
		showVelocities = true
	} else if showVelocities {
		showVelocities = false
	}
	return result
}
