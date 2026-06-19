package view

import (
	"fmt"
	. "trkr"
	"trkr/internal/events"
	ev "trkr/internal/events"
	"trkr/internal/player"
	ui "trkr/internal/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var columnStarts = []int{40, 80, 120, 160, 200, 240, 280, 320}
var columnWidths = []int{40, 40, 40, 40, 40, 40, 40, 40}
var currentCol int
var currentRow int
var SyncToPlay bool
var LastNote Note = Note(37)
var uiElem *ui.Element

const (
	RowsInScreen = 14
)

func NewTrack(p *Project) *Track {
	result := Track{}
	result.ID = uint8(len(p.Tracks) + 1)
	result.IsMultisample = false
	newPhrase := NewPhrase(p)
	result.Phrases = []*Phrase{newPhrase}
	result.PhraseIds = []int32{newPhrase.ID}
	result.Sample = Sample{SampleFile: "./assets/music/key.wav", RootNote: 37}
	p.Tracks = append(p.Tracks, result)
	return &result
}

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
}

func drawTrack(ctx events.EventContext, hasFocus bool) bool {
	if !uiElem.Visible || len(CurrentProject.Tracks) < 1 {
		return false
	}

	currentTrack := CurrentProject.Tracks[Clamp(ui.TrackId, 0, len(CurrentProject.Tracks)-1)]
	currentPhrase := currentTrack.Phrases[Clamp(ui.PhraseId, 0, len(currentTrack.Phrases)-1)]

	firstVerticalItem := max(0, currentRow-RowsInScreen)
	noteColor := rl.Lime

	for i, step := range currentPhrase.Steps {
		if i < firstVerticalItem {
			continue
		} else if i > firstVerticalItem+RowsInScreen {
			break
		}

		verticalAnchor := i - firstVerticalItem
		if i%4 == 0 {
			ui.DrawText(fmt.Sprintf("%02d", int(i)), 10, int32(10+verticalAnchor*20), 20, rl.Lime)
		} else {
			ui.DrawText(fmt.Sprintf("%02d", int(i)), 10, int32(10+verticalAnchor*20), 20, rl.Red)
		}

		for j := 0; j < len(step.Notes); j++ {
			if j >= len(columnStarts) {
				break
			}
			rl.DrawText(step.Notes[j].ToString(), int32(columnStarts[j]), int32(10+verticalAnchor*20), 20, noteColor)
			if step.Notes[j] == 253 {
				noteColor = rl.Gray
			}
		}
		// rl.DrawText(fmt.Sprintf("%d", int(ui.GetOptions().ScreenWidth)), 100.0, int32(10+i*20), 20, rl.Lime)
		if currentPhrase.CurrentStep == i {
			rl.DrawText(">", 0, int32(10+verticalAnchor*20), 20, rl.Maroon)
		}
	}

	bottomOffset := int32(ui.GetOptions().ScreenHeight - 10)
	rl.DrawRectangle(0, int32(ui.GetOptions().VerticalPadding+(currentRow-firstVerticalItem)*ui.GetOptions().RowHeight), int32(ui.GetOptions().ScreenWidth), 20, ui.GetOptions().ColorHighlight)
	rl.DrawRectangle(int32(columnStarts[currentCol%len(columnStarts)]), int32(ui.GetOptions().VerticalPadding+(currentRow-firstVerticalItem)*ui.GetOptions().RowHeight), int32(columnWidths[currentCol%len(columnWidths)]), 20, ui.GetOptions().ColorHighlight)
	rl.DrawText(fmt.Sprintf("TRK: %02d/%02d", ui.TrackId, len(CurrentProject.Tracks)-1), 10, bottomOffset, 10, rl.Maroon)
	rl.DrawText(fmt.Sprintf("PHR: %02d/%02d", ui.PhraseId, len(currentTrack.Phrases)-1), 80, bottomOffset, 10, rl.Maroon)
	rl.DrawText(fmt.Sprintf("STP: %02d", currentRow), 150, bottomOffset, 10, rl.Maroon)
	rl.DrawText(fmt.Sprintf("STP: %02d", CurrentProject.Current().Current().CurrentStep), 200, bottomOffset, 10, rl.Maroon)

	return false
}

var lastSnapshot ev.InputSnapshot

func handleInputTrack(input ev.InputSnapshot, el *ui.Element) bool {
	if !uiElem.Visible || len(CurrentProject.Tracks) < 1 {
		return false
	}

	result := false
	currentTrack := &CurrentProject.Tracks[ui.TrackId]

	if input.Down(ev.InputKindDir) && input.Down(ev.InputKindA) && input.Down(ev.InputKindB) {
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
			fmt.Printf("Adding a track!\n")
			_ = NewTrack(CurrentProject)
		case input.Down(ev.InputKindLeft):
			fmt.Printf("Removing a track!\n")
			ui.TrackId--
			CurrentProject.Tracks[len(CurrentProject.Tracks)-1].Cleanup()
			CurrentProject.Tracks = CurrentProject.Tracks[:len(CurrentProject.Tracks)-1]
		}
		result = true
	}
	if input.Down(ev.InputKindA) && input.Down(ev.InputKindB) && input.Down(ev.InputKindDown) {
		if ui.PhraseId == len(currentTrack.Phrases)-1 {
			result = true
		}
	} else if input.Down(ev.InputKindA) && input.Down(ev.InputKindB) && input.Down(ev.InputKindUp) {
		if ui.PhraseId == len(currentTrack.Phrases)-1 && !player.IsPlaying {
			result = true
		}
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
		noteSlot := &currentTrack.Phrases[ui.PhraseId].Steps[Clamp(currentRow, 0, MaxStepsInPhrase)].Notes[Clamp(currentCol, 0, MaxNotesInStep)]
		if input.Down(ev.InputKindRight) {
			*noteSlot++
		} else if input.Down(ev.InputKindLeft) {
			*noteSlot--
		} else if input.Down(ev.InputKindUp) {
			*noteSlot += 12
		} else {
			*noteSlot -= 12
		}
		if *noteSlot != 0 {
			LastNote = *noteSlot
		}
		result = true
	} else if input.Down(ev.InputKindB) && (input.Tick(ev.InputKindB) > 60 || input.Down(ev.InputKindDown)) {
		_ = ui.NewTransition(ui.SongDialog, rl.NewVector2(0, 0))
		ev.RegisterCallback(ev.EventKindPostUpdate, func(ctx ev.EventContext) bool {
			el.Visible = false
			ui.SongDialog.Visible = true
			ui.RootElement.SetFocus(ui.SongDialog)
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
