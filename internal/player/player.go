package player

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	// "golang.org/x/exp/rand"
	"time"
	. "trkr"
	"trkr/internal/audio"
	ev "trkr/internal/events"
)

var StopChannel chan bool
var IsPlaying bool

func Stop() {
	if IsPlaying {
		StopChannel <- true
		ev.ClearCallbacks(ev.EventKindTick)
	}
}

func Play() {
	if IsPlaying {
		return
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

			for _, note := range phrase.Current().Notes {
				if note == 0 {
					continue
				}

				if !track.IsMultisample {
					pitch := float32(audio.GetPitch(note, track.Sample.RootNote))
					audio.PlaySoundMulti(trackId, pitch)
				} else {
					//if note == 2 {
					//rl.SetSoundVolume(track.Samples[note-1].Sound, rand.Float32())
					//}
					audio.PlaySound(track.Samples[Clamp(int(note-1), 0, len(track.Samples)-1)].Sound)
				}
			}
		}
		return true
	}, "Music Tick")

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
