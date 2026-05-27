package events

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
)

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
	Descriptions       []string
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
	InputKindDir             InputKind = 999
)

var InputInputs = []InputKind{InputKindPressedLeft, InputKindPressedUp, InputKindPressedRight, InputKindPressedDown,
	InputKindPressedA, InputKindPressedB, InputKindPressedL, InputKindPressedR, InputKindPressedSpace, InputKindDir, InputKindPressedPageDown,
	InputKindPressedPageUp, InputKindPressedEnter}

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

func RegisterCallback(eventKind EventKind, callback EventCallback, callbackDescription string) {
	event, ok := EventMap[eventKind]
	if !ok {
		EventMap[eventKind] = &Event{
			EventKind:          eventKind,
			RegistredCallbacks: []EventCallback{callback},
			Descriptions:       []string{callbackDescription},
		}
	} else {
		event.RegistredCallbacks = append(event.RegistredCallbacks, callback)
		event.Descriptions = append(event.Descriptions, callbackDescription)
	}
}

func ClearCallbacks(eventKind EventKind) {
	_, ok := EventMap[eventKind]
	if ok {
		EventMap[eventKind].RegistredCallbacks = []EventCallback{}
	}
}

func PopCallback(eventKind EventKind) {
	ev, ok := EventMap[eventKind]
	if ok && len(ev.RegistredCallbacks) > 0 {
		ev.RegistredCallbacks[len(ev.RegistredCallbacks)-1] = nil
		ev.RegistredCallbacks = ev.RegistredCallbacks[:len(ev.RegistredCallbacks)-1]
		fmt.Printf("Popping %s event.\n", ev.Descriptions[len(ev.Descriptions)-1])
		ev.Descriptions = ev.Descriptions[:len(ev.Descriptions)-1]
	}
}

func Trigger(kind EventKind, ctx EventContext) bool {
	ev, ok := EventMap[kind]
	if !ok {
		return false
	}
	result := false
	for i := len(ev.RegistredCallbacks) - 1; i >= 0; i-- {
		result = ev.RegistredCallbacks[i](ctx)
		if result {
			break
		}
	}
	return result
}
