package player

import (
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
var EffectiveBpm float64

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
		audio.PlaySoundMulti(0, 0, NoteCut, 0, 0, 0.0, 0.0)
		audio.StopFm()
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
			phrase := track.Current()

			if currentTick < 0 {
				// Dry run
				for range phrase.Current().Notes {
				}
				continue
			}

			if track.Skips > 0 && currentTick%int64(track.Skips) > 0 {
				continue
			}

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
					audio.PlaySoundMulti(uint8(columnId), uint8(trackId), note, track.InstrumentId, sampleSource, velocity, track.Volume)
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

	var tickCount int64 = 0
	tickContext := ev.EventContext{EventPayload: &tickCount}

	var (
		tickStart   time.Time
		targetSleep time.Duration
		sleepStart  time.Time
		sleepDrift  time.Duration
	)

	tickStart = time.Now()
	firstStart := tickStart
	targetSleep = tickerPause
	for {
		ev.Trigger(ev.EventKindTick, tickContext)
		targetSleep -= time.Since(tickStart)
		if targetSleep > 0 {
			sleepStart = time.Now()
			time.Sleep(targetSleep)
			sleepDrift = targetSleep - time.Since(sleepStart)
		} else {
			Logf("Beat being skipped.\n")
			sleepDrift = 0
		}

		select {
		case <-StopChannel:
			IsPlaying = false
			return
		default:
		}

		tickCount++
		targetSleep = (tickerPause - sleepDrift)
		tickStart = tickStart.Add(tickerPause)
		EffectiveBpm = float64(tickCount) / (time.Since(firstStart).Minutes() * 4)
	}
}
