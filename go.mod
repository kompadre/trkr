module trkr

go 1.25.0

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
	github.com/c2h5oh/datasize v0.0.0-20231215233829-aa82cc1e6500 // indirect
	github.com/ebitengine/purego v0.10.0 // indirect
	github.com/influxdata/tdigest v0.0.1 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jupiterrider/ffi v0.7.0 // indirect
	github.com/mailru/easyjson v0.9.2 // indirect
	github.com/tsenart/go-tsz v0.0.0-20180814235614-0bd30b3df1c3 // indirect
	github.com/tsenart/vegeta v12.7.0+incompatible // indirect
	golang.org/x/net v0.57.0 // indirect
	golang.org/x/text v0.40.0 // indirect
)
