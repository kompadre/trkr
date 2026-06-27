package trkr

import (
	"encoding/json"
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"golang.org/x/exp/constraints"
	"os"
	"strconv"
)

const (
	MaxNotesInStep    = 8
	MaxStepsInPhrase  = 16
	MaxEffectsInStep  = 4
	SemitonesInOctave = 12
)

var BeatsPerMinute = 105

type Cleanable interface {
	Cleanup()
}

type Note uint8

const (
	NoteNone Note = 0x00 // Cell is empty / No action this step
	// Standard notes reside cleanly between 1 and 127
	NoteSkip Note = 0xFD // Skip to end
	NoteOff  Note = 0xFE // Cut the note on this specific column
	NoteCut  Note = 0xFF // Hard stop the voice immediately (bypass release envelope)
)

type Effect uint8

type Step struct {
	Notes   [MaxNotesInStep]Note
	Effects [MaxEffectsInStep]Effect
}

type Phrase struct {
	ID            int32
	CurrentStep   int `json:"-"`
	Repeats       uint8
	CurrentRepeat uint8 `json:"-"`
	Steps         [MaxStepsInPhrase]Step
}

func NewPhrase(pr *Project) *Phrase {
	pr.Phrases = append(pr.Phrases, Phrase{ID: int32(len(pr.Phrases))})
	return &pr.Phrases[len(pr.Phrases)-1]
}

func NewInstrument(pr *Project) *Instrument {
	id := uint8(len(pr.Instruments))
	pr.Instruments = append(pr.Instruments, Instrument{Id: id})
	return &pr.Instruments[len(pr.Instruments)-1]
}

func (p *Phrase) Clone() *Phrase {
	result := *p
	result.ID = int32(len(CurrentProject.Phrases))
	CurrentProject.Phrases = append(CurrentProject.Phrases, result)
	return &result
}

type _Sample struct {
	Loaded     bool `json:"-"`
	SampleFile string
	Sound      rl.Sound `json:"-"`
	RootNote   Note
	Samples    []float32
}

//go:generate stringer -type=SampleSourceType -trimprefix=SampleSourceType
type SampleSourceType uint8

const (
	SampleSourceTypeWavefile SampleSourceType = iota
	SampleSourceTypeSquare
	SampleSourceTypeSawtooth
	SampleSourceTypeCosine
	SampleSourceTypeFm
	SampleSourceTypeFmPickup
)

const (
	VoiceSampleRate = 48000.0
	VoiceBufferSize = 2048 // 1440 // 512 // 512 // 4096
	VoicePoolSize   = 16
)

func (sst SampleSourceType) UiString() string {
	switch sst {
	case SampleSourceTypeWavefile:
		return "Wavefile"
	case SampleSourceTypeSquare:
		return "Square"
	case SampleSourceTypeCosine:
		return "Cosine"
	case SampleSourceTypeSawtooth:
		return "Sawtooth"
	case SampleSourceTypeFm:
		return "FM"
	case SampleSourceTypeFmPickup:
		return "FM pickup"
	}
	return "Dunno"
}

type Instrument struct {
	Id               uint8
	IsMulti          bool
	RootNote         Note
	SampleSourceType SampleSourceType
	SampleSource     string
	Samples          []float32 `json:"-"`
	LoopStart        float64
	LoopEnd          float64
	Instruments      []*Instrument `json:"-"`
	InstrumentIds    []int
	SamplesLoaded    bool `json:"-"`
	Program          uint8
}

func (in *Instrument) LoadSamples() {
	Logf("Loading samples for instrument %d.\n", in.Id)
	wav := rl.LoadWave(in.SampleSource)
	rl.WaveFormat(&wav, VoiceSampleRate, 32, 1)
	defer rl.UnloadWave(wav)
	tmp := rl.LoadWaveSamples(wav)
	defer rl.UnloadWaveSamples(tmp)
	in.Samples = make([]float32, len(tmp))
	copy(in.Samples, tmp)
	in.SamplesLoaded = true
}

type Track struct {
	Id             uint8
	Instrument     *Instrument `json:"-"`
	InstrumentId   uint8
	CurrentProgram int
	CurrentPhrase  int       `json:"-"`
	Phrases        []*Phrase `json:"-"`
	PhraseIds      []int32
	Volume         float64
	SkipsLeft      uint8
	Skips          uint8
}

func NewTrack(p *Project) *Track {
	result := Track{}
	result.Id = uint8(len(p.Tracks))
	result.Volume = 1.0
	newPhrase := NewPhrase(p)
	newInstrument := NewInstrument(p)
	newInstrument.SampleSourceType = SampleSourceTypeFm
	newInstrument.Program = 0
	result.Phrases = []*Phrase{newPhrase}
	result.PhraseIds = []int32{newPhrase.ID}
	result.InstrumentId = newInstrument.Id
	result.Instrument = newInstrument
	p.Tracks = append(p.Tracks, result)
	return &p.Tracks[len(p.Tracks)-1]
}

type Project struct {
	CurrentTrack int `json:"-"`
	Tracks       []Track
	Filename     string
	Phrases      []Phrase
	Instruments  []Instrument
}

func (p *Phrase) Current() *Step {
	return &p.Steps[p.CurrentStep]
}

func (t *Track) Current() *Phrase {
	if len(t.Phrases) == 0 {
		return nil
	}
	return t.Phrases[t.CurrentPhrase]
}

func (t *Track) Cleanup() {
	//	t.Phrases = nil
	//	fmt.Printf("Cleaning up track %v\n", t)
	//
	//	if t.Sample.Sound.FrameCount != 0 {
	//		rl.UnloadSound(t.Sample.Sound)
	//	}
	//
	//	for sampleId := range t.Samples {
	//		if t.Samples[sampleId].Sound.FrameCount != 0 {
	//			rl.UnloadSound(t.Samples[sampleId].Sound)
	//		}
	//	}
}

func (p *Project) Current() *Track {
	return &p.Tracks[p.CurrentTrack]
}

var CurrentProject *Project

func ResetHead() {
	for trackId := range CurrentProject.Tracks {
		CurrentProject.Tracks[trackId].CurrentPhrase = 0
		for phraseId := range CurrentProject.Tracks[trackId].Phrases {
			CurrentProject.Tracks[trackId].Phrases[phraseId].CurrentStep = 0
		}
	}
}

func Clamp[T constraints.Ordered](value T, min T, max T) T {
	if value < min {
		return min
	} else if value > max {
		return max
	}
	return value
}

func ParseNote(annotatedNote string) Note {
	fmt.Printf("Parsing %s.\n", annotatedNote)

	switch annotatedNote {
	case "--":
		return 0
	case "SK":
		return NoteSkip
	case "OFF":
		return NoteOff
	}
	toneNotation := annotatedNote[0:2]
	octave, _ := strconv.Atoi(annotatedNote[2:3])
	for value, notation := range Notation {
		if notation == toneNotation {
			return Note((value + 1) + (12 * octave))
		}
	}
	return 0
}

func (n Note) ToString() string {
	switch n {
	case NoteNone:
		return "--"
	case NoteCut:
		return "CUT"
	case NoteOff:
		return "OFF"
	case NoteSkip:
		return "SKP"
	default:
		N := n - 11
		octave := int8((N - 1) / 12)
		if octave >= 21 {
			octave = 0
		} else if octave >= 20 {
			octave = -1
		}
		return fmt.Sprintf("%s%d", Notation[(N-1)%SemitonesInOctave], octave)
	}
}

func LoadProject(path string, p *Project) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	if err := json.NewDecoder(file).Decode(p); err != nil {
		return err
	}
	for t := range p.Tracks {
		p.Tracks[t].Phrases = make([]*Phrase, len(p.Tracks[t].PhraseIds))
		for idx, id := range CurrentProject.Tracks[t].PhraseIds {
			fmt.Printf("Appending phrase.\n")
			if id >= int32(len(CurrentProject.Phrases)) {
				id = 0
				rl.TraceLog(rl.LogWarning, "Couldn't match phrase id %d, Project.Phrases is smaller\n", id)
				continue
			}
			p.Tracks[t].Phrases[idx] = &CurrentProject.Phrases[id]
		}
		p.Tracks[t].Instrument = &p.Instruments[p.Tracks[t].InstrumentId]
	}
	return nil
}

func SaveProject() error {
	path := CurrentProject.Filename
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	err = encoder.Encode(CurrentProject)
	if err != nil {
		return err
	}
	return nil
}

var Notation = [SemitonesInOctave]string{"C ", "C#", "D ", "D#", "E ", "F ", "F#", "G ", "G#", "A ", "A#", "B "}

var throttledLogsNum = 0
var previousLog string = ""

func Logf(format string, args ...any) {
	result := fmt.Sprintf(format, args...)
	if result == previousLog {
		throttledLogsNum++
		return
	}

	if throttledLogsNum > 0 {
		fmt.Printf("[%dx]\n%s\n", throttledLogsNum, previousLog)
		throttledLogsNum = 0
	}
	fmt.Print(result)
	previousLog = result
}

// Static compile-time assertion:
// If our RGBA function stops returning a valid raylib.Color, this won't compile.
var _ rl.Color = RGBA(0, 0, 0, 0)

func RGBA(r, g, b, a uint8) rl.Color {
	return rl.Color{R: r, G: g, B: b, A: a}
}
