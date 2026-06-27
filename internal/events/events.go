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
	EventKindAudioUpdate
	EventKindShutdown
)

const EventGamepadDeadZone = 0.1

type EventContext struct {
	EventData    []byte
	EventPayload any
}

type EventCallback func(EventContext) bool

type Event struct {
	EventKind
	RegistredCallbacks map[uint16]EventCallback
	CallbackKeys       []uint16
}

var EventMap = map[EventKind]*Event{}

type InputStatus uint16

var PrevSnapshot InputStatus

const (
	InputKindLeft     InputStatus = 0
	InputKindUp       InputStatus = 1
	InputKindRight    InputStatus = 2
	InputKindDown     InputStatus = 3
	InputKindA        InputStatus = 4
	InputKindB        InputStatus = 5
	InputKindL        InputStatus = 6
	InputKindM        InputStatus = 7
	InputKindR        InputStatus = 8
	InputKindSpace    InputStatus = 9
	InputKindPageDown InputStatus = 10
	InputKindPageUp   InputStatus = 11
	InputKindEnter    InputStatus = 12
	InputKindDir      InputStatus = 13
	InputKindLast     InputStatus = 14
)

const InputDirMask InputStatus = (1 << InputKindUp) | (1 << InputKindLeft) | (1 << InputKindRight) | (1 << InputKindDown)

var RawInputCodes = [...]InputStatus{
	rl.KeyLeft,
	rl.KeyUp,
	rl.KeyRight,
	rl.KeyDown,
	rl.KeyA,
	rl.KeyB,
	rl.KeyL,
	rl.KeyM,
	rl.KeyR,
	rl.KeySpace,
	rl.KeyPageDown,
	rl.KeyPageUp,
	rl.KeyEnter,
}

type InputSnapshot struct {
	//	Holding    InputStatus
	Pressed    InputStatus
	Released   InputStatus
	HoldTimers [InputKindLast]uint16
}

var PrevHoldTimers [InputKindLast]uint16

func (is *InputSnapshot) ClearHoldTimers(args ...InputStatus) {
	for _, v := range args {
		is.HoldTimers[v] = 0
	}
}

func CalculateInput(result *InputSnapshot) {
	var CurrentSnapshot InputStatus = 0
	gamepadAxisLeftX := rl.GetGamepadAxisMovement(0, rl.GamepadAxisLeftX)
	gamepadAxisLeftY := rl.GetGamepadAxisMovement(0, rl.GamepadAxisLeftY)

	if gamepadAxisLeftX < EventGamepadDeadZone*-1 || gamepadAxisLeftX > EventGamepadDeadZone ||
		gamepadAxisLeftY < EventGamepadDeadZone*-1 || gamepadAxisLeftY > EventGamepadDeadZone {
		if gamepadAxisLeftX < EventGamepadDeadZone*-1 {
			CurrentSnapshot = CurrentSnapshot | InputKindLeft
		}
		if gamepadAxisLeftX > EventGamepadDeadZone {
			CurrentSnapshot = CurrentSnapshot | InputKindRight
		}
		if gamepadAxisLeftY < EventGamepadDeadZone*-1 {
			CurrentSnapshot = CurrentSnapshot | InputKindUp
		}
		if gamepadAxisLeftY > EventGamepadDeadZone*1 {
			CurrentSnapshot = CurrentSnapshot | InputKindDown
		}
	}
	// result := InputSnapshot{HoldTimers: PrevHoldTimers}
	dirIsHeld := false
	for i := range RawInputCodes {
		if rl.IsKeyDown(int32(RawInputCodes[i])) {
			CurrentSnapshot = CurrentSnapshot | (1 << i)
			result.HoldTimers[i]++
			dirIsHeld = dirIsHeld || i <= int(InputKindDown)
		} else {
			result.HoldTimers[i] = 0
		}
	}
	if dirIsHeld {
		result.HoldTimers[InputKindDir]++
	} else {
		result.HoldTimers[InputKindDir] = 0
	}
	if CurrentSnapshot&InputDirMask != 0 {
		CurrentSnapshot = CurrentSnapshot | InputKindDir
	}
	//	result.Holding = CurrentSnapshot
	result.Pressed = CurrentSnapshot & ^PrevSnapshot
	result.Released = ^CurrentSnapshot & PrevSnapshot
	PrevSnapshot = CurrentSnapshot
	PrevHoldTimers = result.HoldTimers
}

func (in InputSnapshot) Tick(code InputStatus) uint16 {
	return in.HoldTimers[code]
}
func (in InputSnapshot) Down(code InputStatus) bool {
	return in.HoldTimers[code] > 0
}
func (in InputSnapshot) Up(code InputStatus) bool {
	return in.Released&(1<<code) != 0
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
