package events

import rl "github.com/gen2brain/raylib-go/raylib"

type EventKind int

const (
	EventKindTick EventKind = iota
	EventKindInput
	EventKindGuiDraw
	EventKindShutdown
)

type EventContext struct {
	EventData    []byte
	EventPayload interface{}
}

type EventCallback func(EventContext) bool

type Event struct {
	EventKind
	RegistredCallbacks []EventCallback
}

var EventMap = map[EventKind]*Event{}

type InputKind int32

const (
	InputKindPressedLeft     InputKind = rl.KeyLeft
	InputKindPressedUp       InputKind = rl.KeyUp
	InputKindPressedRight    InputKind = rl.KeyRight
	InputKindPressedDown     InputKind = rl.KeyDown
	InputKindPressedA        InputKind = rl.KeyA
	InputKindPressedB        InputKind = rl.KeyB
	InputKindPressedL        InputKind = rl.KeyL
	InputKindPressedM        InputKind = rl.KeyM
	InputKindPressedR        InputKind = rl.KeyR
	InputKindPressedSpace    InputKind = rl.KeySpace
	InputKindPressedPageDown InputKind = rl.KeyPageDown
	InputKindPressedPageUp   InputKind = rl.KeyPageUp
	InputKindDir             InputKind = 999
)

var InputInputs = []InputKind{InputKindPressedLeft, InputKindPressedUp, InputKindPressedRight, InputKindPressedDown,
	InputKindPressedA, InputKindPressedB, InputKindPressedL, InputKindPressedR, InputKindPressedSpace, InputKindDir}

type InputSnapshot map[InputKind]bool

func CalculateInputSnapshot() InputSnapshot {
	snapshot := InputSnapshot{}
	for _, key := range InputInputs {
		if rl.IsKeyDown(int32(key)) {
			snapshot[key] = true
			if key == InputKindPressedLeft || key == InputKindPressedUp || key == InputKindPressedRight || key == InputKindPressedDown || key == InputKindPressedPageDown || key == InputKindPressedPageUp {
				snapshot[InputKindDir] = true
			}
		}
	}
	return snapshot
}

func RegisterCallback(eventKind EventKind, callback EventCallback) {
	event, ok := EventMap[eventKind]
	if !ok {
		EventMap[eventKind] = &Event{
			EventKind:          eventKind,
			RegistredCallbacks: []EventCallback{callback},
		}
	} else {
		event.RegistredCallbacks = append(event.RegistredCallbacks, callback)
	}
}

func ClearCallbacks(eventKind EventKind) {
	_, ok := EventMap[eventKind]
	if ok {
		EventMap[eventKind].RegistredCallbacks = []EventCallback{}
	}
}

func Trigger(kind EventKind, ctx EventContext) bool {
	event, ok := EventMap[kind]
	if !ok {
		return false
	}
	result := false
	for _, callback := range event.RegistredCallbacks {
		result = result || callback(ctx)
	}
	return result
}
