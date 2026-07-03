package view

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	. "trkr"
	"trkr/internal/audio"
	ev "trkr/internal/events"
	"trkr/internal/player"
	ui "trkr/internal/ui"
)

var songColumnStarts = [16]int{}
var songColumnWidths = [16]int{}
var songCurrentCol int
var songCurrentRow int

type SongView struct {
	SettingsElement *ui.Element
}

func CreateSongView(parent *ui.Element) *ui.Element {
	core := &SongView{}
	ui.SongDialog = ui.NewElement(0, 0, int32(ui.GetOptions().ScreenWidth), int32(ui.GetOptions().ScreenHeight), core, parent)
	ui.SongDialog.Visible = false
	lastStart := 0
	for i := range songColumnStarts {
		if i == 0 {
			songColumnWidths[i] = 80
		} else {
			songColumnWidths[i] = 40
		}
		songColumnStarts[i] = lastStart
		lastStart += songColumnWidths[i]
	}

	return ui.SongDialog
}

func (pv *SongView) Show() {}
func (pv *SongView) Hide() {}
func (pv *SongView) Draw(ctx ev.EventContext, hasFocus bool) bool {
	currentTrack := CurrentProject.Tracks[ui.TrackId]
	uiElem := ctx.EventPayload.(*ui.ElementDrawPayload).Element
	if !uiElem.Visible {
		return false
	}
	firstVerticalItem := max(0, songCurrentRow-RowsInScreen)

	verticalAnchor := 0
	for i := range songColumnStarts {
		if i == 0 {
			ui.DrawText("Cycles", 10, int32(10+verticalAnchor*20), 20, rl.Red)
		} else {
			ui.DrawText(fmt.Sprintf("T%02d", int(i)), int32(songColumnStarts[i]), int32(10+verticalAnchor*20), 20, rl.Red)
		}
	}

	for i := range CurrentProject.Tracks {
		for j, p := range CurrentProject.Tracks[i].Phrases {
			verticalAnchor = (1 + j) - firstVerticalItem
			ui.DrawText(fmt.Sprintf("%02d", p.ID),
				int32(songColumnStarts[Clamp(i+1, 0, len(songColumnStarts)-1)]),
				int32(10+verticalAnchor*20), 20, rl.Lime)
		}
	}

	bottomOffset := int32(ui.GetOptions().ScreenHeight - 10)

	rl.DrawRectangle(
		int32(songColumnStarts[songCurrentCol%len(songColumnStarts)]),
		int32(ui.GetOptions().VerticalPadding+(songCurrentRow-firstVerticalItem)*ui.GetOptions().RowHeight),
		int32(songColumnWidths[songCurrentCol%len(songColumnWidths)]),
		20, rl.Color{0, 255, 0, 64})

	currentPhrase := currentTrack.Phrases[ui.PhraseId]

	laid := ctx.EventPayload.(*ui.ElementDrawPayload).Laid

	bounds := rl.NewRectangle(10, float32(rl.GetScreenHeight()/2), float32(rl.GetScreenWidth()/2), float32(rl.GetScreenHeight()/2))

	DrawMiniPhrase(bounds, currentPhrase)

	bounds.X = float32(rl.GetScreenWidth() / 2)
	laid.PushContext(bounds)
	laid.Pad(10, 10)
	laid.TextBlock(fmt.Sprintf("Track %d", ui.TrackId), 16, rl.White)
	if CurrentProject.Tracks[ui.TrackId].Instrument != nil {
		laid.Pad(10, 0)
		laid.TextBlock(fmt.Sprintf("Volume: %0.2f", currentTrack.Volume*100.0), 16, rl.White)
		if currentTrack.Skips > 0 {
			laid.TextBlock(fmt.Sprintf("Speed: 1/%d", currentTrack.Skips), 16, rl.White)
		} else {
			laid.TextBlock("Speed: Normal", 16, rl.White)
		}
		laid.TextBlock(fmt.Sprintf("Phrase: %d", currentPhrase.ID), 16, rl.White)
		laid.Pad(10, 0)
		laid.TextBlock(fmt.Sprintf("Repeats: %d(%d)", currentPhrase.Repeats, (currentPhrase.Repeats-currentPhrase.CurrentRepeat)), 16, rl.White)
		laid.Pad(-10, 0)
		laid.TextBlock(CurrentProject.Tracks[ui.TrackId].Instrument.SampleSourceType.UiString(), 16, rl.White)
		switch CurrentProject.Tracks[ui.TrackId].Instrument.SampleSourceType {
		case SampleSourceTypeFm:
			laid.TextBlock(CurrentProject.FmPatchName, 16, rl.White)
			laid.TextBlock(fmt.Sprintf("%d: %s", currentTrack.Instrument.Program, audio.SynthProgramName(currentTrack.Id)), 16, rl.White)
		}
	}
	laid.PushContext(bounds)

	rl.DrawText(fmt.Sprintf("TRK: %02d/%02d", ui.TrackId, len(CurrentProject.Tracks)-1), 10, bottomOffset, 10, rl.Maroon)
	rl.DrawText(fmt.Sprintf("PHR: %02d/%02d", ui.PhraseId, len(currentTrack.Phrases)-1), 80, bottomOffset, 10, rl.Maroon)
	rl.DrawText(fmt.Sprintf("STP: %02d", songCurrentRow), 150, bottomOffset, 10, rl.Maroon)
	rl.DrawText(fmt.Sprintf("STP: %02d", CurrentProject.Current().Current().CurrentStep), 200, bottomOffset, 10, rl.Maroon)

	return false
}

func (pv *SongView) HandleInput(input *ev.InputSnapshot, el *ui.Element) bool {

	if songCurrentRow < 1 && input.Down(ev.InputKindA) {
		showSettings()
		return true
	} else if input.Down(ev.InputKindA) || (input.Down(ev.InputKindB) && input.Down(ev.InputKindUp)) {
		fmt.Printf("Current row: %d\n", currentRow)
		_ = ui.NewTransition(ui.TrackDialog, rl.NewVector2(1, 1))
		ev.RegisterCallback(ev.EventKindPostUpdate, func(ctx ev.EventContext) bool {
			ui.SongDialog.Visible = false
			ui.TrackDialog.Visible = true
			ui.RootElement.SetFocus(ui.TrackDialog)
			return true
		}, el.ID)
		return true
	} else if input.Down(ev.InputKindDir) {
		movementMultiplier := int(Clamp(input.Tick(ev.InputKindDir), 1, 10))
		if input.Down(ev.InputKindDown) {
			songCurrentRow = Clamp(songCurrentRow+movementMultiplier, 0, MaxStepsInPhrase-1)
		} else if input.Down(ev.InputKindUp) {
			songCurrentRow = Clamp(songCurrentRow-movementMultiplier, 0, MaxStepsInPhrase-1)
		} else if input.Down(ev.InputKindLeft) {
			songCurrentCol = Clamp(songCurrentCol-1, 0, len(CurrentProject.Tracks)+1)
		} else if input.Down(ev.InputKindRight) {
			songCurrentCol = Clamp(songCurrentCol+1, 0, len(CurrentProject.Tracks)+1)
		}

		ui.TrackId = Clamp(songCurrentCol-1, 0, len(CurrentProject.Tracks)-1)
		ui.PhraseId = Clamp(songCurrentRow-1, 0, len(CurrentProject.Tracks[ui.TrackId].Phrases)-1)

		return true
	} else if input.Down(ev.InputKindSpace) && !player.IsPlaying {
		ev.RegisterCallback(ev.EventKindPostUpdate, func(ctx ev.EventContext) bool {
			if !player.IsPlaying {
				go player.Play()
			} else {
				player.Stop()
			}
			return true
		}, player.GetElem().ID)
		return true
	}
	return false
}

func DrawMiniPhrase(b rl.Rectangle, p *Phrase) {
	b.Height = 4
	b.Width = 10
	originalX := b.X
	for _, s := range p.Steps {
		for _, n := range s.Notes {
			if n != NoteNone {
				rl.DrawRectangleRec(b, NoteColor[n%12]) //  DrawPixel(int32(b.X)+int32(x*2), int32(b.Y)+int32(y*2), rl.Green)
			} else {
				rl.DrawRectangleRec(b, rl.Gray) //  DrawPixel(int32(b.X)+int32(x*2), int32(b.Y)+int32(y*2), rl.Green)
			}
			b.X += 12
		}
		b.Y += 6
		b.X = originalX
	}
}
