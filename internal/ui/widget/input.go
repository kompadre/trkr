package widget

import (
	"fmt"
	"strconv"
	"strings"
	"trkr"
	ev "trkr/internal/events"
	"trkr/internal/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type WidgetInputType uint8

const (
	WidgetInputTypeString WidgetInputType = iota
	WidgetInputTypeNumber
)

type Input struct {
	Action      ui.Action
	Label       string
	Value       [8]rune
	ValueStr    string
	WidgetType  WidgetInputType
	FocusedChar uint8
}

func NewInput(label string, relativeLeft int32, relativeTop int32, width int32, height int32, action ui.Action, parent *ui.Element) *Input {
	fmt.Printf("Creating input l:%d, t:%d, w:%d, h:%d.\n", relativeLeft, relativeTop, width, height)
	in := &Input{Label: label, Action: action, ValueStr: " ", Value: [8]rune{' '}}

	// func NewElement(left int32, top int32, width int32, height int32, core ElementCore, parent *Element) *Element {
	el := ui.NewElement(relativeLeft, relativeTop, width, height, in, parent)
	fmt.Printf("Element is %v\n", el)
	return in
}

func (in Input) Show() {}

func (in Input) Hide()    {}
func (in Input) Cleanup() {}

func (in *Input) HandleInput(input ev.InputSnapshot, el *ui.Element) bool {
	switch in.WidgetType {
	case WidgetInputTypeNumber:
		return in.HandleInputNumber(input, el)
	default:
		return in.HandleInputText(input, el)
	}
}

func (in *Input) SetValue(value string) {
	in.ValueStr = value
	for i, v := range value {
		if i >= 8 {
			break
		}

		in.Value[i] = v
	}
}

func (in *Input) HandleInputNumber(input ev.InputSnapshot, el *ui.Element) bool {
	if input.Down(ev.InputKindDir) {
		var err error
		var inputValue int
		if inputValue, err = strconv.Atoi(in.ValueStr); err != nil {
			inputValue = 0
		}
		if input.Down(ev.InputKindDown) {
			if input.Tick(ev.InputKindDown) > 10 {
				inputValue -= 10
			} else {
				inputValue--
			}
		} else if input.Down(ev.InputKindUp) {
			if input.Tick(ev.InputKindUp) > 10 {
				inputValue += 10
			} else {
				inputValue++
			}
		}

		if inputValue > 60 && inputValue < 300 {
			trkr.BeatsPerMinute = inputValue
			in.ValueStr = strconv.Itoa(inputValue)
			for i, runeValue := range in.ValueStr {
				in.Value[i] = runeValue
			}
		}
		return true
	} else if input.Down(ev.InputKindEnter) {
		if in.Action != nil {
			if !in.Action(in.ValueStr) {
				return false
			}
		}
		el.Parent.FocusJump(1)
		return true
	}

	return false
}

func (in *Input) HandleInputText(input ev.InputSnapshot, el *ui.Element) bool {
	if input.Down(ev.InputKindDir) {
		if input.Down(ev.InputKindUp) {
			in.Value[in.FocusedChar]++
		} else if input.Down(ev.InputKindDown) {
			in.Value[in.FocusedChar]--
		} else if input.Down(ev.InputKindLeft) && in.FocusedChar > 0 {
			in.FocusedChar--
		} else if input.Down(ev.InputKindRight) && in.FocusedChar < 7 {
			if in.Value[in.FocusedChar] == '\x00' {
				in.Value[in.FocusedChar] = ' '
			}
			in.FocusedChar++
		}
		if in.Value[in.FocusedChar] == '\x01' || in.Value[in.FocusedChar] == '\x00' || in.Value[in.FocusedChar] > 'Z' {
			in.Value[in.FocusedChar] = ' '
		} else if in.Value[in.FocusedChar] == (1 + ' ') {
			in.Value[in.FocusedChar] = 'A'
		} else if in.Value[in.FocusedChar] == (-1 + ' ') {
			in.Value[in.FocusedChar] = 'Z'
		} else if in.Value[in.FocusedChar] < '0' {
			in.Value[in.FocusedChar] = 'A'
		}
		in.ValueStr = strings.TrimRight(string(in.Value[:]), "\x00")
		fmt.Printf("ValueStr: %v, FocusedChar: %d.\n", in.ValueStr, in.FocusedChar)
		return true
	} else if input.Down(ev.InputKindEnter) {
		if in.Action != nil {
			if !in.Action(in.ValueStr) {
				return false
			}
		}
		el.Parent.FocusJump(1)
		return true
	}
	return false
}

func (in *Input) Draw(ctx ev.EventContext, hasFocus bool) bool {
	var left, top, width, height int32
	p := ctx.EventPayload.(*ui.ElementDrawPayload)
	if p != nil {
		left = p.Left + p.Element.Left
		top = p.Top + p.Element.Top
		width = p.Element.Width
		height = p.Element.Height
	}

	bgcolor := ui.InputBg1
	fgcolor := ui.WindowFg1
	preffix := ""
	if hasFocus {
		preffix = ">"
	}

	ui.DrawText(preffix+in.Label, left, top+2, 20, fgcolor)
	rl.DrawRectangle(left+60, top, width-60, height, bgcolor)
	var extraLeft int32 = 65
	var underLeft int32
	for i := range 8 {
		if in.Value[i] == 'I' {
			extraLeft += 0
		}
		ui.DrawText(string(in.Value[i]), left+extraLeft, top+2, 20, fgcolor)
		if i == int(in.FocusedChar) {
			underLeft = extraLeft
		}

		extraLeft += 10

		if in.Value[i] == 0 {
			break
		}
	}
	// ui.DrawText(in.ValueStr, left+84, top+2, 20, fgcolor)
	if hasFocus {
		rl.DrawText("_", left+underLeft, top+6, 20, fgcolor)
	}
	return false
}
