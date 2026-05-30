module trkr

go 1.23.1

//replace github.com/gen2brain/raylib-go/raylib => ../raylib-go/raylib
//replace github.com/gen2brain/raylib-go/raygui => ../raylib-go/raygui
//replace github.com/gen2brain/raylib-go/easings => ../raylib-go/easings
//replace github.com/gen2brain/raylib-go/physics => ../raylib-go/physics

require (
	//github.com/faiface/beep v1.1.0
	github.com/gen2brain/raylib-go/raylib v0.0.0-20231118125650-a1c890e8cbfc
	golang.org/x/exp v0.0.0-20240506185415-9bf2ced13842
)

require (
	github.com/ebitengine/purego v0.7.1 // indirect
	golang.org/x/sys v0.20.0 // indirect
)
