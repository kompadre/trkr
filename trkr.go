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
	MaxStepsInPhrase  = 32
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
	ID          int32
	CurrentStep int `json:"-"`
	Steps       [MaxStepsInPhrase]Step
}

func NewPhrase(pr *Project) *Phrase {
	p := Phrase{ID: int32(len(pr.Phrases))}
	pr.Phrases = append(pr.Phrases, p)
	return &p
}

func (p *Phrase) Clone() *Phrase {
	result := *p
	result.ID = int32(len(CurrentProject.Phrases))
	CurrentProject.Phrases = append(CurrentProject.Phrases, result)
	return &result
}

type Sample struct {
	Loaded     bool `json:"-"`
	SampleFile string
	Sound      rl.Sound `json:"-"`
	RootNote   Note
}

type Track struct {
	ID            uint8
	CurrentPhrase int       `json:"-"`
	Phrases       []*Phrase `json:"-"`
	PhraseIds     []int32
	Samples       []Sample
	Sample        Sample
	IsMultisample bool
}

type Project struct {
	CurrentTrack int `json:"-"`
	Tracks       []Track
	Filename     string
	Phrases      []Phrase
}

func (p *Phrase) Current() *Step {
	return &p.Steps[p.CurrentStep]
}

func (t *Track) Current() *Phrase {
	return t.Phrases[t.CurrentPhrase]
}

func (t *Track) Cleanup() {
	t.Phrases = nil
	fmt.Printf("Cleaning up track %v\n", t)
	if t.Sample.Sound.FrameCount != 0 {
		rl.UnloadSound(t.Sample.Sound)
	}
	for sampleId := range t.Samples {
		if t.Samples[sampleId].Sound.FrameCount != 0 {
			rl.UnloadSound(t.Samples[sampleId].Sound)
		}
	}
}

func (p *Project) Current() *Track {
	return &p.Tracks[p.CurrentTrack]
}

var CurrentProject *Project

func ResetHead() {
	for trackId := range CurrentProject.Tracks {
		CurrentProject.Tracks[trackId].CurrentPhrase = 0
		for phraseId := range CurrentProject.Tracks[trackId].Phrases {
			CurrentProject.Tracks[trackId].Phrases[phraseId].CurrentStep = -1
		}
	}
}

func Clamp[T constraints.Integer](value T, min T, max T) T {
	if value < min {
		return min
	} else if value > max {
		return max
	}
	return value
}

func ParseNote(annotatedNote string) Note {
	fmt.Printf("Parsing %s.\n", annotatedNote)
	if annotatedNote == "--" {
		return 0
	} else if annotatedNote == "SK" {
		return NoteSkip
	} else if annotatedNote == "OFF" {
		return NoteOff
	}
	toneNotation := annotatedNote[0:2]
	octave, _ := strconv.Atoi(annotatedNote[2:3])
	for value, notation := range Notation {
		if notation == toneNotation {
			return Note((value + 1) + (12 * octave) + 21)
		}
	}
	return 0
}

func (n Note) ToString() string {
	switch n {
	case NoteNone:
		return "--"
	case NoteOff:
		return "OFF"
	case NoteSkip:
		return "SKP"
	default:
		return fmt.Sprintf("%s%d", Notation[(n-1)%SemitonesInOctave], (n-1)/12)
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
			CurrentProject.Tracks[t].Phrases[idx] = &CurrentProject.Phrases[id]
		}
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

func Logf(format string, args ...any) {
	fmt.Printf(format, args...)
}
