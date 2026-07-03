package audio

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"

	//	"math/rand"
	//	"time"
	"path"
	. "trkr"
	"trkr/external/msfa"
)

func Init() {
	fmt.Print("Initializing...")
	initPitchTable()
	rl.SetAudioStreamBufferSizeDefault(VoiceBufferSize)
	rl.InitAudioDevice() // Initialize audio device
	InitVoiceMasterStream()
	if CurrentProject.FmPatchName == "" {
		CurrentProject.FmPatchName = "./assets/syx/Dexed_01.syx"
	}
	SynthInstance = msfa.NewSynth(VoiceSampleRate)
	SynthInstance.LoadPatch(CurrentProject.FmPatchName)
	synthPlayingNotes = make(map[NoteGridCoord]Note)
	for _, t := range CurrentProject.Tracks {
		if t.Instrument.SampleSourceType == SampleSourceTypeFm {
			SynthInstance.ChangeProgram(t.Id, t.Instrument.Program)
		}
	}
	PlayFmPickup()
}

func PlayFmPickup() {
	// Special synth voice that picks up already mixed multitimbral samples from MSFA
	if VoiceFm == nil {
		for i := range voicePool {
			if voicePool[i].Instrument != nil && voicePool[i].Instrument.SampleSourceType == SampleSourceTypeFmPickup {
				VoiceFm = voicePool[i]
				break
			}
		}
	}
	if VoiceFm == nil {
		VoiceFm = FindFreeVoice(0xff, 0xff, 0xff)
	}

	var instrument *Instrument
	for i := range CurrentProject.Instruments {
		if CurrentProject.Instruments[i].SampleSourceType == SampleSourceTypeFmPickup {
			instrument = &CurrentProject.Instruments[i]
			break
		}
	}

	if instrument == nil {
		instrument = NewInstrument(CurrentProject)
		instrument.SampleSourceType = SampleSourceTypeFmPickup
	}

	VoiceFm.Volume = 1.0
	VoiceFm.TrackId = 0xff
	VoiceFm.ColumnId = 0xff
	VoiceFm.Instrument = instrument
	VoiceFm.InstrumentId = instrument.Id
	VoiceFm.Play()
}

// Global or package-level lookup table
const AliasPoolSize = 32
const A4Hertz = 440.0
const A4Midi = 69

type NoteGridCoord struct {
	TrackId  uint8
	ColumnId uint8
}

var pitchTable [120]float64
var frequenciesTable [120]float64
var SynthInstance *msfa.Synth
var synthPlayingNotes map[NoteGridCoord]Note
var VoiceFm *Voice

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
	if offset > 119 {
		return pitchTable[119]
	}

	return pitchTable[offset]
}

func PlaySound(sound rl.Sound) {
	rl.PlaySound(sound)
}

func PlaySoundMulti(columnId uint8, trackId uint8, note Note, instrumentId uint8, sampleSourceType SampleSourceType, velocity uint8, volume float64) {
	sustainTicks := 0.5 * (VoiceSampleRate / (float64(BeatsPerMinute) / 60))
	//sustainTicks := 4 * 24
	CommandQueue <- VoiceCommand{
		TrackId:          trackId,
		ColumnId:         columnId,
		InstrumentId:     instrumentId,
		Note:             note,
		SustainTicks:     uint32(sustainTicks),
		SampleSourceType: sampleSourceType,
		Velocity:         Clamp(velocity, 0, 127),
		Volume:           Clamp(volume, 0.0, 1.0),
	}
}

func SynthPatchName() string {
	return path.Base(SynthInstance.PatchPath)
}

func SynthProgramName(ch uint8) string {
	if SynthInstance == nil || len(SynthInstance.Programs) < 1 {
		return "-"
	}
	return SynthInstance.Programs[SynthInstance.ChannelPrograms[ch]]
}

func StopSoundFm(columnId uint8, trackId uint8) {
	noteOffCmd := 0x80 + trackId
	coord := NoteGridCoord{TrackId: trackId, ColumnId: columnId}
	if notePlaying, isPlaying := synthPlayingNotes[coord]; isPlaying {
		SynthInstance.WriteMidi([]byte{noteOffCmd, byte(notePlaying), 127})
		delete(synthPlayingNotes, coord)
	}
}

func StopFm() {
	for k := range synthPlayingNotes {
		StopSoundFm(k.ColumnId, k.TrackId)
	}
	// Drain synth's samples
	var drainBuffer [512]int16
	for range 512 {
		SynthInstance.GetSamples(drainBuffer[:])
	}
	VoiceFm.Envelope.State = EnvIdle
}

func PlaySoundFm(columnId uint8, trackId uint8, note Note, velocity uint8) {
	if SynthInstance == nil {
		return // Protect against uninitialized global pointer
	}

	coord := NoteGridCoord{TrackId: trackId, ColumnId: columnId}
	notePlaying, isPlaying := synthPlayingNotes[coord]

	// Calculate correct MIDI command nibbles dynamically per track
	noteOnCmd := 0x90 + trackId
	noteOffCmd := 0x80 + trackId

	if isPlaying {
		SynthInstance.WriteMidi([]byte{noteOffCmd, byte(notePlaying), 127})
		delete(synthPlayingNotes, coord)
	}

	SynthInstance.WriteMidi([]byte{noteOnCmd, byte(note), velocity})
	synthPlayingNotes[coord] = note
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
	// synth.Free()
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
