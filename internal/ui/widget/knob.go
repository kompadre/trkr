package widget

import (
	"fmt"
	"math"
	"trkr"
	ev "trkr/internal/events"
	"trkr/internal/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Knob struct {
	Label    string
	Value    float64
	MinVal   float64
	MaxVal   float64
	Step     float64
	Action   func(value float64) bool
	OnChange func(value float64)
}

func NewKnob(label string, col [4]int, minVal, maxVal, initialVal float64, action func(value float64) bool, parent *ui.Element) *Knob {
	k := &Knob{
		Label:  label,
		MinVal: minVal,
		MaxVal: maxVal,
		Value:  trkr.Clamp(initialVal, minVal, maxVal),
		Step:   (maxVal - minVal) / 100.0,
		Action: action,
	}
	if k.Step == 0 {
		k.Step = 1.0
	}
	el := ui.NewElement(0, 0, 0, 60, k, parent)
	el.Col = col
	el.Core = k
	return k
}

func (k *Knob) Show() {}

func (k *Knob) Hide() {}

func (k *Knob) HandleInput(input *ev.InputSnapshot, el *ui.Element) bool {
	if input.Tick(ev.InputKindB) == 1 {
		return false // Deactivate
	}
	if input.Tick(ev.InputKindA) == 1 || input.Tick(ev.InputKindEnter) == 1 {
		if k.Action != nil {
			k.Action(k.Value)
		}
		return false // Deactivate
	}

	if input.Down(ev.InputKindUp) || input.Down(ev.InputKindRight) || input.Down(ev.InputKindDown) || input.Down(ev.InputKindLeft) || input.Down(ev.InputKindDir) {
		step := k.Step

		oldValue := k.Value
		if input.Down(ev.InputKindUp) || input.Down(ev.InputKindRight) {
			if input.Tick(ev.InputKindUp) > 10 || input.Tick(ev.InputKindRight) > 10 {
				step *= 5
			}
			k.Value = trkr.Clamp(k.Value+step, k.MinVal, k.MaxVal)
		} else if input.Down(ev.InputKindDown) || input.Down(ev.InputKindLeft) {
			if input.Tick(ev.InputKindDown) > 10 || input.Tick(ev.InputKindLeft) > 10 {
				step *= 5
			}
			k.Value = trkr.Clamp(k.Value-step, k.MinVal, k.MaxVal)
		}

		if k.Value != oldValue {
			if k.OnChange != nil {
				k.OnChange(k.Value)
			}
			return true
		}
	}
	return true // Keep focus while editing
}

func (k *Knob) Draw(ctx ev.EventContext, hasFocus bool, isHighlighted bool) bool {
	p, ok := ctx.EventPayload.(*ui.ElementDrawPayload)
	if !ok || p == nil {
		return false
	}

	rec := p.Laid.Bounds()
	p.Laid.SetRowHeight(60)

	// Colors
	bgcolor := ui.WindowBg1
	fgColor := ui.WindowFg1
	activeColor := ui.WindowFg1
	trackColor := trkr.RGBA(150, 150, 160, 255)
	indicatorColor := rl.Red

	if isHighlighted {
		bgcolor = ui.WindowBg2
		if hasFocus {
			bgcolor = ui.WindowBg3
			fgColor = ui.WindowFg3
			activeColor = ui.WindowFg3
			trackColor = trkr.RGBA(60, 120, 20, 255)
			indicatorColor = rl.Yellow
		}
	}

	rl.DrawRectangleRec(rec, bgcolor)
	if isHighlighted {
		lineColor := trkr.RGBA(255, 255, 255, 80)
		if hasFocus {
			lineColor = rl.Yellow
		}
		rl.DrawRectangleLinesEx(rec, 1, lineColor)
	}

	// Layout knob circle and text inside rec
	// Knob circle centered on the left half, label & value on the right
	radius := float32(math.Min(float64(rec.Width/3), float64(rec.Height/2-10)))
	if radius < 12 {
		radius = 12
	}

	center := rl.NewVector2(rec.X+radius+12, rec.Y+rec.Height/2)

	// Horseshoe geometry: 270 degree sweep with bottom 90 degree gap
	// 135 deg = bottom-left, 405 deg (45 deg) = bottom-right
	startAngle := float32(135.0)
	totalSweep := float32(270.0)

	// Normalized value t in [0.0, 1.0]
	t := float32(0.0)
	if k.MaxVal > k.MinVal {
		t = float32(k.Value-k.MinVal) / float32(k.MaxVal-k.MinVal)
	}
	t = float32(trkr.Clamp(float64(t), 0.0, 1.0))

	currentAngle := startAngle + (t * totalSweep)

	innerRadius := radius * 0.6
	outerRadius := radius

	// Draw background arc (full horseshoe)
	rl.DrawRing(center, innerRadius, outerRadius, startAngle, startAngle+totalSweep, 48, trackColor)

	// Draw active value arc (from min to current)
	if t > 0.001 {
		rl.DrawRing(center, innerRadius, outerRadius, startAngle, currentAngle, 48, activeColor)
	}

	// Draw indicator needle line from center towards currentAngle
	rad := float64(currentAngle) * math.Pi / 180.0
	needleStart := rl.NewVector2(
		center.X+float32(math.Cos(rad))*innerRadius*0.4,
		center.Y+float32(math.Sin(rad))*innerRadius*0.4,
	)
	needleEnd := rl.NewVector2(
		center.X+float32(math.Cos(rad))*outerRadius*1.05,
		center.Y+float32(math.Sin(rad))*outerRadius*1.05,
	)
	rl.DrawLineEx(needleStart, needleEnd, 2.0, indicatorColor)

	// Draw label and numeric value text
	textX := int32(center.X + outerRadius + 12)
	prefix := ""
	if isHighlighted {
		prefix = ">"
		if hasFocus {
			prefix = "*"
		}
	}

	ui.DrawText(prefix+k.Label, textX, int32(center.Y-14), 14, fgColor)
	valStr := ""
	if k.MaxVal <= 1.0 {
		valStr = fmt.Sprintf("%.3f", k.Value)
	} else if k.MaxVal <= 10.0 {
		valStr = fmt.Sprintf("%.2f", k.Value)
	} else {
		valStr = fmt.Sprintf("%.0f", k.Value)
	}
	ui.DrawText(valStr, textX, int32(center.Y+2), 14, activeColor)

	return false
}
