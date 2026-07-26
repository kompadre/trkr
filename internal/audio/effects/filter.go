package effects

import (
	"math"
)

// FilterType defines the type of audio filter (LPF or HPF).
type FilterType int

const (
	FilterTypeNone FilterType = iota
	FilterTypeLPF
	FilterTypeHPF
)

func (t FilterType) String() string {
	switch t {
	case FilterTypeLPF:
		return "LPF"
	case FilterTypeHPF:
		return "HPF"
	default:
		return "Unknown"
	}
}

// Filter implements a 2-pole Biquad Low-Pass / High-Pass audio filter (Direct Form II Transposed).
type Filter struct {
	Type       FilterType
	Cutoff     float64 // Cutoff frequency in Hz
	Resonance  float64 // Q factor (default: 0.707 for Butterworth)
	SampleRate float64 // Sample rate in Hz

	// Normalized coefficients
	b0 float64
	b1 float64
	b2 float64
	a1 float64
	a2 float64

	// Internal state variables (Direct Form II Transposed)
	s1 float64
	s2 float64
}

// NewFilter creates and initializes a new Biquad filter.
func NewFilter(filterType FilterType, cutoff float64, resonance float64, sampleRate float64) *Filter {
	if resonance <= 0 {
		resonance = 0.7071067811865475 // Default 1/sqrt(2) Butterworth response
	}
	if sampleRate <= 0 {
		sampleRate = 48000.0
	}
	f := &Filter{
		Type:       filterType,
		Cutoff:     cutoff,
		Resonance:  resonance,
		SampleRate: sampleRate,
	}
	f.RecalculateCoefficients()
	return f
}

// RecalculateCoefficients updates filter coefficients using RBJ Audio EQ Cookbook formulas.
func (f *Filter) RecalculateCoefficients() {
	if f.SampleRate <= 0 {
		return
	}

	// Clamp cutoff frequency between 10 Hz and Nyquist limit (0.49 * SampleRate)
	nyquist := f.SampleRate * 0.49
	cutoff := f.Cutoff
	if cutoff < 10.0 {
		cutoff = 10.0
	} else if cutoff > nyquist {
		cutoff = nyquist
	}

	q := f.Resonance
	if q <= 0.001 {
		q = 0.001
	}

	omega := 2.0 * math.Pi * cutoff / f.SampleRate
	cosW := math.Cos(omega)
	sinW := math.Sin(omega)
	alpha := sinW / (2.0 * q)

	var b0, b1, b2, a0, a1, a2 float64

	switch f.Type {
	case FilterTypeLPF:
		b0 = (1.0 - cosW) / 2.0
		b1 = 1.0 - cosW
		b2 = (1.0 - cosW) / 2.0
		a0 = 1.0 + alpha
		a1 = -2.0 * cosW
		a2 = 1.0 - alpha
	case FilterTypeHPF:
		b0 = (1.0 + cosW) / 2.0
		b1 = -(1.0 + cosW)
		b2 = (1.0 + cosW) / 2.0
		a0 = 1.0 + alpha
		a1 = -2.0 * cosW
		a2 = 1.0 - alpha
	}

	// Normalize coefficients by a0
	f.b0 = b0 / a0
	f.b1 = b1 / a0
	f.b2 = b2 / a0
	f.a1 = a1 / a0
	f.a2 = a2 / a0
}

// SetType updates the filter type (LPF/HPF) and recalculates coefficients.
func (f *Filter) SetType(filterType FilterType) {
	if f.Type != filterType {
		f.Type = filterType
		f.RecalculateCoefficients()
	}
}

// SetCutoff updates the cutoff frequency (in Hz) and recalculates coefficients.
func (f *Filter) SetCutoff(cutoff float64) {
	if f.Cutoff != cutoff {
		f.Cutoff = cutoff
		f.RecalculateCoefficients()
	}
}

// SetResonance updates the Q factor/resonance and recalculates coefficients.
func (f *Filter) SetResonance(resonance float64) {
	if f.Resonance != resonance {
		f.Resonance = resonance
		f.RecalculateCoefficients()
	}
}

// SetSampleRate updates the sample rate and recalculates coefficients.
func (f *Filter) SetSampleRate(sampleRate float64) {
	if f.SampleRate != sampleRate {
		f.SampleRate = sampleRate
		f.RecalculateCoefficients()
	}
}

// Reset clears the filter's internal state memory.
func (f *Filter) Reset() {
	f.s1 = 0.0
	f.s2 = 0.0
}

// ProcessSample filters a single sample using Direct Form II Transposed structure.
func (f *Filter) ProcessSample(sample float32) float32 {
	x := float64(sample)
	y := f.b0*x + f.s1

	f.s1 = f.b1*x - f.a1*y + f.s2
	f.s2 = f.b2*x - f.a2*y

	// --- DENORMAL / LIMIT CYCLE FLUSH ---
	// Math.Abs threshold check (e.g. 1e-15 for float64 precision)
	if f.s1 > -1e-15 && f.s1 < 1e-15 {
		f.s1 = 0
	}
	if f.s2 > -1e-15 && f.s2 < 1e-15 {
		f.s2 = 0
	}

	return float32(y)
}

// Process filters a slice of audio samples in place.
func (f *Filter) Process(samples []float32) {
	for i, sample := range samples {
		samples[i] = f.ProcessSample(sample)
	}
}
