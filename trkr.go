package trkr

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"trkr/external/msfa"
	"trkr/internal/audio/fm"
	"trkr/internal/audio/perc"

	rl "github.com/gen2brain/raylib-go/raylib"
	"golang.org/x/exp/constraints"
)

const (
	MaxNotesInStep    = 8
	MaxStepsInPhrase  = 16
	MaxPhrasesInTrack = 32
	MaxEffectsInStep  = 4
	MaxSections       = 8
	MaxTracks         = 8
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
	Notes      [MaxNotesInStep]Note
	Velocities [MaxNotesInStep]uint8
	Effects    [MaxEffectsInStep]Effect
}

type Phrase struct {
	ID            int32
	CurrentStep   int `json:"-"`
	Steps         [MaxStepsInPhrase]Step
	Repeats       uint8
	CurrentRepeat uint8 `json:"-"`
	EffectiveRows uint8 `json:"-"`
}

func NewPhrase(pr *Project) *Phrase {
	pr.Phrases = append(pr.Phrases, Phrase{ID: int32(len(pr.Phrases))})
	return &pr.Phrases[len(pr.Phrases)-1]
}

func NewInstrument(pr *Project) *Instrument {
	id := uint8(len(pr.Instruments))
	ins := Instrument{Id: id}
	// Default Percussion Setup
	ins.Percs[0] = perc.Percussion{Freq: 45, ModAmt: 0.5, AmpDecay: 0.9998, PitchDecay: 0.992, NoiseDecay: 0.999, NoiseMix: 0.0}
	ins.Percs[1] = perc.Percussion{Freq: 160, ModAmt: 1.5, AmpDecay: 0.9994, PitchDecay: 0.985, NoiseDecay: 0.9996, NoiseMix: 0.6}
	ins.Percs[2] = perc.Percussion{Freq: 400.0, ModAmt: 8.0, AmpDecay: 0.998, PitchDecay: 0.0, NoiseDecay: 0.9992, NoiseMix: 1.0}
	ins.Percs[3] = perc.Percussion{Freq: 400, ModRatio: 2.31, ModAmt: 1.2, AmpDecay: 0.999, PitchDecay: 0.995, NoiseDecay: 0.990, NoiseMix: 0.15}

	pr.Instruments = append(pr.Instruments, ins)
	return &pr.Instruments[len(pr.Instruments)-1]
}

func (p *Phrase) Clone() *Phrase {
	result := *p
	result.ID = int32(len(CurrentProject.Phrases))
	CurrentProject.Phrases = append(CurrentProject.Phrases, result)
	Logf("New ID: %d.\n", result.ID)
	return &CurrentProject.Phrases[result.ID]
}

func (p *Phrase) Rows() uint8 {
	if p.EffectiveRows == 0 {
		for i := range p.Steps {
			for _, n := range p.Steps[i].Notes {
				if n == NoteSkip {
					p.EffectiveRows = uint8(i + 1)
					return p.EffectiveRows
				}
			}
		}
		p.EffectiveRows = uint8(len(p.Steps))
	}
	return p.EffectiveRows
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
	SampleSourceTypePerc
	SampleSourceTypeFmPickup
)

const (
	VoiceSampleRate = 48000.0
	VoiceBufferSize = 2048 // 1440 // 512 // 512 // 4096
	VoicePoolSize   = 64
)

func (sst SampleSourceType) UiString() string {
	switch sst {
	case SampleSourceTypeWavefile:
		return "Wavefile"
	case SampleSourceTypeSquare:
		return "Square"
	case SampleSourceTypeCosine:
		return "Cosine"
	case SampleSourceTypePerc:
		return "Percussion"
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
	SampleSource     [3]string
	Samples          [3][]float32 `json:"-"`
	LoopStart        float64
	LoopEnd          float64
	SamplesLoaded    bool `json:"-"`
	Program          uint8
	Percs            [4]perc.Percussion
}

func (in *Instrument) GetPercSamples(buffer []float32, slot int) int {
	return in.Percs[slot%len(in.Percs)].GetSamples(buffer)
}

func (in *Instrument) LoadSamples() {
	Logf("Loading samples for instrument %d.\n", in.Id)
	var wav rl.Wave
	defer rl.UnloadWave(wav)

	for i := range in.SampleSource {
		if in.SampleSource[i] == "" {
			continue
		}
		wav = rl.LoadWave(in.SampleSource[i])
		rl.WaveFormat(&wav, VoiceSampleRate, 32, 1)
		in.Samples[i] = rl.LoadWaveSamples(wav)
	}
	in.SamplesLoaded = true
}

type Track struct {
	Id             uint8
	Instrument     *Instrument `json:"-"`
	InstrumentId   uint8
	CurrentProgram int
	CurrentPhrase  int `json:"-"`
	CurrentSection int `json:"-"`
	PhraseIds      [MaxSections][]int32
	Volume         float64
	SkipsLeft      uint8
	Skips          uint8
}

func NewTrack(p *Project) *Track {
	if len(p.Sections) == 0 {
		p.Sections = []Section{{Id: 0, Name: "INTRO", Rows: 224}, {Id: 1, Name: "DEV 1", Rows: 64}}
	}
	result := Track{}
	result.Id = uint8(len(p.Tracks))
	result.Volume = 1.0
	result.CurrentSection = p.CurrentSection
	newPhrase := NewPhrase(p)
	newInstrument := NewInstrument(p)
	newInstrument.SampleSourceType = SampleSourceTypeFm
	newInstrument.Program = 0
	result.PhraseIds[result.CurrentSection] = []int32{newPhrase.ID}
	result.InstrumentId = newInstrument.Id
	result.Instrument = newInstrument
	p.Tracks = append(p.Tracks, result)
	return &p.Tracks[len(p.Tracks)-1]
}

type Section struct {
	Id   uint8
	Name string
	Rows uint32
}

type Project struct {
	CurrentSection int
	CurrentTrack   int `json:"-"`
	Tracks         []Track
	Filename       string
	Phrases        []Phrase
	Instruments    []Instrument
	Sections       []Section
	FmPatchName    string
	BPM            int
}

func (p *Phrase) Current() *Step {
	return &p.Steps[Clamp(p.CurrentStep, 0, len(p.Steps)-1)]
}

func (t *Track) Current() *Phrase {
	if len(t.PhraseIds[t.CurrentSection]) > 0 {
		idx := Clamp(t.CurrentPhrase, 0, len(t.PhraseIds[t.CurrentSection])-1)
		phraseId := t.PhraseIds[t.CurrentSection][idx]
		if int(phraseId) < len(CurrentProject.Phrases) {
			return &CurrentProject.Phrases[phraseId]
		}
	}
	return nil
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
var IsExporting bool
var audioSynthInstance *msfa.Synth

func SetAudioSynthInstance(s *msfa.Synth) {
	audioSynthInstance = s
}

func ResetHead() {
	Logf("Resetting head!\n")
	for trackId := range CurrentProject.Tracks {
		CurrentProject.Tracks[trackId].CurrentSection = 0
		CurrentProject.Tracks[trackId].CurrentPhrase = 0
	}
	for phraseId := range CurrentProject.Phrases {
		p := &CurrentProject.Phrases[phraseId]
		p.CurrentStep = 0
		p.CurrentRepeat = 0
	}
}

func Clamp[T constraints.Ordered](val, minval, maxval T) T {
	return min(max(val, minval), maxval)
}

func Abs[T constraints.Signed](value T) T {
	return max(value, -value)
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
		Logf("Allocating %d sections.\n", len(p.Sections))
		Logf("PhraseIds: %v.\n", p.Tracks[t].PhraseIds)
		for s := range p.Sections {
			if p.Tracks[t].PhraseIds[s] == nil {
				p.Tracks[t].PhraseIds[s] = make([]int32, 0)
			}
		}
		p.Tracks[t].Instrument = &p.Instruments[p.Tracks[t].InstrumentId]
	}
	for n := range p.Instruments {
		i := &p.Instruments[n]
		if i.SampleSourceType == SampleSourceTypeWavefile && !i.SamplesLoaded {
			i.LoadSamples()
			Logf("Loaded samples %v. Len samples %d. Len samples[0] %d.\n", i.SampleSource, len(i.Samples), len(i.Samples[0]))
		}
	}
	if p.BPM > 0 {
		BeatsPerMinute = p.BPM
	}
	return nil
}

func SaveProject() error {
	path := CurrentProject.Filename
	CurrentProject.BPM = BeatsPerMinute
	
	// Also save the FM bank
	if audioSynthInstance != nil {
		bankData := make([]byte, 4096)
		audioSynthInstance.GetBank(bankData)
		bank := fm.Bank{}
		for i := 0; i < 32; i++ {
			copy(bank.Voices[i][:], bankData[i*128:(i+1)*128])
		}
		syxData := bank.ToSysex()
		bankPath := fm.SanitizeFilename(path) + ".syx"
		os.WriteFile(bankPath, syxData, 0644)
	}

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

// Static compile-time assertion:
// If our RGBA function stops returning a valid raylib.Color, this won't compile.
var _ rl.Color = RGBA(0, 0, 0, 0)

func RGBA(r, g, b, a uint8) rl.Color {
	return rl.Color{R: r, G: g, B: b, A: a}
}

var NoteColor = [12]rl.Color{
	RGBA(255, 64+32, 128+32, 255),
	RGBA(255, 64+64, 128+64, 255),
	RGBA(255, 64+96, 128+96, 255),
	RGBA(64+32, 255, 128, 255),
	RGBA(64+64, 255, 128, 255),
	RGBA(64+96, 255, 128, 255),
	RGBA(64+32, 128+32, 255, 255),
	RGBA(64+64, 128+64, 255, 255),
	RGBA(64+96, 128+96, 255, 255),
	RGBA(128+32, 255, 64+32, 255),
	RGBA(128+64, 255, 64+64, 255),
	RGBA(128+96, 255, 64+96, 255),
}
