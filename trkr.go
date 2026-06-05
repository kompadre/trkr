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

type Note uint8
type Effect uint8
type Step struct {
	Notes   [MaxNotesInStep]Note
	Effects [MaxEffectsInStep]Effect
}
type Phrase struct {
	CurrentStep int `json:"-"`
	Steps       [MaxStepsInPhrase]Step
}

func (p *Phrase) Clone() Phrase {
	result := Phrase{}
	result = *p
	return result
}

type Sample struct {
	Loaded     bool `json:"-"`
	SampleFile string
	Sound      rl.Sound `json:"-"`
	RootNote   Note
}

type Track struct {
	Id            uint8
	CurrentPhrase int `json:"-"`
	Phrases       []Phrase
	Samples       []Sample
	Sample        Sample
	IsMultisample bool
}

type Project struct {
	CurrentTrack int `json:"-"`
	Tracks       []Track
	Filename     string
}

func (p *Phrase) Current() *Step {
	return &p.Steps[p.CurrentStep]
}

func (t *Track) Current() *Phrase {
	return &t.Phrases[t.CurrentPhrase]
}

func (t *Track) Cleanup() {
	t.Phrases = nil
	fmt.Printf("Cleaning up track %v\n", t)
	for i := range t.Samples {
		if t.Samples[i].Loaded {
			fmt.Printf("Unloading sound %s\n", t.Samples[i].SampleFile)
			rl.UnloadSound(t.Samples[i].Sound)
			t.Samples[i].Loaded = false
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
	toneNotation := annotatedNote[0:2]
	octave, _ := strconv.Atoi(annotatedNote[2:3])
	for value, notation := range Notation {
		if notation == toneNotation {
			return Note((value + 1) + (12 * octave))
		}
	}
	return 0
}

func LoadProject(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	if err := json.NewDecoder(file).Decode(CurrentProject); err != nil {
		return err
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
	encoder.SetIndent("", "  ")
	err = encoder.Encode(CurrentProject)
	if err != nil {
		return err
	}
	return nil
}

var Notation = [SemitonesInOctave]string{"C ", "C#", "D ", "D#", "E ", "F ", "F#", "G ", "G#", "A ", "A#", "B "}
