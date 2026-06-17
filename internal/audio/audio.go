package audio

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"

	//	"math/rand"
	//	"time"
	. "trkr"
)

func Init() {
	fmt.Print("Initializing...")
	initPitchTable()
	rl.SetAudioStreamBufferSizeDefault(VoiceBufferSize)
	rl.InitAudioDevice() // Initialize audio device
	InitVoiceMasterStream()
}

// Global or package-level lookup table
const AliasPoolSize = 32
const A4Hertz = 440.0
const A4Midi = 69

var pitchTable [120]float64
var frequenciesTable [120]float64

func initPitchTable() {
	// Pre-calculate pitch multipliers for relative note offsets from -128 to 127

	for i := range pitchTable {
		distance := float64(i - A4Midi)
		pitchTable[i] = math.Pow(2.0, distance/12.0)
		frequenciesTable[i] = A4Hertz * math.Pow(2.0, distance/12.0)
	}
}

func InitializeAliases(trackId int, track *Track) {
	// var volume float32 = 1.0
	// track.Sample.Sound = rl.LoadSound(track.Sample.SampleFile)
	// for i := 0; i < AliasPoolSize; i++ {
	// 	aliasesPool[trackId][i] = rl.LoadSoundAlias(track.Sample.Sound)
	// 	rl.SetSoundVolume(aliasesPool[trackId][i], volume)
	// }
}

var aliasesPool [9][AliasPoolSize]rl.Sound
var aliasesPlaying [9][AliasPoolSize]bool
var nextAlias = [9]int{}

func GetPitch(note Note, rootNote Note) float64 {
	// Calculate the offset and map it to our 0-255 table index
	offset := (note - rootNote) + 128

	// Bounds check to prevent panics if notes go wildly out of range
	if offset < 0 {
		return pitchTable[0]
	}
	if offset > 119 {
		return pitchTable[119]
	}

	return pitchTable[offset]
}

func PlaySound(sound rl.Sound) {
	rl.PlaySound(sound)
}

func PlaySoundMulti(columnId uint8, trackId uint8, note Note) {
	// Inside your Player goroutine:
	CommandQueue <- VoiceCommand{
		TrackID:  trackId,
		ColumnID: columnId,
		Note:     note,
		Waveform: WaveSawtooth,
	}
}

func Cleanup() {
	for track := range aliasesPool {
		for i := range aliasesPool[track] {
			if rl.IsSoundPlaying(aliasesPool[track][i]) {
				rl.StopSound(aliasesPool[track][i])
			}
		}
	}
	for trackId := range CurrentProject.Tracks {
		CurrentProject.Tracks[trackId].Cleanup()
	}

	rl.CloseAudioDevice()
}

func waveHeader(buf *bytes.Buffer, nSamples uint32, sampleRate uint32) {
	// 1. Write the standard RIFF/WAVE header fields
	buf.Write([]byte("RIFF"))                                   // ChunkID
	binary.Write(buf, binary.LittleEndian, uint32(36+nSamples)) // ChunkSize
	buf.Write([]byte("WAVE"))                                   // Format
	buf.Write([]byte("fmt "))                                   // Subchunk1ID
	binary.Write(buf, binary.LittleEndian, uint32(16))          // Subchunk1Size (16 for PCM)
	binary.Write(buf, binary.LittleEndian, uint16(1))           // AudioFormat (1 for PCM)
	binary.Write(buf, binary.LittleEndian, uint16(1))           // NumChannels (1 = Mono)
	binary.Write(buf, binary.LittleEndian, sampleRate)          // SampleRate
	binary.Write(buf, binary.LittleEndian, sampleRate*1)        // ByteRate (SampleRate * NumChannels * BitsPerSample/8)
	binary.Write(buf, binary.LittleEndian, uint16(1))           // BlockAlign (NumChannels * BitsPerSample/8)
	binary.Write(buf, binary.LittleEndian, uint16(8))           // BitsPerSample (8 bits)
	buf.Write([]byte("data"))                                   // Subchunk2ID
	binary.Write(buf, binary.LittleEndian, nSamples)            // Subchunk2Size

}

func DemoNotes() {

	// v := NewVoice(ParseNote("C 3") + 21)
	// v.Play()
	// v2 := NewVoice(ParseNote("D#3") + 21)
	// v2.Volume = 0.3
	// v2.Play()
	// v3 := NewVoice(ParseNote("G 3") + 21)
	// v3.Volume = 0.3
	// v3.Play()
	// v4 := NewVoice(ParseNote("C 2") + 21)
	// v4.Waveform = WaveCosine
	// v4.Volume = 0.9
	// v4.Play()
	// go (func() {
	// 	i := 0
	// 	notes := [...]Note{ParseNote("G 3") + 21, ParseNote("F 3") + 21}
	// 	for {
	// 		// 1. Let the note play sustained for 2.5 seconds
	// 		time.Sleep(2500 * time.Millisecond)
	//
	// 		v3.Stop() // Start the 400ms fade-out
	//
	// 		if v4.IsPlaying() {
	// 			v4.Stop()
	// 		} else {
	// 			v4.Play()
	// 		}
	//
	// 		// 2. Wait 500ms for the tail to fade into the background
	// 		time.Sleep(500 * time.Millisecond)
	//
	// 		// 3. Strike the next note clean
	// 		v3.Note = notes[i%len(notes)]
	// 		v3.Play()
	// 		i++
	// 	}
	// })()

}
