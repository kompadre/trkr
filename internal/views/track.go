package track

import (
	"fmt"
	. "trkr"
	ui "trkr/internal"
	ev "trkr/internal/events"
	"trkr/internal/player"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var columnStarts = []int{10, 40, 80, 120, 160, 200, 240, 280, 320}
var columnWidths = []int{30, 40, 40, 40, 40, 30}
var currentCol int
var currentRow int

const (
	RowsInScreen = 14
)

func Show() {
	ev.RegisterCallback(ev.EventKindInput, func(ctx ev.EventContext) bool {
		return handleInput(ctx.EventPayload.(ev.InputSnapshot))
	})
	//ev.RegisterCallback(ev.EventKindTick, func(ctx ev.EventContext) bool {
	//	// currentRow++
	//	return true
	//})
	ev.RegisterCallback(ev.EventKindGuiDraw, func(ctx ev.EventContext) bool {
		return Draw()
	})
}

func Draw() bool {
	currentTrack := CurrentProject.Tracks[CurrentProject.CurrentTrack]
	currentPhrase := currentTrack.Phrases[currentTrack.CurrentPhrase]
	//if player.IsPlaying && currentPhrase.CurrentStep > -1 && currentPhrase.CurrentStep < len(currentPhrase.Steps)-1 {
	//	currentRow = currentPhrase.CurrentStep
	//}
	//rowsToSkip := currentRow
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
	rl.DrawRectangle(0, int32(ui.GetOptions().VerticalPadding+(currentRow-firstVerticalItem)*ui.GetOptions().RowHeight), int32(ui.GetOptions().ScreenWidth), 20, ui.GetOptions().ColorHighlight)
	rl.DrawRectangle(int32(columnStarts[currentCol%len(columnStarts)]), int32(ui.GetOptions().VerticalPadding+(currentRow-firstVerticalItem)*ui.GetOptions().RowHeight), int32(columnWidths[currentCol%len(columnWidths)]), 20, ui.GetOptions().ColorHighlight)
	rl.DrawText(fmt.Sprintf("TRACK:%02d", CurrentProject.CurrentTrack), int32(ui.GetOptions().ScreenWidth-120), int32(10), 10, rl.Maroon)
	rl.DrawText(fmt.Sprintf("PHRAS:%02d", currentTrack.CurrentPhrase), int32(ui.GetOptions().ScreenWidth-120), int32(20), 10, rl.Maroon)
	return true
}

func handleInput(input ev.InputSnapshot) bool {
	result := false
	currentTrack := &CurrentProject.Tracks[CurrentProject.CurrentTrack]
	currentPhrase := &currentTrack.Phrases[currentTrack.CurrentPhrase]
	currentStep := &currentPhrase.Steps[currentRow]
	if (input[ev.InputKindPressedA] || input[ev.InputKindPressedL] || input[ev.InputKindPressedM]) && input[ev.InputKindDir] {
		if input[ev.InputKindPressedRight] {
			currentStep.Notes[currentCol-1]++
		} else if input[ev.InputKindPressedLeft] {
			currentStep.Notes[currentCol-1]--
		} else if input[ev.InputKindPressedDown] {
			currentStep.Notes[currentCol-1] += 12
		} else if input[ev.InputKindPressedUp] {
			currentStep.Notes[currentCol-1] -= 12
		}
		result = true
	} else if input[ev.InputKindPressedR] && input[ev.InputKindDir] {
		if input[ev.InputKindPressedLeft] {
			CurrentProject.CurrentTrack = Clamp(CurrentProject.CurrentTrack-1, 0, len(CurrentProject.Tracks)-1)
		} else if input[ev.InputKindPressedRight] {
			CurrentProject.CurrentTrack = Clamp(CurrentProject.CurrentTrack+1, 0, len(CurrentProject.Tracks)-1)
		}
		if !player.IsPlaying {
			if input[ev.InputKindPressedUp] {
				CurrentProject.Current().CurrentPhrase = Clamp(CurrentProject.Current().CurrentPhrase-1, 0, len(CurrentProject.Current().Phrases)-1)
			} else if input[ev.InputKindPressedDown] {
				CurrentProject.Current().CurrentPhrase = Clamp(CurrentProject.Current().CurrentPhrase+1, 0, len(CurrentProject.Current().Phrases)-1)
			}
		}
		result = true
	} else if input[ev.InputKindDir] {
		if input[ev.InputKindPressedDown] {
			currentRow = Clamp(currentRow+1, 0, MaxStepsInPhrase-1)
		} else if input[ev.InputKindPressedUp] {
			currentRow = Clamp(currentRow-1, 0, MaxStepsInPhrase-1)
		} else if input[ev.InputKindPressedLeft] {
			currentCol = Clamp(currentCol-1, 0, len(columnWidths)-1)
		} else if input[ev.InputKindPressedRight] {
			currentCol = Clamp(currentCol+1, 0, len(columnWidths)-1)
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
	}
	return result
}
