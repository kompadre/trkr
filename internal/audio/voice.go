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
	VoiceBufferSize = 4048
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
)

type Envelope struct {
	State EnvelopeState
	Value float32 // The current instantaneous volume multiplier (0.0 to 1.0)

	// Step sizes added/subtracted per sample frame
	AttackStep   float32
	DecayStep    float32
	SustainLevel float32 // Target volume level during sustain (0.0 to 1.0)
	SustainTicks uint8
	ReleaseStep  float32
}

type Voice struct {
	ID       uint8
	TrackID  uint8
	ColumnID uint8
	Active   bool
	Envelope
	Note
	Volume   float32
	Waveform // Embedded type. Public field named 'Waveform'

	_frequency float64
	_phase     float64
	_phaseStep float64
}

type VoiceCommand struct {
	TrackID  uint8
	ColumnID uint8
	Note     Note
	Waveform Waveform
}

var voicePool []*Voice
var isMasterInitialized bool
var masterBuffer [VoiceBufferSize]float32
var masterStream rl.AudioStream
var activeVoices uint8
var freeVoices chan *Voice
var CommandQueue = make(chan VoiceCommand, 64)

func InitVoiceMasterStream() {
	masterStream = rl.LoadAudioStream(VoiceSampleRate, 32, 1)
	ev.RegisterCallback(ev.EventKindAudioUpdate, MasterUpdate, 1)
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
			return voicePool[i]
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
	v._frequency = frequenciesTable[v.Note]
	v._phaseStep = v._frequency / VoiceSampleRate
	v._phase = rand.Float64()

	// Envelope settings
	v.Envelope.AttackStep = 1.0 / (0.15 * float32(VoiceSampleRate))
	v.Envelope.DecayStep = 1.0 / (0.2 * float32(VoiceSampleRate))
	v.Envelope.SustainLevel = 0.6
	v.Envelope.SustainTicks = 0
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

func (v *Voice) IsPlaying() bool {
	return v.Envelope.State != EnvIdle && v.Envelope.State != EnvRelease
}

func (v *Voice) UpdateVoice(writeBuffer []float32, headroomScale float32) {
	volScale := v.Volume * headroomScale

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
		}

		writeBuffer[i] += sample * combinedVolume

		v._phase += v._phaseStep
		if v._phase >= 1.0 {
			v._phase -= 1.0
		}
	}
}

func MasterQueueCommands() {
	// Label the outer loop so we can break out of it from inside the switch
CommandLoop:
	for {
		select {
		case cmd := <-CommandQueue:
			switch cmd.Note {
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

	for _, v := range voicePool {
		if v.Envelope.State == EnvIdle {
			continue
		}
		v.UpdateVoice(masterBuffer[:], headroomScale)
	}
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
		// Do nothing! Hold the sustain level indefinitely while key is pressed
		v.Envelope.Value = v.Envelope.SustainLevel
		if v.Envelope.SustainTicks > 0 {
			v.Envelope.SustainTicks--
			if v.Envelope.SustainTicks <= 0 {
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
