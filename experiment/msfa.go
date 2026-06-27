package main

import (
	"os"
	"runtime"
	"trkr/external/msfa"
	"unsafe"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	SampleRate = 48000
	BlockSize  = 4096
)

func main() {
	runtime.LockOSThread()

	rl.InitWindow(400, 200, "trkr - Double Stride Link")
	defer func() {
		rl.CloseAudioDevice()
		rl.CloseWindow()
	}()
	rl.SetAudioStreamBufferSizeDefault(BlockSize)
	rl.InitAudioDevice()

	synth := msfa.NewSynth(SampleRate)
	defer synth.Free()

	sayAgainSysEx, _ := os.ReadFile("assets/Dexed_01.syx")
	synth.WriteMidi(sayAgainSysEx)

	// Stream initialized at 32-bit Float Mono
	stream := rl.LoadAudioStream(SampleRate, 32, 1)
	defer rl.UnloadAudioStream(stream)
	rl.PlayAudioStream(stream)

	msfaBuffer := make([]int16, BlockSize)

	// Create the float slice with DOUBLE the element footprint
	// to satisfy the 4-byte float width requirement in C
	raylibBuffer := make([]float32, BlockSize*2)

	noteActive := false
	rl.SetTargetFPS(60)
	msfa.ChangeVolume(0, 10)
	for !rl.WindowShouldClose() {
		// Secure note triggering to avoid auto-repeat resets
		if rl.IsKeyDown(rl.KeySpace) && !noteActive {
			synth.WriteMidi([]byte{0x90, 60, 100})
			noteActive = true
		}
		if rl.IsKeyReleased(rl.KeySpace) {
			synth.WriteMidi([]byte{0x80, 60, 0})
			noteActive = false
		}

		if rl.IsAudioStreamProcessed(stream) {
			synth.GetSamples(msfaBuffer)

			// Clean 1:1 Mono Mapping
			for i := 0; i < BlockSize; i++ {
				raylibBuffer[i] = float32(msfaBuffer[i]) / 32768.0
			}

			// Pass the exact BlockSize since 1 sample = 1 frame in Mono.
			// We match the length exactly to what we allocated.
			unsafeData := (*[1 << 30]float32)(unsafe.Pointer(&raylibBuffer[0]))[:BlockSize:BlockSize]

			rl.UpdateAudioStream(stream, unsafeData)
		}
		rl.BeginDrawing()
		rl.ClearBackground(rl.GetColor(0x101010FF))
		rl.DrawText("Press SPACE to hear FM Synthesis", 40, 90, 18, rl.RayWhite)
		rl.EndDrawing()
	}
}
