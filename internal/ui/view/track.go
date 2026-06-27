package view

import (
	"fmt"
	"math"
	. "trkr"
	"trkr/external/msfa"
	"trkr/internal/audio"
	"trkr/internal/events"
	ev "trkr/internal/events"
	"trkr/internal/player"
	ui "trkr/internal/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var columnStarts = []int{-160, 40, 80, 120, 160, 200, 240, 280, 320}
var columnWidths = []int{160, 40, 40, 40, 40, 40, 40, 40, 40}
var currentCol int
var currentRow int
var SyncToPlay bool
var LastNote Note = Note(37)
var uiElem *ui.Element
var leftPaneKnobs map[uint8](func(delta int))

var RowsInScreen = int((math.Floor(float64(ui.GetOptions().ScreenHeight-40) / 20)))

func trackShow() {
	uiElem.Visible = true
}

func trackHide() {
	uiElem.Visible = false
}

func GetTrackElement() *ui.Element {
	return uiElem
}

func CreateTrack(rootElement *ui.Element) {
	core := ui.NewElementCoreInstance(trackShow, trackHide, handleInputTrack, drawTrack)
	uiElem = ui.NewElement(10, 10, int32(ui.GetOptions().ScreenWidth), int32(ui.GetOptions().ScreenHeight), core, rootElement)
	ui.TrackDialog = uiElem
	leftPaneKnobs = make(map[uint8]func(delta int))
	// Skips

	leftPaneKnobs[1] = func(delta int) {
		currentTrack := &CurrentProject.Tracks[Clamp(ui.TrackId, 0, len(CurrentProject.Tracks)-1)]
		currentTrack.Volume = Clamp(currentTrack.Volume+(float64(delta)*0.1), 0.0, 1.0)
		msfa.ChangeVolume(currentTrack.Id, int32(Clamp(256*currentTrack.Volume, 0, 256)))
	}
	leftPaneKnobs[2] = func(delta int) {
		currentTrack := &CurrentProject.Tracks[Clamp(ui.TrackId, 0, len(CurrentProject.Tracks)-1)]
		currentTrack.Skips = Clamp(uint8(int(currentTrack.Skips)+delta), 0, 16)
	}
	leftPaneKnobs[3] = func(delta int) {
		currentTrack := &CurrentProject.Tracks[Clamp(ui.TrackId, 0, len(CurrentProject.Tracks)-1)]
		currentPhrase := currentTrack.Phrases[ui.PhraseId]
		currentPhrase.Repeats = Clamp(uint8(int(currentPhrase.Repeats)+delta), 0, 16)
		currentPhrase.CurrentRepeat = 0
	}

	leftPaneKnobs[6] = func(delta int) {
		currentTrack := &CurrentProject.Tracks[Clamp(ui.TrackId, 0, len(CurrentProject.Tracks)-1)]
		currentTrack.CurrentProgram = Clamp(int(currentTrack.CurrentProgram)+delta, 0, 31)
		currentTrack.Instrument.Program = uint8(currentTrack.CurrentProgram)
		audio.SynthInstance.ChangeProgram(currentTrack.Id, uint8(currentTrack.CurrentProgram))
	}
}

func drawTrack(ctx events.EventContext, hasFocus bool) bool {
	if !uiElem.Visible || len(CurrentProject.Tracks) < 1 || ui.TrackId > len(CurrentProject.Tracks)-1 {
		return false
	}
	currentTrack := CurrentProject.Tracks[Clamp(ui.TrackId, 0, len(CurrentProject.Tracks)-1)]
	currentPhrase := currentTrack.Phrases[Clamp(ui.PhraseId, 0, len(currentTrack.Phrases)-1)]

	firstVerticalItem := max(0, currentRow-RowsInScreen)
	noteColor := rl.Lime
	laid := ctx.EventPayload.(*ui.ElementDrawPayload).Laid

	if v := laid.EnterCol(4, 4, 4, 4); v {
		laid.Pad(10, 10)
		laid.TextBlock(fmt.Sprintf("Track %d", ui.TrackId), 20, rl.White)
		if v = laid.EnterCol(12, 12, 12, 12); v {
			if CurrentProject.Tracks[ui.TrackId].Instrument != nil {
				laid.Pad(10, 0)
				laid.TextBlock(fmt.Sprintf("Volume: %0.2f", currentTrack.Volume*100.0), 20, rl.White)
				laid.TextBlock(fmt.Sprintf("Skips: %d", currentTrack.Skips), 20, rl.White)
				laid.TextBlock(fmt.Sprintf("Repeats: %d(%d)", currentPhrase.Repeats, (currentPhrase.Repeats-currentPhrase.CurrentRepeat)), 20, rl.White)
				laid.TextBlock(CurrentProject.Tracks[ui.TrackId].Instrument.SampleSourceType.UiString(), 20, rl.White)
				switch CurrentProject.Tracks[ui.TrackId].Instrument.SampleSourceType {
				case SampleSourceTypeFm:
					laid.TextBlock(audio.SynthPatchName(), 20, rl.White)
					laid.TextBlock(audio.SynthProgramName(currentTrack.Id), 20, rl.White)
				}
			}
			laid.ExitCol()
		}
		laid.ExitCol()
	}

	laid.SetRowHeight(0)
	bounds, v := laid.Col(8, 8, 8, 8)
	if !v {
		return false
	}

	for i, step := range currentPhrase.Steps {
		if i < firstVerticalItem {
			continue
		} else if i > firstVerticalItem+RowsInScreen {
			break
		}

		y := int(bounds.Y) + (i - firstVerticalItem)
		x := int(bounds.X)
		if i%4 == 0 {
			ui.DrawText(fmt.Sprintf("%02d", int(i)), int32(x), int32(10+y*20), 20, rl.Lime)
		} else {
			ui.DrawText(fmt.Sprintf("%02d", int(i)), int32(x), int32(10+y*20), 20, rl.Red)
		}

		for j := 0; j < len(step.Notes); j++ {
			if j >= len(columnStarts) {
				break
			}
			rl.DrawText(step.Notes[j].ToString(), int32(x+columnStarts[j+1]), int32(10+y*20), 20, noteColor)
			if step.Notes[j] == 253 {
				noteColor = rl.Gray
			}
		}
		// rl.DrawText(fmt.Sprintf("%d", int(ui.GetOptions().ScreenWidth)), 100.0, int32(10+i*20), 20, rl.Lime)
		if currentPhrase.CurrentStep == i {
			rl.DrawText(">", 0, int32(10+y*20), 20, rl.Maroon)
		}
	}

	bottomOffset := int32(ui.GetOptions().ScreenHeight - 10)
	rl.DrawRectangle(0, int32(ui.GetOptions().VerticalPadding+(currentRow-firstVerticalItem)*ui.GetOptions().RowHeight), int32(ui.GetOptions().ScreenWidth), 20, ui.GetOptions().ColorHighlight)
	rl.DrawRectangle(int32(bounds.X)+int32(columnStarts[currentCol%len(columnStarts)]), int32(ui.GetOptions().VerticalPadding+(currentRow-firstVerticalItem)*ui.GetOptions().RowHeight), int32(columnWidths[currentCol%len(columnWidths)]), 20, ui.GetOptions().ColorHighlight)
	rl.DrawText(fmt.Sprintf("TRK: %02d/%02d", ui.TrackId, len(CurrentProject.Tracks)-1), 10, bottomOffset, 10, rl.Maroon)
	rl.DrawText(fmt.Sprintf("PHR: %02d/%02d", ui.PhraseId, len(currentTrack.Phrases)-1), 80, bottomOffset, 10, rl.Maroon)
	rl.DrawText(fmt.Sprintf("STP: %02d", currentRow), 150, bottomOffset, 10, rl.Maroon)
	rl.DrawText(fmt.Sprintf("STP: %02d", CurrentProject.Current().Current().CurrentStep), 200, bottomOffset, 10, rl.Maroon)

	return false
}

var lastSnapshot ev.InputSnapshot

func handleInputTrack(input *ev.InputSnapshot, el *ui.Element) bool {
	if !uiElem.Visible || len(CurrentProject.Tracks) < 1 {
		return false
	}

	result := false
	currentTrack := &CurrentProject.Tracks[ui.TrackId]

	if input.Down(ev.InputKindDir) && input.Down(ev.InputKindA) && input.Down(ev.InputKindR) {
		switch true {
		case input.Down(ev.InputKindDown):
			fmt.Printf("Adding a new phrase!\n")
			if input.Down(ev.InputKindR) {
				clone := CurrentProject.Tracks[ui.TrackId].Phrases[ui.PhraseId].Clone()
				currentTrack.Phrases = append(currentTrack.Phrases, clone)
				currentTrack.PhraseIds = append(currentTrack.PhraseIds, clone.ID)
			} else {
				newPhrase := NewPhrase(CurrentProject)
				currentTrack.Phrases = append(currentTrack.Phrases, newPhrase)
				currentTrack.PhraseIds = append(currentTrack.PhraseIds, newPhrase.ID)
			}
			ui.PhraseId = len(currentTrack.Phrases) - 1

		case input.Down(ev.InputKindUp):
			fmt.Printf("Removing a phrase!\n")
			currentTrack.Phrases = currentTrack.Phrases[:len(currentTrack.Phrases)-1]
			currentTrack.PhraseIds = currentTrack.PhraseIds[:len(currentTrack.PhraseIds)-1]
			ResetHead()
			if ui.PhraseId == 0 {
				newPhrase := NewPhrase(CurrentProject)
				currentTrack.Phrases = append(currentTrack.Phrases, newPhrase)
				currentTrack.PhraseIds = append(currentTrack.PhraseIds, newPhrase.ID)
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
		ui.PhraseId = Clamp(ui.PhraseId+1, 0, len(currentTrack.Phrases)-1)
		result = true
	} else if input.Down(ev.InputKindR) && input.Down(ev.InputKindUp) {
		ui.PhraseId = Clamp(ui.PhraseId-1, 0, len(currentTrack.Phrases)-1)
		result = true
	} else if input.Down(ev.InputKindR) && input.Down(ev.InputKindDir) {
		tv := rl.NewVector2(0, 0)
		oldTrackId, oldPhraseId := ui.TrackId, ui.PhraseId
		if input.Down(ev.InputKindLeft) {
			newTrackId := Clamp(ui.TrackId-1, 0, len(CurrentProject.Tracks)-1)
			if newTrackId != oldTrackId {
				tv.X = 20
				_ = ui.NewTransition(el, tv)
				ui.TrackId = Clamp(ui.TrackId-1, 0, len(CurrentProject.Tracks)-1)
				ui.PhraseId = 0
			}
		} else if input.Down(ev.InputKindRight) {
			newTrackId := Clamp(ui.TrackId+1, 0, len(CurrentProject.Tracks)-1)
			if newTrackId != oldTrackId {
				tv.X = -20
				_ = ui.NewTransition(el, tv)
				ui.TrackId = Clamp(ui.TrackId+1, 0, len(CurrentProject.Tracks)-1)
				ui.PhraseId = 0
			}
		} else if input.Down(ev.InputKindUp) {
			newPhraseId := Clamp(ui.PhraseId-1, 0, len(CurrentProject.Tracks[ui.TrackId].Phrases)-1)
			if newPhraseId != oldPhraseId {
				tv.Y = 20
				_ = ui.NewTransition(el, tv)
				ui.PhraseId = newPhraseId
			}
		} else if input.Down(ev.InputKindDown) {
			newPhraseId := Clamp(ui.PhraseId+1, 0, len(CurrentProject.Tracks[ui.TrackId].Phrases)-1)
			if newPhraseId != oldPhraseId {
				tv.Y = -20
				_ = ui.NewTransition(el, tv)
				ui.PhraseId = newPhraseId
			}
		}
		currentRow = 0
		currentCol = 0
		result = true
	} else if input.Down(ev.InputKindA) && input.Down(ev.InputKindDir) {
		if currentCol > 0 {
			noteSlot := &currentTrack.Phrases[ui.PhraseId].Steps[Clamp(currentRow, 0, MaxStepsInPhrase)].Notes[Clamp(currentCol-1, 0, MaxNotesInStep)]
			if input.Down(ev.InputKindRight) {
				*noteSlot++
			} else if input.Down(ev.InputKindLeft) {
				*noteSlot--
			} else if input.Down(ev.InputKindUp) {
				*noteSlot += 12
			} else {
				*noteSlot -= 12
			}
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
			}
		}
		result = true
	} else if input.Down(ev.InputKindB) && (input.Tick(ev.InputKindB) > 120 || input.Down(ev.InputKindDown)) {
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
		noteSlot := &currentTrack.Phrases[ui.PhraseId].Steps[Clamp(currentRow-1, 0, MaxStepsInPhrase)].Notes[Clamp(currentCol-1, 0, MaxNotesInStep)]
		*noteSlot = NoteNone
		input.ClearHoldTimers(ev.InputKindA, ev.InputKindB)
		result = true
	} else if input.Down(ev.InputKindSpace) {
		if player.IsPlaying {
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
	return result
}
