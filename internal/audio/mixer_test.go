package audio

import (
	"testing"
	. "trkr"
)

func TestMixerFoundations(t *testing.T) {
	// 1. Setup a voice with a simple waveform (Square)
	InitVoiceMasterStream()
	
	ins := Instrument{
		Id:               0,
		SampleSourceType: SampleSourceTypeSquare,
	}
	
	v := voicePool[0]
	v.Instrument = &ins
	v.InstrumentId = ins.Id
	v.Volume = 1.0
	v.Note = 60 // C-4
	v.Play()
	
	// Ensure envelope is at full for testing
	v.Envelope.State = EnvSustain
	v.Envelope.Value = 1.0

	// 2. Run Mix
	out := make([]float32, 256)
	Mix(out)

	// 3. Verify that we have non-zero samples
	hasSound := false
	for _, s := range out {
		if s != 0 {
			hasSound = true
			break
		}
	}

	if !hasSound {
		t.Error("Mix produced silence, expected audio from square wave voice")
	}
}

func TestSaturatorLimiting(t *testing.T) {
	saturator := NewSaturator()
	
	// Create a longer buffer with huge peaks to allow attack to settle
	out := make([]float32, 200)
	for i := range out {
		out[i] = 2.0
	}
	
	// Run saturator multiple times to simulate multiple blocks if needed, 
	// or just one long block.
	saturator(out, 2.0)
	
	// The end of the buffer should be limited
	lastSample := out[len(out)-1]
	if lastSample > 1.1 {
		t.Errorf("Final sample (%f) is still too high after 200 samples of saturation", lastSample)
	}
}

func TestVoiceLifecycle(t *testing.T) {
	v := &Voice{Volume: 1.0}
	v.Instrument = &Instrument{SampleSourceType: SampleSourceTypeSquare}
	
	// Test Play transitions to Attack
	v.Play()
	if v.Envelope.State != EnvAttack {
		t.Errorf("Expected EnvAttack, got %v", v.Envelope.State)
	}
	
	// Test Cut transitions to Idle
	v.Cut()
	if v.Envelope.State != EnvIdle {
		t.Errorf("Expected EnvIdle, got %v", v.Envelope.State)
	}
}
