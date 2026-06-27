package player

import (

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
		audio.PlaySoundMulti(0, 0, NoteCut, 0, 0, 0.0)
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
	if audio.VoiceFm != nil {
		audio.VoiceFm.Play()
	}
	audio.PlayFmPickup()

	ev.RegisterCallback(ev.EventKindTick, func(ctx ev.EventContext) bool {
		for trackId := range CurrentProject.Tracks {

			track := &CurrentProject.Tracks[trackId]

			currentTick := *(ctx.EventPayload.(*int64))

			if track.Skips > 0 && currentTick%int64(track.Skips) != 0 {
				continue
			}

			phrase := track.Current()
			var velocity uint8 = 100
			if phrase.CurrentStep%4 == 0 {
				velocity = 127
			}
			//			if phrase

		NoteLoop:
			for columnId, note := range phrase.Current().Notes {
				switch note {
				case NoteNone:
					continue
				case NoteSkip:
					phrase.CurrentStep = len(phrase.Steps) - 1
					break NoteLoop
				default:
					instrument := track.Instrument
					var sampleSource SampleSourceType
					if instrument != nil {
						sampleSource = instrument.SampleSourceType
					} else {
						sampleSource = SampleSourceTypeFm
					}
					audio.PlaySoundMulti(uint8(columnId), uint8(trackId), note, track.InstrumentId, sampleSource, velocity)
				}
			}

			phrase.CurrentStep++
			if phrase.CurrentStep > len(phrase.Steps)-1 {
				if phrase.Repeats > 0 && phrase.CurrentRepeat < phrase.Repeats {
					phrase.CurrentRepeat++
				} else {
					track.CurrentPhrase++
					if track.CurrentPhrase > len(track.Phrases)-1 {
						track.CurrentPhrase = 0
					}
					phrase = track.Current()
					phrase.CurrentRepeat = 0
				}
				phrase.CurrentStep = 0
			}

		}
		return true
	}, uiElem.ID)

	var tickerPause time.Duration = time.Minute / time.Duration((BeatsPerMinute)*4.0)
	IsPlaying = true

	anchorTime := time.Now()
	var tickCount int64 = 0
	tickContext := ev.EventContext{EventPayload: &tickCount}
	for {
		// The playback engine dictates the current speed/subdivision dynamically
		ticksPerPatternRow := 4 // engine.GetTicksPerRow()
		isRowChange := (tickCount%int64(ticksPerPatternRow) == 0)

		if isRowChange {
			// SACRED STEP: Row alignment (e.g., Note On, Instrument Swap)
			// Precise alignment using our "gravel nap + spin"
			targetTime := anchorTime.Add(time.Duration(tickCount) * tickerPause)
			coarseSleep := time.Until(targetTime) - (2 * time.Millisecond)
			if coarseSleep > 0 {
				time.Sleep(coarseSleep)
			}
			for time.Now().Before(targetTime) {
				time.Sleep(0)
			}
		} else {
			// FLEXIBLE STEP: Inside the row (e.g., Arpeggio tick, Volume Slide)
			// Loose, low-overhead sleep that absorbs the slack
			time.Sleep(tickerPause)
		}
		// Trigger the tick and let the engine figure out what to do with it
		ev.Trigger(ev.EventKindTick, tickContext)

		select {
		case <-StopChannel:
			IsPlaying = false
			return
		default:
		}

		tickCount++
	}
}
