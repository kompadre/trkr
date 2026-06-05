package audio

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"math"
	"trkr"
)

func Init() {
	fmt.Print("Initializing...")
	initPitchTable()
	rl.SetAudioStreamBufferSizeDefault(4096)
	rl.InitAudioDevice() // Initialize audio device
}

// Global or package-level lookup table
const ALIAS_POOL_SIZE = 8

var pitchTable [256]float64

func initPitchTable() {
	// Pre-calculate pitch multipliers for relative note offsets from -128 to 127
	for i := 0; i < 256; i++ {
		relativeNote := float64(i - 128)
		pitchTable[i] = math.Pow(2.0, relativeNote/12.0)
	}
}

func InitializeAliases(trackId int, track *trkr.Track) {
	track.Sample.Sound = rl.LoadSound(track.Sample.SampleFile)
	for i := 0; i < ALIAS_POOL_SIZE; i++ {
		aliasesPool[trackId][i] = rl.LoadSoundAlias(track.Sample.Sound)
	}
}

var aliasesPool [3][ALIAS_POOL_SIZE]rl.Sound
var aliasesPlaying [3][ALIAS_POOL_SIZE]bool
var nextAlias = [3]int{0, 0, 0}

func GetPitch(note trkr.Note, rootNote trkr.Note) float64 {
	// Calculate the offset and map it to our 0-255 table index
	offset := (note - rootNote) + 128

	// Bounds check to prevent panics if notes go wildly out of range
	if offset < 0 {
		return pitchTable[0]
	}
	if offset > 255 {
		return pitchTable[255]
	}

	return pitchTable[offset]
}

func PlaySound(sound rl.Sound) {
	rl.PlaySound(sound)
}

func PlaySoundMulti(trackId int, pitch float32) {
	nextAlias[trackId]++
	if nextAlias[trackId] >= ALIAS_POOL_SIZE {
		nextAlias[trackId] = 0
	}
	trackNextAlias := nextAlias[trackId]
	if aliasesPlaying[trackId][trackNextAlias] {
		rl.StopSound(aliasesPool[trackId][trackNextAlias])
	}
	rl.SetSoundPitch(aliasesPool[trackId][trackNextAlias], pitch)
	rl.PlaySound(aliasesPool[trackId][trackNextAlias])
	aliasesPlaying[trackId][trackNextAlias] = true
}

func CleanUp() {
	fmt.Println("Cleaning up audio")
	for track := range aliasesPool {
		for i := range aliasesPool[track] {
			if rl.IsSoundPlaying(aliasesPool[track][i]) {
				rl.StopSound(aliasesPool[track][i])
			}
			if aliasesPool[track][i].FrameCount != 0 {
				// rl.UnloadSound(aliasesPool[track][i])
			}
		}
	}
	for trackId := range trkr.CurrentProject.Tracks {
		if trkr.CurrentProject.Tracks[trackId].Sample.Sound.FrameCount != 0 {
			rl.UnloadSound(trkr.CurrentProject.Tracks[trackId].Sample.Sound)
		}
		for sampleId := range trkr.CurrentProject.Tracks[trackId].Samples {
			if trkr.CurrentProject.Tracks[trackId].Samples[sampleId].Sound.FrameCount != 0 {
				rl.UnloadSound(trkr.CurrentProject.Tracks[trackId].Samples[sampleId].Sound)
			}
		}
	}

	rl.CloseAudioDevice()
}
