package audio

import (
	"math"
	"math/rand"
	. "trkr"
	ev "trkr/internal/events"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	VoiceSampleRate = 48000.0
	VoiceBufferSize = 4096
	VoicePoolSize   = 32
)

type EnvelopeState uint8

const (
	EnvIdle EnvelopeState = iota
	EnvAttack
	EnvDecay
	EnvSustain
	EnvRelease
)

type Waveform uint8

const (
	WaveSquare Waveform = iota
	WaveSawtooth
	WaveCosine
	WaveCustom
)

type Envelope struct {
	State EnvelopeState
	Value float32 // The current instantaneous volume multiplier (0.0 to 1.0)

	// Step sizes added/subtracted per sample frame
	AttackStep   float32
	DecayStep    float32
	SustainLevel float32 // Target volume level during sustain (0.0 to 1.0)
	SustainTicks uint32
	ReleaseStep  float32
}

type Voice struct {
	ID       uint8
	TrackID  uint8
	ColumnID uint8
	Active   bool
	Envelope
	Note
	Volume           float32
	Waveform         // Embedded type. Public field named 'Waveform'
	CustomWaveBuffer []float32
	SampleLoopStart  float64
	SampleLoopEnd    float64

	_frequency float64
	_phase     float64
	_phaseStep float64
}

type VoiceCommand struct {
	TrackID      uint8
	ColumnID     uint8
	Note         Note
	Waveform     Waveform
	SustainTicks uint32
}

var voicePool []*Voice
var isMasterInitialized bool
var masterBuffer [VoiceBufferSize]float32
var masterStream rl.AudioStream
var activeVoices uint8
var freeVoices chan *Voice
var CommandQueue = make(chan VoiceCommand, 64)
var saturator func([]float32, float32)

func InitVoiceMasterStream() {
	masterStream = rl.LoadAudioStream(VoiceSampleRate, 32, 1)
	wave := rl.LoadWave("./assets/music/key.wav")
	rl.WaveFormat(&wave, 48000, 32, 1)
	CurrentProject.Tracks[0].CustomWaveSamples = rl.LoadWaveSamples(wave)
	CurrentProject.Tracks[0].SampleLoopStart, CurrentProject.Tracks[0].SampleLoopEnd =
		5363.0, 5363.0+(1312.0*2)
	// Permanently smooth out the raw data transitions at the seam
	//	SmoothSampleLoop(CurrentProject.Tracks[0].CustomWaveSamples, int(CurrentProject.Tracks[0].SampleLoopStart), int(CurrentProject.Tracks[0].SampleLoopEnd))
	Logf("BestStart: %f, BestEnd: %f.\n", CurrentProject.Tracks[0].SampleLoopStart, CurrentProject.Tracks[0].SampleLoopEnd)

	ev.RegisterCallback(ev.EventKindAudioUpdate, MasterUpdate, 1)
	saturator = NewSaturator()
	voicePool = make([]*Voice, VoicePoolSize)
	for i := range VoicePoolSize {
		voicePool[i] = &Voice{Note: 69, ID: uint8(i), Volume: 0.5, _phase: rand.Float64()}
	}
	isMasterInitialized = true
	rl.PlayAudioStream(masterStream)
}

func FindFreeVoice(columnId uint8, trackId uint8) *Voice {
	var oldestVoice, currentVoice *Voice
	var lowestEnvVal float32 = 1.0 // Track the quietest voice to steal safely

	// Pass 1: Look for a completely free voice
	for i := range voicePool {
		if voicePool[i].ColumnID == columnId && voicePool[i].TrackID == trackId {
			currentVoice = voicePool[i]
		} else if voicePool[i].Envelope.State == EnvIdle {
			return voicePool[i] // Found an empty slot instantly!
		}

		// Keep track of candidates to steal just in case Pass 1 fails.
		// We prefer stealing a voice that is already fading out (EnvRelease)
		// and has the lowest amplitude to minimize audible pops.
		if voicePool[i].Envelope.Value < lowestEnvVal {
			lowestEnvVal = voicePool[i].Envelope.Value
			oldestVoice = voicePool[i]
		}
	}

	if currentVoice != nil {
		return currentVoice
	}

	// Pass 2: Pool is oversaturated. Steal the quietest/oldest sounding voice.
	if oldestVoice != nil {
		// Cut it instantly so its internal phase/state resets cleanly
		oldestVoice.Envelope.Value = 0.0
		oldestVoice.Envelope.State = EnvIdle
		return oldestVoice
	}

	// Fallback if everything is literally max volume (fallback to first slot)
	return voicePool[0]
}

func (v *Voice) Play() {
	// For a sample with root note C3 (MIDI 48)
	Logf("Note: %d.\n", v.Note)

	// Define C3 using your tracker's native index system (C0 = 1, C1 = 13, C2 = 25, C3 = 37)
	const NativeC3 = 37

	// Inside your note-on trigger logic:
	c3RootMultiplier := pitchTable[NativeC3]
	v._phaseStep = pitchTable[v.Note] / c3RootMultiplier
	//	v._phaseStep = v._frequency / VoiceSampleRate
	v._phase = 0.0 // rand.Float64()

	// Envelope settings
	v.Envelope.AttackStep = 1.0 / (0.45 * float32(VoiceSampleRate))
	v.Envelope.DecayStep = 1.0 / (0.2 * float32(VoiceSampleRate))
	v.Envelope.SustainLevel = 0.6
	v.Envelope.ReleaseStep = 1.0 / (0.4 * float32(VoiceSampleRate))

	// The envelope is now the master switch
	v.Envelope.Value = 0.0
	v.Envelope.State = EnvAttack
}

func (v *Voice) Stop() {
	// Only release if it's currently sounding (Attack, Decay, or Sustain)
	if v.Envelope.State != EnvIdle && v.Envelope.State != EnvRelease {
		v.Envelope.State = EnvRelease
	}
}

func (v *Voice) Cut() {
	v.Envelope.State = EnvIdle
}

func (v *Voice) IsPlaying() bool {
	return v.Envelope.State != EnvIdle && v.Envelope.State != EnvRelease
}

func (v *Voice) LoadWave(path string) {
	wave := rl.LoadWave(path)
	rl.WaveFormat(&wave, VoiceSampleRate, 32, 1)
	v.CustomWaveBuffer = rl.LoadWaveSamples(wave)
	v._frequency = 48000

}

func (v *Voice) UpdateVoice(writeBuffer []float32, headroomScale float32) float32 {
	volScale := v.Volume * headroomScale
	if v.TrackID == 1 {
		volScale *= 2
		v.Waveform = WaveSquare
	}
	var maxAbs float32

	waveLen := float64(len(v.CustomWaveBuffer))

	if v.Waveform == WaveCustom {
		Logf("START: %f,%f END: %f,%f | SLOPES: %f -> %f, PHASE: %f\n",
			v.CustomWaveBuffer[int(v.SampleLoopStart)], v.SampleLoopStart,
			v.CustomWaveBuffer[int(v.SampleLoopEnd)], v.SampleLoopEnd,
			v.CustomWaveBuffer[int(v.SampleLoopStart)+1]-v.CustomWaveBuffer[int(v.SampleLoopStart)],
			v.CustomWaveBuffer[int(v.SampleLoopEnd)+1]-v.CustomWaveBuffer[int(v.SampleLoopEnd)],
			v._phase,
		)
	}

	for i := range writeBuffer {
		envVolume := v.TickEnvelope()
		if v.Envelope.State == EnvIdle {
			break
		}
		combinedVolume := volScale * envVolume
		var sample float32 = 1.0

		switch v.Waveform {
		case WaveSquare:
			if v._phase >= 0.5 {
				sample = -1.0
			}
		case WaveSawtooth:
			sample = float32((v._phase * 2.0) - 1.0)
		case WaveCosine:
			sample = float32(math.Cos(v._phase * 2.0 * math.Pi))
		case WaveCustom:
			// 1. Core Linear Interpolation (Single Step Lookup)
			idxA := int(v._phase)
			frac := float32(v._phase - float64(idxA))
			idxB := idxA + 1

			// Protect look-ahead index bounds
			if v.SampleLoopEnd != 0.0 && float64(idxB) >= v.SampleLoopEnd {
				idxB = int(v.SampleLoopStart) + (idxB - int(v.SampleLoopEnd))
			} else if idxB >= int(waveLen) {
				idxB = int(waveLen) - 1
			}

			if idxA < int(waveLen) && idxA >= 0 {
				// Linear blend kills the digital fuzz
				sample = v.CustomWaveBuffer[idxA] + frac*(v.CustomWaveBuffer[idxB]-v.CustomWaveBuffer[idxA])
			} else {
				sample = 0.0
			}
		}

		finalVolume := (combinedVolume + (0.1 * (1 + float32(v.ColumnID))))
		maxAbs = max(maxAbs, sample*finalVolume, -(sample * finalVolume))
		writeBuffer[i] += sample * finalVolume

		// UNIFIED ACCUMULATOR ADVANCE (Happens exactly once per loop iteration)
		v._phase += v._phaseStep

		// UNIFIED WRAP BEHAVIOR

		// Replace your wrapping logic with a strict modulo/truncation fit
		if v.Waveform == WaveCustom {
			if v.SampleLoopEnd != 0.0 {
				if v._phase >= v.SampleLoopEnd {
					// Calculate exact distance past the edge
					overshoot := v._phase - v.SampleLoopEnd

					// Clean snap: Force it back into the safe loop window
					v._phase = v.SampleLoopStart + math.Mod(overshoot, v.SampleLoopEnd-v.SampleLoopStart)
				}
			} else {
				if v._phase >= waveLen {
					v.Envelope.State = EnvIdle
				}
			}
		}
	}

	return maxAbs
}

var slider uint8

func MasterQueueCommands() {
	// Label the outer loop so we can break out of it from inside the switch
CommandLoop:
	for {
		select {
		case cmd := <-CommandQueue:
			switch cmd.Note {
			case NoteCut:
				for i := range voicePool {
					if (voicePool[i].TrackID == cmd.TrackID || cmd.TrackID == 0) && (voicePool[i].ColumnID == cmd.ColumnID || cmd.ColumnID == 0) && voicePool[i].IsPlaying() {
						Logf("Cutting %d.\n", i)
						voicePool[i].Cut()
						// break // Column coordinates are unique, so we can stop searching immediately
					}
				}

			case NoteOff:
				for i := range voicePool {
					if voicePool[i].TrackID == cmd.TrackID && voicePool[i].ColumnID == cmd.ColumnID && voicePool[i].IsPlaying() {
						voicePool[i].Stop()
						break // Column coordinates are unique, so we can stop searching immediately
					}
				}
			default:
				Logf("Queing %d.\n", cmd.Note)
				targetVoice := FindFreeVoice(cmd.ColumnID, cmd.TrackID)
				targetVoice.TrackID = cmd.TrackID
				targetVoice.ColumnID = cmd.ColumnID
				targetVoice.Note = cmd.Note
				targetVoice.Waveform = cmd.Waveform
				targetVoice.CustomWaveBuffer = CurrentProject.Tracks[cmd.TrackID].CustomWaveSamples
				targetVoice.SampleLoopStart = 34113.0
				targetVoice.SampleLoopEnd = 34113.0 + 5396.0 // 6000.0 + float64(slider) // 6671.0
				/*
					targetVoice.SampleLoopStart = CurrentProject.Tracks[cmd.TrackID].SampleLoopStart
					targetVoice.SampleLoopEnd = CurrentProject.Tracks[cmd.TrackID].SampleLoopEnd
				*/
				targetVoice.Envelope.SustainTicks = cmd.SustainTicks
				targetVoice.Play()
			}
		default:
			break CommandLoop
		}
	}
}

func MasterUpdate(ctx ev.EventContext) bool {
	if !rl.IsAudioStreamProcessed(masterStream) {
		return false
	}

	MasterQueueCommands()

	currentHeadroomCount := 0
	for _, v := range voicePool {
		if v.Envelope.State != EnvIdle {
			currentHeadroomCount++
		}
	}
	if currentHeadroomCount < 1 {
		return false
	}

	headroomScale := 1.0 / float32(currentHeadroomCount+1)

	for i := range masterBuffer {
		masterBuffer[i] = 0.0
	}
	var maxAbs float32
	for _, v := range voicePool {
		if v.Envelope.State == EnvIdle {
			continue
		}
		maxAbs = max(maxAbs, v.UpdateVoice(masterBuffer[:], headroomScale))
	}
	saturator(masterBuffer[:], maxAbs)
	rl.UpdateAudioStream(masterStream, masterBuffer[:])
	return true
}

func (v *Voice) TickEnvelope() float32 {
	switch v.Envelope.State {
	case EnvIdle:
		v.Envelope.Value = 0.0
		v.Active = false // If the envelope is dead, the voice is dead

	case EnvAttack:
		// Ramp up to maximum (1.0)
		v.Envelope.Value += v.Envelope.AttackStep
		if v.Envelope.Value >= 1.0 {
			v.Envelope.Value = 1.0
			v.Envelope.State = EnvDecay // Transition automatically
		}

	case EnvDecay:
		// Ramp down to the Sustain Level
		v.Envelope.Value -= v.Envelope.DecayStep
		if v.Envelope.Value <= v.Envelope.SustainLevel {
			v.Envelope.Value = v.Envelope.SustainLevel
			v.Envelope.State = EnvSustain // Hold here until Stop() is called
		}

	case EnvSustain:
		v.Envelope.Value = v.Envelope.SustainLevel
		if v.Envelope.SustainTicks > 0 {
			v.Envelope.SustainTicks--
			if v.Envelope.SustainTicks == 0 {
				Logf("Releasing ...\n")
				v.Envelope.State = EnvRelease
			}
		}

	case EnvRelease:
		// Fade out completely to 0.0
		v.Envelope.Value -= v.Envelope.ReleaseStep
		if v.Envelope.Value <= 0.0 {
			v.Envelope.Value = 0.0
			v.Envelope.State = EnvIdle // Voice is now officially spent
		}
	}

	return v.Envelope.Value
}

func NewSaturator() func([]float32, float32) {
	// This variable is captured and persists across callbacks
	var lastGain float32 = 1.0

	return func(masterBuffer []float32, blockPeak float32) {
		nextGain := float32(1.0)
		if blockPeak > 1.0 {
			nextGain = 1.0 / blockPeak
		}

		// Calculate the smooth slope over the 4096 samples
		startGain := lastGain
		gainDelta := (nextGain - startGain) / float32(len(masterBuffer))

		for i := range masterBuffer {
			masterBuffer[i] *= startGain + (gainDelta * float32(i))
		}

		// Update the captured state for the next Raylib callback
		lastGain = nextGain
	}
}
