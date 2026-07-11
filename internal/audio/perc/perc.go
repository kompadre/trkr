package perc

import (
	"math"
	"math/rand/v2"
)

type PercussionType uint8

const (
	PercussionVoiceTypeKick PercussionType = iota
	PercussionVoiceTypeSnare
	PercussionVoiceTypeHihat
)

type Percussion struct {
	CurrentFreqC float64
	Freq         float64
	PitchEnv     float64
	PitchDecay   float64
	AmpEnv       float64
	AmpDecay     float64
	NoiseEnv     float64
	NoiseDecay   float64
	ModAmt       float64
	ModRatio     float64
	PhaseC       float64
	PhaseM       float64
	LastNoise    float64
	NoiseMix     float64
}

func (v *Percussion) ProcessSample(idx int) float32 {
	const TwoPi = 2.0 * math.Pi

	// 1. Calculate the raw target carrier frequency
	targetFreqC := v.Freq + (50.0 * v.PitchEnv * v.PitchEnv)

	// 2. Smooth the persistent struct variable toward the target
	if v.CurrentFreqC == 0 {
		v.CurrentFreqC = targetFreqC // Initialize on first sample so it doesn't sweep from 0Hz
	}
	v.CurrentFreqC = (v.CurrentFreqC * 0.995) + (targetFreqC * 0.005)

	// Modulator frequency can just be local now, since it derives from the smoothed carrier
	currentFreqM := v.CurrentFreqC * v.ModRatio

	// 3. Step the FM phases
	v.PhaseC += (TwoPi * v.CurrentFreqC) / 48000.0
	if v.PhaseC > TwoPi {
		v.PhaseC -= TwoPi
	}

	v.PhaseM += (TwoPi * currentFreqM) / 48000.0
	if v.PhaseM > TwoPi {
		v.PhaseM -= TwoPi
	}

	// 4. Render the Tonal FM Component
	modulatorSignal := math.Sin(v.PhaseM) * v.ModAmt * v.AmpEnv
	toneOutput := math.Sin(v.PhaseC+modulatorSignal) * v.AmpEnv

	// 5. Render the Noise Component
	rawNoise := (rand.Float64() * 2.0) - 1.0

	var noiseOutput float64
	if v.NoiseMix > 0.9 {
		// HI-HAT MODE: Raw noise gives a solid acoustic cymbal sizzle
		noiseOutput = rawNoise * v.NoiseEnv
	} else {
		// SNARE/KICK MODE: High-pass filtered bite
		filteredNoise := rawNoise - v.LastNoise
		v.LastNoise = rawNoise
		noiseOutput = filteredNoise * v.NoiseEnv
	}

	// 6. Tick down all envelopes
	v.AmpEnv *= v.AmpDecay
	v.NoiseEnv *= v.NoiseDecay

	v.PitchEnv *= v.PitchDecay
	if v.PitchEnv < 0.0001 {
		v.PitchEnv = 0.0
	}

	// 7. Clean mathematical balance using your NoiseMix optimization
	toneWeight := 1.0 - v.NoiseMix
	noiseWeight := v.NoiseMix

	return float32((toneOutput * toneWeight) + (noiseOutput * noiseWeight))
}

func (v *Percussion) GetSamples(buffer []float32) int {
	idx := 0
	for idx < len(buffer) /* && v.AmpEnv >= 0.001 */ {
		buffer[idx] = v.ProcessSample(idx)
		idx++
	}
	return idx
}
