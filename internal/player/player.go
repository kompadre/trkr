package player

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"golang.org/x/exp/rand"
	"math"
	"time"
	. "trkr"
	"trkr/internal/audio"
	ev "trkr/internal/events"
)

var StopChannel chan bool
var IsPlaying bool

func Play() {
	StopChannel = make(chan bool)
	audio.Init()
	defer audio.CleanUp()
	ev.RegisterCallback(ev.EventKindTick, func(ctx ev.EventContext) bool {
		var sound *rl.Sound
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
			for noteId, note := range phrase.Current().Notes {
				if note == 0 {
					continue
				}
				if trackId > 0 {
					// fmt.Printf("Phrase: %d, Step: %d, Note: %d\n", track.CurrentPhrase, phrase.CurrentStep, note)
				}
				if track.IsMultisample {
					if int(note) > len(track.Samples)-1 {
						panic(fmt.Errorf("Missing sample for note %d.\n", note))
					}
					// fmt.Printf("Playing %s.\n", track.Samples[note].SampleFile)
					sound = &track.Samples[note].Sound
					if sound.FrameCount == 0 {
						track.Samples[note].Sound = rl.LoadSound(track.Samples[note].SampleFile)
						sound = &track.Samples[note].Sound
					}
				} else {
					if noteId > len(track.Samples)-1 {
						to := len(track.Samples) - 1
						for i := noteId; i > to; i-- {
							track.Samples = append(track.Samples, Sample{SampleFile: track.Samples[0].SampleFile, RootNote: track.Samples[0].RootNote})
						}
					}
					sound = &track.Samples[noteId].Sound
					if sound.FrameCount == 0 {
						track.Samples[0].Sound = rl.LoadSound(track.Samples[0].SampleFile)
						sound = &track.Samples[0].Sound
					}
					rl.StopSound(*sound)

					pitch := 1.0 * math.Pow(2.0, float64(note-track.Samples[noteId].RootNote)/12)
					fmt.Printf("Pitch: %f\n", note-track.Samples[noteId].RootNote)
					rl.SetSoundPitch(*sound, float32(pitch))
					// rl.SetSoundVolume(*sound, 0.5+(rand.Float32()*0.5))
				}
				rl.SetSoundVolume(*sound, 0.5+(rand.Float32()*0.5))
				go audio.PlaySound(*sound)
			}
		}

		return true
	})

	tickerPause := (time.Minute / BeatsPerMinute) / 4
	IsPlaying = true
	// ticker := time.NewTicker(tickerPause)
	for {
		//newTs := time.Now()
		go ev.Trigger(ev.EventKindTick, ev.EventContext{})
		select {
		case <-time.After(tickerPause):
			continue
		case <-StopChannel:
			IsPlaying = false
			return
		}
	}
}

//
