package audio

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"trkr"
	"trkr/internal/audio/mixer"
)

var music rl.Music
var Kick rl.Sound
var Hat rl.Sound
var Snare rl.Sound
var hatVolume float32 = 0.25
var Sounds []rl.Sound
var SoundMap map[trkr.Note]*rl.Sound
var keyPitch float32 = 1.0

func Init() {
	rl.InitAudioDevice() // Initialize audio device
	Sounds = make([]rl.Sound, 4)
	SoundMap = make(map[trkr.Note]*rl.Sound)

	Sounds[0] = rl.LoadSound("./assets/music/key.wav")
	Sounds[1] = rl.LoadSound("./assets/music/kick.wav")
	Sounds[2] = rl.LoadSound("./assets/music/snare.wav")
	Sounds[3] = rl.LoadSound("./assets/music/hat.wav")

	SoundMap[26] = &Sounds[0]
	SoundMap[1] = &Sounds[1]
	SoundMap[3] = &Sounds[2]
	SoundMap[5] = &Sounds[3]

}

func PlaySound(sound rl.Sound) {
	if sound.FrameCount == 0 {
		fmt.Println("sound is empty")
	}
	rl.PlaySound(sound)
}

func CleanUp() {
	rl.UnloadMusicStream(music)
	rl.UnloadSound(Kick)
	rl.CloseAudioDevice()
}

func Update() {
	if rl.IsMusicReady(music) {
		rl.UpdateMusicStream(music)
		rl.UpdateAudioStream(music.Stream, mixer.ExposedDelayBuffer)
		fmt.Println("Updating stream")
	}
}
