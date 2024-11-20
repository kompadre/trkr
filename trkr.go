package trkr

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"golang.org/x/exp/constraints"
	"strconv"
)

const (
	MaxNotesInStep    = 8
	MaxStepsInPhrase  = 32
	MaxEffectsInStep  = 4
	SemitonesInOctave = 12
	BeatsPerMinute    = 105
)

type Note uint8
type Effect uint8
type Step struct {
	Notes   [MaxNotesInStep]Note
	Effects [MaxEffectsInStep]Effect
}
type Phrase struct {
	CurrentStep int
	Steps       [MaxStepsInPhrase]Step
}

func (p *Phrase) Current() *Step {
	return &p.Steps[p.CurrentStep]
}

type Sample struct {
	SampleFile string
	Sound      rl.Sound
	RootNote   Note
}

type Track struct {
	CurrentPhrase int
	Phrases       []Phrase
	Samples       []Sample
	IsMultisample bool
}

func (t *Track) Current() *Phrase {
	return &t.Phrases[t.CurrentPhrase]
}

type Project struct {
	CurrentTrack int
	Tracks       []Track
}

func (p *Project) Current() *Track {
	return &p.Tracks[p.CurrentTrack]
}

var CurrentProject *Project

func ResetHead() {
	currentTrack := &CurrentProject.Tracks[CurrentProject.CurrentTrack]
	currentTrack.CurrentPhrase = 0
	currentPhrase := &currentTrack.Phrases[currentTrack.CurrentPhrase]
	currentPhrase.CurrentStep = -1
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

var Notation = [SemitonesInOctave]string{"C ", "C#", "D ", "D#", "E ", "F ", "F#", "G ", "G#", "A ", "A#", "B "}
