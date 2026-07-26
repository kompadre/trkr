package effects

import (
	"math"
	"testing"
)

func TestDelayEcho(t *testing.T) {
	sampleRate := 48000.0
	delayTime := 0.1 // 100ms = 4800 samples
	delaySamples := int(delayTime * sampleRate)
	feedback := 0.5
	mix := 0.5

	delay := NewDelay(delayTime, feedback, mix, sampleRate)

	numSamples := delaySamples*3 + 100
	samples := make([]float32, numSamples)
	samples[0] = 1.0 // Impulse at sample 0

	delay.Process(samples)

	// Sample 0: dry = 1.0 * (1-0.5) = 0.5, wet = 0.0 * 0.5 = 0.0 -> total = 0.5
	if math.Abs(float64(samples[0])-0.5) > 1e-4 {
		t.Errorf("Expected sample 0 output ~0.5, got %f", samples[0])
	}

	// Samples 1 to delaySamples-1 should be 0
	for i := 1; i < delaySamples; i++ {
		if math.Abs(float64(samples[i])) > 1e-4 {
			t.Errorf("Expected sample %d to be 0, got %f", i, samples[i])
		}
	}

	// First echo at sample delaySamples: wet = 1.0 * 0.5 = 0.5, dry = 0.0 -> total = 0.5
	if math.Abs(float64(samples[delaySamples])-0.5) > 1e-4 {
		t.Errorf("Expected first echo at sample %d ~0.5, got %f", delaySamples, samples[delaySamples])
	}

	// Second echo at sample 2*delaySamples: wet = 1.0 * feedback * mix = 0.25 -> total = 0.25
	secondEcho := delaySamples * 2
	if math.Abs(float64(samples[secondEcho])-0.25) > 1e-4 {
		t.Errorf("Expected second echo at sample %d ~0.25, got %f", secondEcho, samples[secondEcho])
	}
}

func TestDelayDryMix(t *testing.T) {
	sampleRate := 48000.0
	delay := NewDelay(0.1, 0.5, 0.0, sampleRate) // Mix = 0.0 (100% dry)

	samples := []float32{0.5, -0.25, 0.8, -0.1}
	original := make([]float32, len(samples))
	copy(original, samples)

	delay.Process(samples)

	for i, s := range samples {
		if math.Abs(float64(s-original[i])) > 1e-5 {
			t.Errorf("Expected 100%% dry sample %d to be %f, got %f", i, original[i], s)
		}
	}
}

func TestDelayReset(t *testing.T) {
	sampleRate := 48000.0
	delayTime := 0.01 // 10ms = 480 samples
	delaySamples := int(delayTime * sampleRate)

	delay := NewDelay(delayTime, 0.5, 0.5, sampleRate)

	// Feed impulse
	delay.ProcessSample(1.0)

	delay.Reset()

	// Process remaining silence after reset
	samples := make([]float32, delaySamples+50)
	delay.Process(samples)

	for i, s := range samples {
		if s != 0.0 {
			t.Errorf("Expected output sample %d to be 0.0 after Reset(), got %f", i, s)
		}
	}
}

func TestDelaySetters(t *testing.T) {
	delay := NewDelay(0.1, 0.3, 0.5, 44100.0)

	delay.SetDelayTime(0.2)
	if delay.DelayTime != 0.2 {
		t.Errorf("Expected DelayTime 0.2, got %f", delay.DelayTime)
	}

	delay.SetFeedback(1.5) // Clamped to 0.99
	if delay.Feedback != 0.99 {
		t.Errorf("Expected Feedback clamped to 0.99, got %f", delay.Feedback)
	}

	delay.SetMix(-0.5) // Clamped to 0.0
	if delay.Mix != 0.0 {
		t.Errorf("Expected Mix clamped to 0.0, got %f", delay.Mix)
	}

	delay.SetSampleRate(48000.0)
	if delay.SampleRate != 48000.0 {
		t.Errorf("Expected SampleRate 48000.0, got %f", delay.SampleRate)
	}
}
