package audio

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"math"
	"math/rand"
	. "trkr"
	"trkr/internal/audio/effects"
	"trkr/internal/audio/miniaudio"
	ev "trkr/internal/events"
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
	Id           uint8
	TrackId      uint8
	ColumnId     uint8
	InstrumentId uint8
	Instrument   *Instrument
	Active       bool
	Envelope
	Note
	Volume float64
	Waveform
	Samples []float32

	_frequency float64 `json:"-"`
	_phase     float64 `json:"-"`
	_phaseStep float64 `json:"-"`
}

type VoiceCommand struct {
	TrackId      uint8
	ColumnId     uint8
	InstrumentId uint8
	Note         Note
	Waveform     Waveform
	SampleSourceType
	SustainTicks uint32
	Velocity     uint8
	Volume       float64
}

var voicePool []*Voice
var isMasterInitialized bool
var masterBuffer [VoiceBufferSize]float32
var masterBuffer16x1 [VoiceBufferSize]int16
var masterBuffer32x2 [VoiceBufferSize]float32
var masterStream rl.AudioStream
var activeVoices uint8
var freeVoices chan *Voice
var CommandQueue = make(chan VoiceCommand, 64)
var saturator func([]float32, float32)
var Filter *effects.Filter
var Delay *effects.Delay
var masterRunningDry = 0

func InitVoiceMasterStream() {
	ev.RegisterCallback(ev.EventKindAudioUpdate, MasterUpdate, 1)
	saturator = NewSaturator()
	// filter = NewLowPassFilter()
	voicePool = make([]*Voice, VoicePoolSize)
	for i := range VoicePoolSize {
		voicePool[i] = &Voice{Id: uint8(i + 1), Volume: 1.0, _phase: rand.Float64()}
	}
	isMasterInitialized = true

	// Warm tape-style delay settings
	Delay = effects.NewDelay(
		60.0/105.0, // 160ms delay time
		0.28,       // 28% feedback
		0.45,       // 15% wet mix
		48000.0,
	)
	Filter = effects.NewFilter(effects.FilterTypeLPF, 2500.0, 0, 48000)
	/*
		Delay = effects.NewDelay(0.25, 0.4, 0.0, 48000)
	*/

}

func FindFreeVoice(columnId uint8, trackId uint8, instrumentId uint8) *Voice {
	var oldestVoice, currentVoice *Voice
	var lowestEnvVal float32 = 1.0 // Track the quietest voice to steal safely

	// Pass 1: Look for a completely free voice
	for i := range voicePool {
		if voicePool[i].ColumnId == columnId && voicePool[i].TrackId == trackId {
			currentVoice = voicePool[i]
		} else if instrumentId == voicePool[i].InstrumentId && voicePool[i].Envelope.State == EnvIdle && voicePool[i].TrackId != 0xff {
			return voicePool[i] // Found an empty slot with the same instrument!
		} else if voicePool[i].Envelope.State == EnvIdle && voicePool[i].TrackId != 0xff {
			return voicePool[i] // Found an empty slot instantly!
		}

		// Keep track of candidates to steal just in case Pass 1 fails.
		// We prefer stealing a voice that is already fading out (EnvRelease)
		// and has the lowest amplitude to minimize audible pops.
		if voicePool[i].Envelope.Value < lowestEnvVal && voicePool[i].TrackId != 0xff {
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
	if v.InstrumentId == 0xff || (v.Instrument != nil && v.Instrument.SampleSourceType == SampleSourceTypeFmPickup) {
		v.Envelope.State = EnvSustain
		v.Envelope.Value = 1.0
		return
	} else if v.Instrument.SampleSourceType == SampleSourceTypePerc {
		v.Envelope.State = EnvSustain
		slot := int(v.Note) % len(v.Instrument.Percs)
		v.Instrument.Percs[slot].Trigger()
		return
	}

	// For a sample with root note C3 (MIDI 48)

	// Define C3 using your tracker's native index system (C0 = 1, C1 = 13, C2 = 25, C3 = 37)
	const NativeC3 = 37
	// Inside your note-on trigger logic:
	if v.Instrument.SampleSourceType == SampleSourceTypeWavefile {
		if v.Instrument.RootNote != 0 {
			c3RootMultiplier := pitchTable[v.Instrument.RootNote]
			v._phaseStep = pitchTable[v.Note%120] / c3RootMultiplier
		} else {
			v._phaseStep = 1.0
		}

		v._phase = 0.0 //512.0 //1024.0
		//	v._phaseStep = v._frequency / VoiceSampleRate
	} else {
		v._frequency = frequenciesTable[v.Note%120] //130.81278265 //frequenciesTable[v.Note]
		v._phaseStep = v._frequency / VoiceSampleRate
		v._phase = rand.Float64()
	}

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
	return v.Envelope.State != EnvIdle
}

func (v *Voice) UpdateVoice(writeBuffer []float32) {
	var volScale float32 = float32(v.Volume)
	var bufferPerc []float32
	var waveLenInt int
	var waveLen float64
	if v.Instrument.SampleSourceType == SampleSourceTypePerc {
		bufferPerc = make([]float32, len(writeBuffer))
		waveLenInt = v.Instrument.GetPercSamples(bufferPerc, int(v.Note))
		waveLen = float64(waveLenInt)
	} else {
		waveLenInt = len(v.Samples)
		waveLen = float64(waveLenInt)
	}

	for i := range writeBuffer {
		if v.TrackId != 0xff {
			envVolume := v.TickEnvelope()
			if v.Envelope.State == EnvIdle {
				break
			}
			// Only apply the software envelope to simple waveforms (Square, Saw, Cosine)
			// Perc and Wavefiles handle their own amplitude/attenuation.
			isSimpleWaveform := v.Instrument.SampleSourceType == SampleSourceTypeSquare ||
				v.Instrument.SampleSourceType == SampleSourceTypeSawtooth ||
				v.Instrument.SampleSourceType == SampleSourceTypeCosine

			if isSimpleWaveform {
				volScale = float32(v.Volume) * envVolume
			} else {
				volScale = float32(v.Volume)
			}
		}
		var sample float32 = 0.0
		switch v.Instrument.SampleSourceType {
		case SampleSourceTypeSquare:
			if v._phase >= 0.5 {
				sample = -1.0
			} else {
				sample = 1.0
			}
		case SampleSourceTypeSawtooth:
			sample = float32((v._phase * 2.0) - 1.0)
		case SampleSourceTypeCosine:
			sample = float32(math.Cos(v._phase * 2.0 * math.Pi))
		case SampleSourceTypePerc:
			if waveLenInt > i {
				sample = bufferPerc[i]
			} else {
				sample = 0.0
				v.Envelope.State = EnvIdle
			}
		case SampleSourceTypeWavefile:
			if !v.Instrument.SamplesLoaded {
				return
			}
			idxA := int(v._phase)
			if v.Instrument.LoopEnd == 0 {
				if waveLenInt > idxA {
					sample = v.Samples[idxA]
				} else {
					sample = 0.0
					v.Envelope.State = EnvIdle
				}
			} else {
				frac := float32(v._phase - float64(idxA))
				idxB := idxA + 1

				if float64(idxB) >= v.Instrument.LoopEnd {
					idxB = int(v.Instrument.LoopStart) + (idxB - int(v.Instrument.LoopEnd))
				} else if idxB >= int(waveLen) {
					idxB = int(waveLen) - 1
				}

				if idxA < int(waveLen) && idxA >= 0 {
					sample = v.Samples[idxA] + frac*(v.Samples[idxB]-v.Samples[idxA])
				} else {
					sample = 0.0
				}
			}
		}

		sample *= volScale
		writeBuffer[i] += sample
		v._phase += v._phaseStep

		if v.Instrument.SampleSourceType == SampleSourceTypeWavefile {
			if v.Instrument.LoopEnd != 0.0 {
				if v._phase >= v.Instrument.LoopEnd {
					overshoot := v._phase - v.Instrument.LoopEnd
					v._phase = v.Instrument.LoopStart + math.Mod(overshoot, v.Instrument.LoopEnd-v.Instrument.LoopStart)
				}
			} else {
				if v._phase >= waveLen {
					v.Envelope.State = EnvIdle
				}
			}
		}
	}
}

var slider uint8

func MasterQueueCommands() {
	// Label the outer loop so we can break out of it from inside the switch
CommandLoop:
	for {
		select {
		case cmd := <-CommandQueue:
			if cmd.SampleSourceType == SampleSourceTypeFm {
				if cmd.Note == NoteOff {
					StopSoundFm(cmd.ColumnId, cmd.TrackId)
				} else {
					PlaySoundFm(cmd.ColumnId, cmd.TrackId, cmd.Note, cmd.Velocity)
				}
				//break CommandLoop
			} else {
				switch cmd.Note {
				case NoteCut:
					for i := range voicePool {
						if (voicePool[i].TrackId == cmd.TrackId || cmd.TrackId == 0) && (voicePool[i].ColumnId == cmd.ColumnId || cmd.ColumnId == 0) && voicePool[i].IsPlaying() {
							Logf("Cutting %d.\n", i)
							voicePool[i].Cut()
							// break // Column coordinates are unique, so we can stop searching immediately
						}
					}

				case NoteOff:
					for i := range voicePool {
						if voicePool[i].TrackId == cmd.TrackId && voicePool[i].ColumnId == cmd.ColumnId && voicePool[i].IsPlaying() {
							voicePool[i].Stop()
							break // Column coordinates are unique, so we can stop searching immediately
						}
					}

				default:
					ins := &CurrentProject.Instruments[cmd.InstrumentId]
					if ins.SampleSourceType == SampleSourceTypeWavefile && !ins.SamplesLoaded {
						ins.LoadSamples()
						Logf("Filenames: %v, len(Samples[0]): %d.\n", ins.SampleSource, len(ins.Samples[0]))
						continue
					}
					targetVoice := FindFreeVoice(cmd.ColumnId, cmd.TrackId, cmd.InstrumentId)
					targetVoice.InstrumentId = ins.Id
					targetVoice.Instrument = ins
					if ins.SampleSourceType == SampleSourceTypeWavefile {
						idx := int(cmd.Note) % (len(targetVoice.Instrument.Samples))
						if !(len(targetVoice.Instrument.Samples[idx]) > 0) {
							continue
						}
						targetVoice.Samples = targetVoice.Instrument.Samples[idx][:]
					}
					targetVoice.TrackId = cmd.TrackId
					targetVoice.ColumnId = cmd.ColumnId
					targetVoice.Note = cmd.Note
					targetVoice.Envelope.SustainTicks = cmd.SustainTicks
					targetVoice.Volume = cmd.Volume
					targetVoice.Play()
				}
			}
		default:
			break CommandLoop
		}
	}
}

func MasterUpdate(ctx ev.EventContext) bool {
	if IsExporting {
		return true
	}
	for miniaudio.AvailableWriteSpace() >= VoiceBufferSize {
		MasterQueueCommands()
		Mix(masterBuffer[:])
		miniaudio.WriteChannels(masterBuffer[:], masterBuffer[:])
	}
	return true
}

var msfaScratchBuffer []int16

func Mix(out []float32) {
	if SynthInstance != nil {
		if len(msfaScratchBuffer) < len(out) {
			msfaScratchBuffer = make([]int16, len(out))
		}
		SynthInstance.GetSamples(msfaScratchBuffer[:len(out)])
	}

	for i := range out {
		out[i] = 0.0
	}

	hasContent := false
	for _, v := range voicePool {
		if !v.IsPlaying() || v.Instrument.SampleSourceType == SampleSourceTypeFm {
			continue
		}
		hasContent = true
		// Special case for FM Pickup: we already fetched the samples
		if v.Instrument.SampleSourceType == SampleSourceTypeFmPickup {
			for i := range out {
				out[i] += (float32(msfaScratchBuffer[i]) / 32768.0) * float32(v.Volume)
			}
			continue
		}
		v.UpdateVoice(out)
	}

	if !hasContent && SynthInstance == nil {
		return
	}

	// Calculate peak and apply saturator (limiter)
	var maxAbs float32
	for _, s := range out {
		a := s
		if a < 0 {
			a = -a
		}
		if a > maxAbs {
			maxAbs = a
		}
	}

	if saturator != nil {
		saturator(out, maxAbs)
	}

	if Delay != nil && Delay.Mix > 0 {
		//			Delay.Process(out)
	}
	if Filter.Type != effects.FilterTypeNone {
		//		Filter.Process(out)
	}
}

func ResetSaturator() {
	saturator = NewSaturator()
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

type LowPassFilter struct {
	v0 float32
	v1 float32
}

// Process takes the current raw audio sample and returns the filtered sample.
// 'cut' (0.0 to 1.0) and 'res' (0.0 to 1.0) are your real-time knob values.
func (f *LowPassFilter) Process(input float32, cut float32, res float32) float32 {
	feedback := res * 1.5 // Scale it so it screams but doesn't blow up your speakers
	f.v0 = (1.0-cut)*f.v0 + cut*(input-f.v1*feedback)
	f.v1 = (1.0-cut)*f.v1 + cut*f.v0
	return f.v1
}

func NewLowPassFilter() func(masterBuffer []float32) {
	f := LowPassFilter{}
	var res float32 = 0.0
	var cut float32 = 0.6

	samplesPerBeat := float32(48000 / (BeatsPerMinute / 60))
	step := 1.0 / samplesPerBeat

	direction := float32(-1.0)

	return func(masterBuffer []float32) {
		for i := range masterBuffer {
			masterBuffer[i] = f.Process(masterBuffer[i], cut, res)

			// Move the knobs based on current direction
			cut += step * direction

			// Map resonance so it climbs as cutoff drops (classic 303 squelch)
			// When cut is low (0.15), res is high (0.75). When cut is high (0.6), res is low (0.0).
			res = (0.6 - cut) * 1.66

			// Reversing logic instead of snapping
			if direction < 0 && cut <= 0.15 {
				direction = 1.0 // Turn around and go back UP
			} else if direction > 0 && cut >= 0.6 {
				direction = -1.0 // Turn around and go back DOWN
			}
		}
	}
}

func NewSaturator() func([]float32, float32) {
	// This variable is captured and persists across callbacks
	var currentGain float32 = 1.0

	return func(masterBuffer []float32, blockPeak float32) {
		targetGain := float32(0.8)
		if blockPeak > 1.0 {
			targetGain = 1.0 / blockPeak
		}

		// Calculate how many samples we'll process
		n := float32(len(masterBuffer))
		if n <= 0 {
			return
		}

		// Smoothing factor: how fast we reach the target gain.
		// We want a very fast attack (drop) and a slower release.
		for i := range masterBuffer {
			var alpha float32
			if targetGain < currentGain {
				alpha = 0.05 // Fast attack to stop clipping (~1ms)
			} else {
				alpha = 0.0005 // Slower release to avoid pumping (~100ms)
			}
			currentGain = currentGain*(1.0-alpha) + targetGain*alpha
			masterBuffer[i] *= currentGain
		}
	}
}

func CutAllVoices() {
	for _, v := range voicePool {
		if v.IsPlaying() {
			v.Cut()
		}
	}
}

func CutTrackVoices(trackId uint8) {
	for _, v := range voicePool {
		if v.TrackId == trackId && v.IsPlaying() {
			v.Cut()
		}
	}
}
