package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestLaidBlockWidthTracking(t *testing.T) {
	laid := Laid{}
	laid.PushContext(rl.Rectangle{X: 0, Y: 0, Width: 480, Height: 320})
	laid.SetBreakpoint(480)

	if laid.BlockWidth() != 480 {
		t.Fatalf("expected BlockWidth 480, got %f", laid.BlockWidth())
	}
	if laid.AllocatedWidth() != 0 {
		t.Fatalf("expected AllocatedWidth 0, got %f", laid.AllocatedWidth())
	}
	if laid.RemainingWidth() != 480 {
		t.Fatalf("expected RemainingWidth 480, got %f", laid.RemainingWidth())
	}

	// Allocate 6 of 12 columns
	bounds, ok := laid.Col(6, 6, 6, 6)
	if !ok {
		t.Fatalf("expected Col to succeed")
	}
	if bounds.Width != 240 {
		t.Fatalf("expected column width 240, got %f", bounds.Width)
	}
	if laid.AllocatedWidth() != 240 {
		t.Fatalf("expected AllocatedWidth 240, got %f", laid.AllocatedWidth())
	}
	if laid.RemainingWidth() != 240 {
		t.Fatalf("expected RemainingWidth 240, got %f", laid.RemainingWidth())
	}

	// Allocate remaining 6 columns (totaling 12)
	bounds2, ok := laid.Col(6, 6, 6, 6)
	if !ok {
		t.Fatalf("expected Col to succeed")
	}
	if bounds2.Width != 240 {
		t.Fatalf("expected column width 240, got %f", bounds2.Width)
	}
	if laid.AllocatedWidth() != 480 {
		t.Fatalf("expected AllocatedWidth 480, got %f", laid.AllocatedWidth())
	}
	if laid.RemainingWidth() != 0 {
		t.Fatalf("expected RemainingWidth 0, got %f", laid.RemainingWidth())
	}
}

func TestLaidLayoutCaching(t *testing.T) {
	laid := Laid{}
	laid.BeginFrame(480, 320)

	id := uint16(42)
	rect := rl.Rectangle{X: 10, Y: 20, Width: 100, Height: 50}

	laid.SetCachedBounds(id, rect)
	cached, ok := laid.GetCachedBounds(id)
	if !ok || cached != rect {
		t.Fatalf("expected cached bounds %v, got %v (ok=%v)", rect, cached, ok)
	}

	// Begin frame with same dimensions should keep cache
	laid.BeginFrame(480, 320)
	cached, ok = laid.GetCachedBounds(id)
	if !ok || cached != rect {
		t.Fatalf("expected cache retention on same frame dims")
	}

	// Window resize should invalidate cache
	laid.BeginFrame(800, 600)
	_, ok = laid.GetCachedBounds(id)
	if ok {
		t.Fatalf("expected cache invalidation after resize")
	}
}

func TestElementTreeLayoutIntegration(t *testing.T) {
	laid := Laid{}
	laid.BeginFrame(480, 320)

	root := NewElement(0, 0, 480, 320, nil, nil)
	root.Col = [4]int{12, 12, 12, 12}

	child1 := NewElement(0, 0, 0, 40, nil, root)
	child1.Col = [4]int{6, 6, 6, 6}

	child2 := NewElement(0, 0, 0, 40, nil, root)
	child2.Col = [4]int{6, 6, 6, 6}

	computedRoot := laid.ComputeElementLayout(root)
	if computedRoot.Width != 480 {
		t.Fatalf("expected root width 480, got %f", computedRoot.Width)
	}

	if child1.Bounds().Width != 240 {
		t.Fatalf("expected child1 width 240, got %f", child1.Bounds().Width)
	}
	if child2.Bounds().Width != 240 {
		t.Fatalf("expected child2 width 240, got %f", child2.Bounds().Width)
	}
}
