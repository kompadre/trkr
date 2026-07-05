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

type PhraseRuntime struct {
	CurrentRow     *Step
	RowCounter     int
	RepeatsCounter uint8
}

type TrackRuntime struct {
	TrackId         uint8
	PhraseCounter   uint8
	CurrentPhraseId int32
	CurrentPhrase   *Phrase
	Phrase          PhraseRuntime
}

type SectionRuntime struct {
	TotalRowsCounter int32
	Tracks           [MaxTracks]TrackRuntime
}

type Playhead struct {
	CurrentSectionId uint8
	CurrentSection   *Section
	Section          SectionRuntime
}

func NewPlayhead() Playhead {
	var currentSection uint8 = 0
	p := Playhead{CurrentSectionId: currentSection, CurrentSection: &CurrentProject.Sections[currentSection]}
	for i, t := range CurrentProject.Tracks {
		p.Section.Tracks[i].TrackId = t.Id
		p.Section.Tracks[i].CurrentPhrase = &CurrentProject.Phrases[t.PhraseIds[currentSection][0]]
		p.Section.Tracks[i].CurrentPhraseId = p.Section.Tracks[i].CurrentPhrase.ID
	}
	return p
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
	playhead := NewPlayhead()
	ev.RegisterCallback(ev.EventKindTick, func(ctx ev.EventContext) bool {
		for i := range playhead.Section.Tracks {
			trackRuntime := &playhead.Section.Tracks[i]
			phraseRuntime := &trackRuntime.Phrase
			track := &CurrentProject.Tracks[trackRuntime.TrackId]
			trackId := trackRuntime.TrackId

			currentTick := *(ctx.EventPayload.(*int64))
			phrase := &CurrentProject.Phrases[track.PhraseIds[playhead.CurrentSectionId][trackRuntime.PhraseCounter]]

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

			phraseRuntime.RowCounter++
			if phraseRuntime.RowCounter > len(phrase.Steps)-1 {
				if phrase.Repeats > 0 && phraseRuntime.RepeatsCounter < phrase.Repeats {
					Logf("RepeatsCounter updated.\n")
					phraseRuntime.RepeatsCounter++
				} else {
					trackRuntime.PhraseCounter++
					if trackRuntime.PhraseCounter > uint8(len(track.PhraseIds[playhead.CurrentSectionId])-1) {
						trackRuntime.PhraseCounter = 0
					}
					trackRuntime.Phrase = PhraseRuntime{}
					trackRuntime.CurrentPhraseId = track.PhraseIds[playhead.CurrentSectionId][trackRuntime.PhraseCounter]
					phrase = &CurrentProject.Phrases[trackRuntime.CurrentPhraseId]
				}
				phraseRuntime.RowCounter = 0
			}
			phrase.CurrentStep = phraseRuntime.RowCounter
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
