package track

import (
	"fmt"
	. "trkr"
	ev "trkr/internal/events"
	"trkr/internal/player"
	ui "trkr/internal/ui"
	"trkr/internal/views/settings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var columnStarts = []int{10, 40, 80, 120, 160, 200, 240, 280, 320}
var columnWidths = []int{30, 40, 40, 40, 40, 30}
var currentCol int
var currentRow int
var UiTrackId int
var UiPhraseId int
var SyncToPlay bool
var LastNote Note = Note(37)
var MovementLastDir ev.InputKind
var MovementLastDirRepeated int
var MovementMultiplier int = 1

const (
	RowsInScreen = 14
)

func Show() {
	ev.RegisterCallback(ev.EventKindInput, func(ctx ev.EventContext) bool {
		return handleInput(ctx.EventPayload.(ev.InputSnapshot))
	}, "Tracker Input")

	ev.RegisterCallback(ev.EventKindGuiDraw, func(ctx ev.EventContext) bool {
		return Draw()
	}, "Tracker Draw")
}

func Draw() bool {
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
			rl.DrawText(fmt.Sprintf("%02d", int(i)), int32(columnStarts[0]), int32(10+verticalAnchor*20), 20, rl.Red)
		} else {
			rl.DrawText(fmt.Sprintf("%02d", int(i)), int32(columnStarts[0]), int32(10+verticalAnchor*20), 20, rl.Lime)
		}

		for j := 0; j < len(step.Notes)-1; j++ {
			if step.Notes[j] > 0 {
				rl.DrawText(fmt.Sprintf("%s%d", Notation[(step.Notes[j]-1)%SemitonesInOctave], (step.Notes[j]-1)/12), int32(columnStarts[j+1]), int32(10+verticalAnchor*20), 20, rl.Lime)
			} else {
				rl.DrawText("--", int32(columnStarts[j+1]), int32(10+verticalAnchor*20), 20, rl.Maroon)
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

	return true
}

func handleInput(input ev.InputSnapshot) bool {
	result := false
	currentTrack := &CurrentProject.Tracks[UiTrackId]

	// Turbo
	if input[ev.InputKindDir] && input[MovementLastDir] {
		fmt.Printf("Repeat detected.\n")
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
			currentTrack.Phrases = append(currentTrack.Phrases, Phrase{})
			UiPhraseId++
			return true
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
			return true
		}
	} else if input[ev.InputKindPressedR] && input[ev.InputKindDir] {
		if input[ev.InputKindPressedLeft] {
			UiTrackId = Clamp(UiTrackId-1, 0, len(CurrentProject.Tracks)-1)
			UiPhraseId = 0
		} else if input[ev.InputKindPressedRight] {
			UiTrackId = Clamp(UiTrackId+1, 0, len(CurrentProject.Tracks)-1)
			UiPhraseId = 0
		} else if input[ev.InputKindPressedUp] {
			UiPhraseId = Clamp(UiPhraseId-1, 0, len(CurrentProject.Tracks[UiTrackId].Phrases)-1)
		} else if input[ev.InputKindPressedDown] {
			UiPhraseId = Clamp(UiPhraseId+1, 0, len(CurrentProject.Tracks[UiTrackId].Phrases)-1)
		}
		currentRow = 0
		currentCol = 0
		result = true
	} else if input[ev.InputKindPressedA] && input[ev.InputKindDir] {
		noteSlot := &currentTrack.Phrases[UiPhraseId].Steps[Clamp(currentRow, 0, 31)].Notes[Clamp(currentCol-1, 0, 4)]
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
		settings.Show()
		result = true
	}

	return result
}
