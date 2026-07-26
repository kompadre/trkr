package events

import (
	"testing"
)

func TestInputSnapshot(t *testing.T) {
	var snap InputSnapshot
	snap.HoldTimers[InputKindUp] = 5
	snap.Released = 1 << InputKindDown

	if snap.Tick(InputKindUp) != 5 {
		t.Errorf("expected Tick(InputKindUp) == 5, got %d", snap.Tick(InputKindUp))
	}

	if !snap.Down(InputKindUp) {
		t.Errorf("expected Down(InputKindUp) to be true")
	}

	if snap.Down(InputKindLeft) {
		t.Errorf("expected Down(InputKindLeft) to be false")
	}

	if !snap.Up(InputKindDown) {
		t.Errorf("expected Up(InputKindDown) to be true")
	}

	if snap.Up(InputKindUp) {
		t.Errorf("expected Up(InputKindUp) to be false")
	}

	snap.ClearHoldTimers(InputKindUp)
	if snap.HoldTimers[InputKindUp] != 0 {
		t.Errorf("expected HoldTimers[InputKindUp] to be 0 after ClearHoldTimers, got %d", snap.HoldTimers[InputKindUp])
	}
}

func TestCallbacks(t *testing.T) {
	// Clean up EventMap state before testing
	EventMap = make(map[EventKind]*Event)

	var callOrder []uint16

	cb1 := func(ctx EventContext) bool {
		callOrder = append(callOrder, 1)
		return false
	}
	cb2 := func(ctx EventContext) bool {
		callOrder = append(callOrder, 2)
		return false
	}

	RegisterCallback(EventKindTick, cb2, 20)
	RegisterCallback(EventKindTick, cb1, 10)

	ctx := EventContext{EventData: []byte("test")}
	triggered := Trigger(EventKindTick, ctx)

	if triggered {
		t.Errorf("expected Trigger to return false when callbacks return false")
	}

	if len(callOrder) != 2 || callOrder[0] != 1 || callOrder[1] != 2 {
		t.Errorf("expected callbacks in order [1, 2], got %v", callOrder)
	}

	// Test callback returning true stopping execution
	callOrder = nil
	cbStop := func(ctx EventContext) bool {
		callOrder = append(callOrder, 1)
		return true
	}
	ClearCallbacks(EventKindTick)
	RegisterCallback(EventKindTick, cbStop, 10)
	RegisterCallback(EventKindTick, cb2, 20)

	triggered = Trigger(EventKindTick, ctx)
	if !triggered {
		t.Errorf("expected Trigger to return true when a callback returns true")
	}
	if len(callOrder) != 1 || callOrder[0] != 1 {
		t.Errorf("expected only first callback to execute, got %v", callOrder)
	}

	// Test RemoveCallback
	RemoveCallback(EventKindTick, 10)
	callOrder = nil
	Trigger(EventKindTick, ctx)
	if len(callOrder) != 1 || callOrder[0] != 2 {
		t.Errorf("expected only callback 2 to remain, got %v", callOrder)
	}

	// Test EventKindPostUpdate auto-removal on return true
	ClearCallbacks(EventKindPostUpdate)
	RegisterCallback(EventKindPostUpdate, cbStop, 100)
	Trigger(EventKindPostUpdate, ctx)

	if _, exists := EventMap[EventKindPostUpdate].RegistredCallbacks[100]; exists {
		t.Errorf("expected callback to be removed after returning true on PostUpdate event")
	}
}
