package events

import (
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type EventKind int

const (
	EventKindTick EventKind = iota
	EventKindInput
	EventKindUpdate
	EventKindPostUpdate
	EventKindShutdown
)

const EventGamepadDeadZone = 0.1

type EventContext struct {
	EventData    []byte
	EventPayload interface{}
}

type EventCallback func(EventContext) bool

type Event struct {
	EventKind
	RegistredCallbacks map[uint16]EventCallback
	CallbackKeys       []uint16
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
	InputKindPressedEnter    InputKind = rl.KeyEnter
	InputKindGamepadLeft     InputKind = 998
	InputKindDir             InputKind = 999
)

var InputInputs = []InputKind{InputKindPressedLeft, InputKindPressedUp, InputKindPressedRight, InputKindPressedDown,
	InputKindPressedA, InputKindPressedB, InputKindPressedL, InputKindPressedR, InputKindPressedSpace, InputKindDir, InputKindPressedPageDown,
	InputKindPressedPageUp, InputKindPressedEnter}

type InputSnapshot map[InputKind]bool

func CalculateInputSnapshot() InputSnapshot {
	snapshot := InputSnapshot{}
	gamepadAxisLeftX := rl.GetGamepadAxisMovement(0, rl.GamepadAxisLeftX)
	gamepadAxisLeftY := rl.GetGamepadAxisMovement(0, rl.GamepadAxisLeftY)

	if gamepadAxisLeftX < EventGamepadDeadZone*-1 || gamepadAxisLeftX > EventGamepadDeadZone ||
		gamepadAxisLeftY < EventGamepadDeadZone*-1 || gamepadAxisLeftY > EventGamepadDeadZone {
		if gamepadAxisLeftX < EventGamepadDeadZone*-1 {
			snapshot[InputKindPressedLeft] = true
		} else if gamepadAxisLeftX > EventGamepadDeadZone {
			snapshot[InputKindPressedRight] = true
		} else if gamepadAxisLeftY < EventGamepadDeadZone*-1 {
			snapshot[InputKindPressedUp] = true
		} else if gamepadAxisLeftY > EventGamepadDeadZone*1 {
			snapshot[InputKindPressedDown] = true
		}
		snapshot[InputKindDir] = true
	}

	for _, key := range InputInputs {
		if rl.IsKeyDown(int32(key)) {
			snapshot[key] = true
			if key == InputKindPressedLeft || key == InputKindPressedUp || key == InputKindPressedRight ||
				key == InputKindPressedDown || key == InputKindPressedPageDown || key == InputKindPressedPageUp {
				snapshot[InputKindDir] = true
			}
		}
	}
	return snapshot
}

func RegisterCallback(eventKind EventKind, callback EventCallback, ID uint16) {
	eventQueue, ok := EventMap[eventKind]
	if !ok {
		eventQueue := &Event{
			EventKind:          eventKind,
			RegistredCallbacks: make(map[uint16]EventCallback),
		}

		eventQueue.RegistredCallbacks[ID] = callback
		eventQueue.CallbackKeys = []uint16{ID}
		EventMap[eventKind] = eventQueue

	} else {
		eventQueue.RegistredCallbacks[ID] = callback
		eventQueue.CallbackKeys = append(eventQueue.CallbackKeys, ID)
		slices.Sort(eventQueue.CallbackKeys)
	}
}

func ClearCallbacks(eventKind EventKind) {
	_, ok := EventMap[eventKind]
	if ok {
		for id := range EventMap[eventKind].RegistredCallbacks {
			delete(EventMap[eventKind].RegistredCallbacks, id)
		}
	}
}

func RemoveCallback(eventKind EventKind, ID uint16) {
	_, ok := EventMap[eventKind].RegistredCallbacks[ID]
	if ok {
		delete(EventMap[eventKind].RegistredCallbacks, ID)
		EventMap[eventKind].CallbackKeys = slices.DeleteFunc(EventMap[eventKind].CallbackKeys, func(k uint16) bool {
			return k == ID
		})
	}
}

func Trigger(kind EventKind, ctx EventContext) bool {
	ev, ok := EventMap[kind]
	if !ok {
		return false
	}
	result := false
	for _, k := range ev.CallbackKeys {
		result = ev.RegistredCallbacks[k](ctx)
		if result && kind == EventKindPostUpdate {
			RemoveCallback(kind, k)
		}
		if result {
			break
		}
	}
	return result
}
