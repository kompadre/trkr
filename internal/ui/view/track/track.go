package track

import (
	"fmt"
	. "trkr"
	"trkr/internal/events"
	ev "trkr/internal/events"
	"trkr/internal/player"
	ui "trkr/internal/ui"
	"trkr/internal/ui/view/settings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var columnStarts = []int{40, 80, 120, 160, 200, 240, 280, 320}
var columnWidths = []int{40, 40, 40, 40, 40, 40, 40, 40}
var currentCol int
var currentRow int
var UiTrackId int
var UiPhraseId int
var SyncToPlay bool
var LastNote Note = Note(37)
var MovementLastDir ev.InputKind
var MovementLastDirRepeated int
var MovementMultiplier int = 1
var uiElem *ui.Element

const (
	RowsInScreen = 14
)

func NewTrack(id uint8) *Track {
	result := &Track{}
	result.Id = id
	result.IsMultisample = false
	result.Phrases = make([]Phrase, 1)
	return result
}

func Show() {
	uiElem.Visible = true
}

func Hide() {
	uiElem.Visible = false
}

func GetElement() *ui.Element {
	return uiElem
}

func Create(rootElement *ui.Element) {
	core := ui.NewElementCoreInstance(Show, Hide, handleInput, Draw)
	uiElem = ui.NewElement(10, 10, int32(ui.GetOptions().ScreenWidth), int32(ui.GetOptions().ScreenHeight), core, rootElement)
}

func Draw(ctx events.EventContext, hasFocus bool) bool {
	if !uiElem.Visible {
		return false
	}
	var currentTrack Track
	var currentPhrase Phrase

	if SyncToPlay {
		currentTrack = CurrentProject.Tracks[CurrentProject.CurrentTrack]
		currentPhrase = currentTrack.Phrases[currentTrack.CurrentPhrase]
	} else {
		currentTrack = CurrentProject.Tracks[UiTrackId]
		currentPhrase = currentTrack.Phrases[UiPhraseId]
	}

	firstVerticalItem := max(0, currentRow-RowsInScreen)
	for i, step := range currentPhrase.Steps {
		if i < firstVerticalItem {
			continue
		} else if i > firstVerticalItem+RowsInScreen {
			break
		}

		verticalAnchor := i - firstVerticalItem
		if i%4 == 0 {
			ui.DrawText(fmt.Sprintf("%02d", int(i)), 10, int32(10+verticalAnchor*20), 20, rl.Red)
		} else {
			ui.DrawText(fmt.Sprintf("%02d", int(i)), 10, int32(10+verticalAnchor*20), 20, rl.Lime)
		}

		for j := 0; j < len(step.Notes); j++ {
			if j >= len(columnStarts) {
				break
			}
			if step.Notes[j] > 0 {
				rl.DrawText(fmt.Sprintf("%s%d", Notation[(step.Notes[j]-1)%SemitonesInOctave], (step.Notes[j]-1)/12), int32(columnStarts[j]), int32(10+verticalAnchor*20), 20, rl.Lime)
			} else {
				rl.DrawText("--", int32(columnStarts[j])+10, int32(10+verticalAnchor*20), 20, rl.Maroon)
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
	rl.DrawText(fmt.Sprintf("TRK: %02d/%02d", UiTrackId, len(CurrentProject.Tracks)-1), 10, bottomOffset, 10, rl.Maroon)
	rl.DrawText(fmt.Sprintf("PHR: %02d/%02d", UiPhraseId, len(currentTrack.Phrases)-1), 80, bottomOffset, 10, rl.Maroon)
	rl.DrawText(fmt.Sprintf("STP: %02d", currentRow), 150, bottomOffset, 10, rl.Maroon)
	rl.DrawText(fmt.Sprintf("STP: %02d", CurrentProject.Current().Current().CurrentStep), 200, bottomOffset, 10, rl.Maroon)

	return false
}

func handleInput(input ev.InputSnapshot, el *ui.Element) bool {
	if !uiElem.Visible {
		return false
	}

	result := false
	currentTrack := &CurrentProject.Tracks[UiTrackId]

	// Turbo
	if input[ev.InputKindDir] && input[MovementLastDir] {
		MovementLastDirRepeated++
		if MovementLastDirRepeated > 8 {
			MovementMultiplier = 5
		} else if MovementLastDirRepeated > 3 {
			MovementMultiplier = 3
		}
	} else {
		MovementMultiplier = 1
		MovementLastDirRepeated = 0
	}

	if input[ev.InputKindPressedA] && input[ev.InputKindPressedB] && input[ev.InputKindPressedDown] {
		if UiPhraseId == len(currentTrack.Phrases)-1 {
			fmt.Printf("Adding a new phrase!\n")
			if input[ev.InputKindPressedR] {
				currentTrack.Phrases = append(currentTrack.Phrases, CurrentProject.Tracks[UiTrackId].Phrases[UiPhraseId].Clone())
			} else {
				currentTrack.Phrases = append(currentTrack.Phrases, Phrase{})
			}
			UiPhraseId = len(currentTrack.Phrases) - 1
			result = true
		}
	} else if input[ev.InputKindPressedA] && input[ev.InputKindPressedB] && input[ev.InputKindPressedUp] {
		if UiPhraseId == len(currentTrack.Phrases)-1 && !player.IsPlaying {
			fmt.Printf("Removing a phrase!\n")
			currentTrack.Phrases = currentTrack.Phrases[:len(currentTrack.Phrases)-1]
			ResetHead()
			if UiPhraseId == 0 {
				currentTrack.Phrases = append(currentTrack.Phrases, Phrase{})
			} else {
				UiPhraseId--
			}
			result = true
		}
	} else if input[ev.InputKindPressedR] && input[ev.InputKindDir] {
		tv := rl.NewVector2(0, 0)
		oldTrackId, oldPhraseId := UiTrackId, UiPhraseId
		if input[ev.InputKindPressedLeft] {
			newTrackId := Clamp(UiTrackId-1, 0, len(CurrentProject.Tracks)-1)
			if newTrackId != oldTrackId {
				tv.X = 20
				_ = ui.NewTransition(el, tv)
				UiTrackId = Clamp(UiTrackId-1, 0, len(CurrentProject.Tracks)-1)
				UiPhraseId = 0
			}
		} else if input[ev.InputKindPressedRight] {
			newTrackId := Clamp(UiTrackId+1, 0, len(CurrentProject.Tracks)-1)
			if newTrackId != oldTrackId {
				tv.X = -20
				_ = ui.NewTransition(el, tv)
				UiTrackId = Clamp(UiTrackId+1, 0, len(CurrentProject.Tracks)-1)
				UiPhraseId = 0
			}
		} else if input[ev.InputKindPressedUp] {
			newPhraseId := Clamp(UiPhraseId-1, 0, len(CurrentProject.Tracks[UiTrackId].Phrases)-1)
			if newPhraseId != oldPhraseId {
				tv.Y = 20
				_ = ui.NewTransition(el, tv)
				UiPhraseId = newPhraseId
			}
		} else if input[ev.InputKindPressedDown] {
			newPhraseId := Clamp(UiPhraseId+1, 0, len(CurrentProject.Tracks[UiTrackId].Phrases)-1)
			if newPhraseId != oldPhraseId {
				tv.Y = -20
				_ = ui.NewTransition(el, tv)
				UiPhraseId = newPhraseId
			}
		}
		currentRow = 0
		currentCol = 0
		result = true
	} else if input[ev.InputKindPressedA] && input[ev.InputKindDir] {
		noteSlot := &currentTrack.Phrases[UiPhraseId].Steps[Clamp(currentRow, 0, 31)].Notes[Clamp(currentCol, 0, 4)]
		if *noteSlot == 0 {
			*noteSlot = LastNote
		} else if input[ev.InputKindPressedRight] {
			*noteSlot++
		} else if input[ev.InputKindPressedLeft] {
			*noteSlot--
		} else if input[ev.InputKindPressedUp] {
			*noteSlot += 12
		} else {
			*noteSlot -= 12
		}
		if *noteSlot != 0 {
			LastNote = *noteSlot
		}
		result = true
	} else if input[ev.InputKindPressedB] && input[ev.InputKindDir] {
		if input[ev.InputKindPressedUp] {
			_ = ui.NewTransition(el, rl.NewVector2(1, 1))
		} else if input[ev.InputKindPressedDown] {
			_ = ui.NewTransition(el, rl.NewVector2(0, 0))
		}
	} else if input[ev.InputKindDir] {
		if input[ev.InputKindPressedDown] {
			currentRow = Clamp(currentRow+MovementMultiplier, 0, MaxStepsInPhrase-1)
			MovementLastDir = ev.InputKindPressedDown
		} else if input[ev.InputKindPressedUp] {
			currentRow = Clamp(currentRow-MovementMultiplier, 0, MaxStepsInPhrase-1)
			MovementLastDir = ev.InputKindPressedUp
		} else if input[ev.InputKindPressedLeft] {
			currentCol = Clamp(currentCol-MovementMultiplier, 0, len(columnWidths)-1)
			MovementLastDir = ev.InputKindPressedLeft
		} else if input[ev.InputKindPressedRight] {
			currentCol = Clamp(currentCol+MovementMultiplier, 0, len(columnWidths)-1)
			MovementLastDir = ev.InputKindPressedRight
		}
		result = true
	} else if input[ev.InputKindPressedSpace] {
		if player.IsPlaying {
			player.Stop()
		} else {
			ResetHead()
			currentRow = 0
			go player.Play()
		}
		result = true
	} else if input[ev.InputKindPressedEnter] {
		if !ui.SettingsDialog.Visible {
			settings.Show()
			result = true
		}
	} else if player.IsPlaying && (input[ev.InputKindPressedA] || input[ev.InputKindPressedB]) {
		player.Stop()
		result = true
	} else if input[ev.InputKindPressedB] {
		noteSlot := &currentTrack.Phrases[UiPhraseId].Steps[Clamp(currentRow, 0, 31)].Notes[Clamp(currentCol, 0, 4)]
		if *noteSlot != 0 {
			LastNote = *noteSlot
		}
		*noteSlot = 0
		result = true
	}

	return result
}
