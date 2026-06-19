package player

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	// "golang.org/x/exp/rand"
	"fmt"
	"time"
	. "trkr"
	"trkr/internal/audio"
	ev "trkr/internal/events"
	"trkr/internal/ui"
)

var StopChannel chan bool
var IsPlaying bool
var uiElem *ui.Element

func CreatePlayer(parent *ui.Element) {
	uiElem = ui.NewElement(0, 0, 0, 0, nil, parent)
	uiElem.Visible = false
}

func GetElem() *ui.Element {
	return uiElem
}

func Stop() {
	fmt.Printf("Player is stopping...\n")
	if IsPlaying {
		audio.PlaySoundMulti(0, 0, NoteCut)
		StopChannel <- true
		ev.ClearCallbacks(ev.EventKindTick)
	}
}

func Play() {
	if IsPlaying {
		return
	}
	if uiElem == nil {
		uiElem = ui.NewElement(0, 0, 0, 0, nil, nil)
	}

	StopChannel = make(chan bool)
	for trackId := range CurrentProject.Tracks {
		if !CurrentProject.Tracks[trackId].IsMultisample {
			audio.InitializeAliases(trackId, &CurrentProject.Tracks[trackId])
		} else {
			for sampleId := range CurrentProject.Tracks[trackId].Samples {
				CurrentProject.Tracks[trackId].Samples[sampleId].Sound = rl.LoadSound(CurrentProject.Tracks[trackId].Samples[sampleId].SampleFile)
			}
		}
	}

	ev.RegisterCallback(ev.EventKindTick, func(ctx ev.EventContext) bool {
		for trackId := range CurrentProject.Tracks {
			track := &CurrentProject.Tracks[trackId]
			phrase := track.Current()
			phrase.CurrentStep++
			if phrase.CurrentStep > len(phrase.Steps)-1 {
				phrase.CurrentStep = 0
				track.CurrentPhrase++
				if track.CurrentPhrase > len(track.Phrases)-1 {
					track.CurrentPhrase = 0
				}
				phrase = track.Current()
				phrase.CurrentStep = 0
			}

			for columnId, note := range phrase.Current().Notes {
				if note == NoteNone {
					continue
				}
				if note == NoteSkip {
					phrase.CurrentStep = 31
					break
				}
				audio.PlaySoundMulti(uint8(columnId), uint8(trackId), note)
			}
		}
		return true
	}, uiElem.ID)

	var tickerPause time.Duration = time.Minute / time.Duration((BeatsPerMinute)*4.0)
	IsPlaying = true
	timer := time.NewTimer(tickerPause)
	defer timer.Stop() // Clean up when the playback loop terminates

	for {
		ev.Trigger(ev.EventKindTick, ev.EventContext{})

		select {
		case <-timer.C:
			timer.Reset(tickerPause)
			continue

		case <-StopChannel:
			IsPlaying = false
			return
		}
	}
}
