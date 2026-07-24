package player

import (
	"fmt"
	"time"
	. "trkr"
	"trkr/internal/audio"
	"trkr/internal/audio/effects"
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
	TrackId         int8
	PhraseCounter   uint8
	CurrentPhraseId int32
	CurrentPhrase   *Phrase
	Phrase          PhraseRuntime
}

func (tr TrackRuntime) RepeatsLeft(phraseSlot int) uint8 {
	phrase := CurrentProject.Tracks[tr.TrackId].Phrases[Head.CurrentSectionId][phraseSlot]

	if uint8(phraseSlot) == tr.PhraseCounter {
		return min(phrase.Repeats, max(0, phrase.Repeats-tr.Phrase.RepeatsCounter))
	}
	return phrase.Repeats
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

var Head Playhead

func NewPlayhead(currentSection uint8) Playhead {
	Logf("NewPlayhead for %d.\n", currentSection)
	p := Playhead{CurrentSectionId: currentSection, CurrentSection: &CurrentProject.Sections[currentSection]}
	for i, t := range CurrentProject.Tracks {
		if len(CurrentProject.Tracks[i].Phrases[currentSection]) > 0 {
			p.Section.Tracks[i].TrackId = int8(t.Id)
			p.Section.Tracks[i].CurrentPhrase = &CurrentProject.Phrases[t.PhraseIds[currentSection][0]]
			p.Section.Tracks[i].CurrentPhraseId = p.Section.Tracks[i].CurrentPhrase.ID
		} else {
			p.Section.Tracks[i].TrackId = -1
		}
	}
	return p
}

func Stop() {
	fmt.Printf("Player is stopping...\n")
	if IsPlaying {
		audio.PlaySoundMulti(0, 0, NoteCut, 0, 0, 0.0, 0.0, 0.0)
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
	Head = NewPlayhead(0)
	ev.RegisterCallback(ev.EventKindTick, func(ctx ev.EventContext) bool {
		for i := range CurrentProject.Tracks {
			if len(CurrentProject.Tracks[i].Phrases[Head.CurrentSectionId]) < 1 {
				continue
			}

			trackRuntime := &Head.Section.Tracks[i]
			phraseRuntime := &trackRuntime.Phrase
			track := &CurrentProject.Tracks[trackRuntime.TrackId]
			trackId := trackRuntime.TrackId

			currentTick := *(ctx.EventPayload.(*int64))
			phrase := &CurrentProject.Phrases[trackRuntime.CurrentPhraseId]

			if currentTick < 0 {
				// Dry run
				for range phrase.Current().Notes {
				}
				continue
			}

			if track.Skips > 0 && currentTick%int64(track.Skips) > 0 {
				continue
			}

			// var velocity uint8 = 100
			// if phrase.CurrentStep%4 == 0 {
			// 	velocity = 127
			// }
			//			if phrase

		NoteLoop:

			for columnId, note := range phrase.Steps[phraseRuntime.RowCounter].Notes {
				switch note {
				case NoteNone:
					continue
				case NoteSkip:
					// phrase.CurrentStep = len(phrase.Steps) - 1
					phraseRuntime.RowCounter = len(phrase.Steps) - 1
					break NoteLoop
				default:
					instrument := track.Instrument
					var sampleSource SampleSourceType
					if instrument != nil {
						sampleSource = instrument.SampleSourceType
					} else {
						sampleSource = SampleSourceTypeFm
					}
					if sampleSource == SampleSourceTypePerc {
						Logf("Perc on %d.\n", Head.Section.TotalRowsCounter)
					}
					velocity := 1.0
					if phrase.Steps[phraseRuntime.RowCounter].Velocities[columnId] > 0 {
						velocity = Clamp(float64(phrase.Steps[phraseRuntime.RowCounter].Velocities[columnId])/100.0, 0.0, 1.0)
					}
					audio.PlaySoundMulti(uint8(columnId), uint8(trackId), note, track.InstrumentId, sampleSource, 255, track.Volume*velocity, 0.0)
					/*
						if trackId == 0 {
							audio.PlaySoundMulti(uint8(columnId), uint8(trackId), note, track.InstrumentId, sampleSource, 255, track.Volume*velocity*0.75, 0.25)
							audio.PlaySoundMulti(uint8(columnId), uint8(trackId), note, track.InstrumentId, sampleSource, 255, track.Volume*velocity*0.5, 0.5)
							audio.PlaySoundMulti(uint8(columnId), uint8(trackId), note, track.InstrumentId, sampleSource, 255, track.Volume*velocity*0.25, 0.75)
						}
					*/
					// Logf("Queued %d %d %v %d %s", columnId, trackId, note, track.InstrumentId, sampleSource.UiString())
				}
			}
			phraseRuntime.RowCounter++
			if phraseRuntime.RowCounter > len(phrase.Steps)-1 {
				if phrase.Repeats > 0 && phraseRuntime.RepeatsCounter < phrase.Repeats {
					Logf("RepeatsCounter updated.\n")
					phraseRuntime.RepeatsCounter++
				} else {
					trackRuntime.PhraseCounter++
					if trackRuntime.PhraseCounter > uint8(len(track.PhraseIds[Head.CurrentSectionId])-1) {
						trackRuntime.PhraseCounter = 0
					}
					trackRuntime.Phrase = PhraseRuntime{}
					trackRuntime.CurrentPhraseId = track.PhraseIds[Head.CurrentSectionId][trackRuntime.PhraseCounter]
					phrase = &CurrentProject.Phrases[trackRuntime.CurrentPhraseId]
				}
				phraseRuntime.RowCounter = 0
			}
			phrase.CurrentStep = phraseRuntime.RowCounter
		}
		if audio.Filter.Type != effects.FilterTypeHPF {
			audio.Filter.Type = effects.FilterTypeHPF
		} else {
			audio.Filter.Type = effects.FilterTypeNone
		}

		Head.Section.TotalRowsCounter++
		if Head.Section.TotalRowsCounter >= int32(CurrentProject.Sections[Head.CurrentSectionId].Rows) {
			Head.CurrentSectionId++
			if Head.CurrentSectionId > uint8(len(CurrentProject.Sections)-1) {
				Head.CurrentSectionId = 0
			}
			Head = NewPlayhead(Head.CurrentSectionId)
		}

		return true
	}, uiElem.ID)

	audio.TickDuration = time.Minute / time.Duration((BeatsPerMinute)*4.0)
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
	targetSleep = audio.TickDuration
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
		targetSleep = (audio.TickDuration - sleepDrift)
		tickStart = tickStart.Add(audio.TickDuration)
		EffectiveBpm = float64(tickCount) / (time.Since(firstStart).Minutes() * 4)
	}
}
