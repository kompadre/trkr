package widget

import (
	"testing"
	"trkr"
	ev "trkr/internal/events"
)

func TestKnobValueClampingAndStep(t *testing.T) {
	var actionValue float64
	actionCalled := false
	col := [4]int{6, 6, 6, 6}
	knob := NewKnob("Volume", col, 0, 100, 50, func(val float64) bool {
		actionCalled = true
		actionValue = val
		return true
	}, nil)

	if knob.Value != 50 {
		t.Errorf("expected initial value 50, got %f", knob.Value)
	}

	// Test increment (default step is 1.0 here because 100/100)
	snap := ev.InputSnapshot{}
	snap.HoldTimers[ev.InputKindUp] = 1
	knob.HandleInput(&snap, nil)
	if knob.Value != 51 {
		t.Errorf("expected value 51 after increment, got %f", knob.Value)
	}

	// Test accelerated decrement
	snap = ev.InputSnapshot{}
	snap.HoldTimers[ev.InputKindDown] = 15
	knob.HandleInput(&snap, nil)
	if knob.Value != 46 { // 51 - (1.0 * 5) = 46
		t.Errorf("expected value 46 after accelerated decrement, got %f", knob.Value)
	}

	// Test lower bound clamp
	snap = ev.InputSnapshot{}
	snap.HoldTimers[ev.InputKindLeft] = 20
	for range 60 {
		knob.HandleInput(&snap, nil)
	}
	if knob.Value != 0 {
		t.Errorf("expected value clamped to minVal 0, got %f", knob.Value)
	}

	// Test upper bound clamp
	snap = ev.InputSnapshot{}
	snap.HoldTimers[ev.InputKindRight] = 20
	for range 120 {
		knob.HandleInput(&snap, nil)
	}
	if knob.Value != 100 {
		t.Errorf("expected value clamped to maxVal 100, got %f", knob.Value)
	}

	// Test Enter triggers action
	snap = ev.InputSnapshot{}
	snap.HoldTimers[ev.InputKindEnter] = 1
	snap.Pressed = 1 << ev.InputKindEnter
	knob.HandleInput(&snap, nil)
	if !actionCalled || actionValue != 100 {
		t.Errorf("expected action to be called with value 100, got called=%v val=%f", actionCalled, actionValue)
	}
}

func TestKnobOnChangeCallback(t *testing.T) {
	var changedVal float64 = -1
	knob := NewKnob("Cutoff", [4]int{12, 12, 12, 12}, 0, 100, 50, nil, nil)
	knob.OnChange = func(val float64) {
		changedVal = val
	}

	snap := ev.InputSnapshot{}
	snap.HoldTimers[ev.InputKindRight] = 1
	knob.HandleInput(&snap, nil)

	if changedVal != 51 {
		t.Errorf("expected OnChange callback with 51, got %f", changedVal)
	}
}

func TestKnobInitialClamp(t *testing.T) {
	knobUnder := NewKnob("Pan", [4]int{12, 12, 12, 12}, -50, 50, -100, nil, nil)
	if knobUnder.Value != -50 {
		t.Errorf("expected initial value clamped to min -50, got %f", knobUnder.Value)
	}

	knobOver := NewKnob("Pan", [4]int{12, 12, 12, 12}, -50, 50, 100, nil, nil)
	if knobOver.Value != 50 {
		t.Errorf("expected initial value clamped to max 50, got %f", knobOver.Value)
	}
}

// Quiet compiler check for trkr import
var _ = trkr.BeatsPerMinute
