module trkr

go 1.25

//replace github.com/gen2brain/raylib-go/raylib => ../raylib-go/raylib
//replace github.com/gen2brain/raylib-go/raygui => ../raylib-go/raygui
//replace github.com/gen2brain/raylib-go/easings => ../raylib-go/easings
//replace github.com/gen2brain/raylib-go/physics => ../raylib-go/physics

require (
	//github.com/faiface/beep v1.1.0
	github.com/gen2brain/raylib-go/raylib v0.60.0
	golang.org/x/exp v0.0.0-20240506185415-9bf2ced13842
)

replace github.com/gen2brain/raylib-go/raylib => ./external/raylib-go/raylib

require (
	github.com/ebitengine/purego v0.10.0 // indirect
	github.com/jupiterrider/ffi v0.7.0 // indirect
)
