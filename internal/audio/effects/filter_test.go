package effects

import (
	"math"
	"testing"
)

func generateSineWave(freq float64, sampleRate float64, numSamples int) []float32 {
	buf := make([]float32, numSamples)
	for i := 0; i < numSamples; i++ {
		t := float64(i) / sampleRate
		buf[i] = float32(math.Sin(2.0 * math.Pi * freq * t))
	}
	return buf
}

func calcRMS(samples []float32) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sumSq float64
	for _, s := range samples {
		sumSq += float64(s * s)
	}
	return math.Sqrt(sumSq / float64(len(samples)))
}

func TestLPF(t *testing.T) {
	sampleRate := 48000.0
	numSamples := 4800 // 100ms

	lowFreqSignal := generateSineWave(100.0, sampleRate, numSamples)
	highFreqSignal := generateSineWave(10000.0, sampleRate, numSamples)

	lpf := NewFilter(FilterTypeLPF, 500.0, 0.707, sampleRate)

	lowFiltered := make([]float32, len(lowFreqSignal))
	copy(lowFiltered, lowFreqSignal)
	lpf.Reset()
	lpf.Process(lowFiltered)

	highFiltered := make([]float32, len(highFreqSignal))
	copy(highFiltered, highFreqSignal)
	lpf.Reset()
	lpf.Process(highFiltered)

	// Ignore initial transient (first 500 samples)
	lowRMS := calcRMS(lowFiltered[500:])
	highRMS := calcRMS(highFiltered[500:])

	t.Logf("LPF (Cutoff 500Hz) - Low Freq (100Hz) RMS: %f, High Freq (10kHz) RMS: %f", lowRMS, highRMS)

	if lowRMS < 0.6 {
		t.Errorf("Expected low frequency signal to pass with high RMS, got %f", lowRMS)
	}
	if highRMS > 0.05 {
		t.Errorf("Expected high frequency signal to be heavily attenuated, got %f", highRMS)
	}
}

func TestHPF(t *testing.T) {
	sampleRate := 48000.0
	numSamples := 4800 // 100ms

	lowFreqSignal := generateSineWave(100.0, sampleRate, numSamples)
	highFreqSignal := generateSineWave(10000.0, sampleRate, numSamples)

	hpf := NewFilter(FilterTypeHPF, 5000.0, 0.707, sampleRate)

	lowFiltered := make([]float32, len(lowFreqSignal))
	copy(lowFiltered, lowFreqSignal)
	hpf.Reset()
	hpf.Process(lowFiltered)

	highFiltered := make([]float32, len(highFreqSignal))
	copy(highFiltered, highFreqSignal)
	hpf.Reset()
	hpf.Process(highFiltered)

	// Ignore initial transient
	lowRMS := calcRMS(lowFiltered[500:])
	highRMS := calcRMS(highFiltered[500:])

	t.Logf("HPF (Cutoff 5000Hz) - Low Freq (100Hz) RMS: %f, High Freq (10kHz) RMS: %f", lowRMS, highRMS)

	if highRMS < 0.6 {
		t.Errorf("Expected high frequency signal to pass with high RMS, got %f", highRMS)
	}
	if lowRMS > 0.05 {
		t.Errorf("Expected low frequency signal to be heavily attenuated, got %f", lowRMS)
	}
}

func TestFilterReset(t *testing.T) {
	f := NewFilter(FilterTypeLPF, 1000.0, 0.707, 48000.0)
	f.ProcessSample(1.0)
	if f.s1 == 0 && f.s2 == 0 {
		t.Errorf("Expected filter state to be non-zero after processing sample")
	}
	f.Reset()
	if f.s1 != 0 || f.s2 != 0 {
		t.Errorf("Expected filter state to be 0 after Reset(), got s1=%f, s2=%f", f.s1, f.s2)
	}
}
