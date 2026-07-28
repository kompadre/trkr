package ui

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	ev "trkr/internal/events"
)

type Transition struct {
	Render rl.RenderTexture2D
	Offset float32
	Ttl    int
	Pos    rl.Vector2
	Dir    rl.Vector2
	Scale  float32
	Tint   rl.Color
}

func NewTransition(parent *Element, dir rl.Vector2) *Element {
	t := &Transition{Ttl: 3}

	// image := rl.LoadImage("./assets/images/bg.png")
	//	bg := rl.LoadTextureFromImage(image)
	t.Render = rl.LoadRenderTexture(int32(GetOptions().ScreenWidth), int32(GetOptions().ScreenHeight))
	inverted := rl.LoadRenderTexture(int32(GetOptions().ScreenWidth), int32(GetOptions().ScreenHeight))
	rl.SetTextureFilter(t.Render.Texture, rl.TextureFilterLinear)
	t.Ttl = 3
	t.Dir = dir
	t.Tint = rl.NewColor(255, 255, 255, 255)
	t.Scale = 1.0
	fmt.Printf("Dir: %v\n", t.Dir)
	// rl.UnloadImage(image)

	rl.BeginTextureMode(inverted)
	rl.ClearBackground(WindowBg5)
	rl.DrawRectangleLines(0, 0, int32(GetOptions().ScreenWidth), int32(GetOptions().ScreenHeight), rl.White)
	// rl.DrawTexture(bg, 0, 0, rl.White)
	laid := Laid{}
	laid.SetBreakpoint(rl.GetScreenWidth())
	ev.Trigger(ev.EventKindUpdate, ev.EventContext{EventPayload: &ElementDrawPayload{Laid: &laid}})
	rl.EndTextureMode()

	rl.BeginTextureMode(t.Render)
	rl.DrawTexture(inverted.Texture, 0, 0, rl.White)
	rl.EndTextureMode()

	rl.UnloadRenderTexture(inverted)

	el := NewElement(0, 0, int32(GetOptions().ScreenWidth), int32(GetOptions().ScreenHeight), t, parent)
	parent.SetFocus(el)
	return el
}

func (t *Transition) Show() {}
func (t *Transition) Hide() {}
func (t *Transition) Draw(ctx ev.EventContext, hasFocus bool, isHighlighted bool) bool {
	if t.Dir.X != t.Dir.Y {
		srcRect := rl.NewRectangle(0, 0, float32(GetOptions().ScreenWidth), float32(GetOptions().ScreenHeight))
		rl.DrawTextureRec(t.Render.Texture, srcRect, t.Pos, t.Tint)
	} else {
		rl.DrawTextureEx(t.Render.Texture, t.Pos, 0.0, t.Scale, t.Tint)
	}

	t.Ttl--
	if t.Ttl <= 0 {
		if t.Dir.X == 0 && t.Dir.Y == 0 {
			t.Scale += -0.01
			t.Pos.X = float32(GetOptions().ScreenWidth) * (1 - t.Scale) / 2
			t.Pos.Y = float32(GetOptions().ScreenHeight) * (1 - t.Scale) / 2
		} else if t.Dir.X > 0 && t.Dir.Y > 0 {
			t.Scale += 0.01
			t.Pos.X = float32(GetOptions().ScreenWidth) * (1 - t.Scale) / 2
			t.Pos.Y = float32(GetOptions().ScreenHeight) * (1 - t.Scale) / 2
		} else {
			t.Pos.X += t.Dir.X
			t.Pos.Y += t.Dir.Y
		}

		if t.Tint.A > 10 {
			t.Tint.A -= 10
		} else {
			t.Tint.A = 0
		}

		if t.Ttl < -50 {
			el := ctx.EventPayload.(*ElementDrawPayload).Element
			if el.Visible {
				el.Visible = false
				fmt.Printf("Removing transition...")
				el.Remove()
				rl.UnloadRenderTexture(t.Render)
			}
		}
	}
	return true
}

func (t *Transition) HandleInput(input *ev.InputSnapshot, el *Element) bool {
	fmt.Printf("Handling Input for transition.\n")
	return true
}
