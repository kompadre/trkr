package player

import (
	"testing"
	. "trkr"
	"trkr/internal/audio"
)

func TestSequencerIntegration(t *testing.T) {
	// 0. Drain command queue from previous tests
	for {
		select {
		case <-audio.CommandQueue:
		default:
			goto Drained
		}
	}
Drained:

	// 1. Setup a minimal functional project
	p0 := Phrase{ID: 0, Repeats: 0}
	p0.Steps[0].Notes[0] = Note(60) // C-4
	p0.Steps[1].Notes[0] = Note(62) // D-4

	p1 := Phrase{ID: 1, Repeats: 1} // Repeat once (total 2 plays)
	p1.Steps[0].Notes[0] = Note(64) // E-4

	instr := Instrument{
		Id:               0,
		SampleSourceType: SampleSourceTypeFm,
	}

	track := Track{
		Id:             0,
		InstrumentId:   0,
		Instrument:     &instr,
		CurrentSection: 0,
		PhraseIds:      [MaxSections][]int32{0: {0, 1}},
		Volume:         1.0,
	}

	CurrentProject = &Project{
		CurrentSection: 0,
		Tracks:         []Track{track},
		Phrases:        []Phrase{p0, p1},
		Instruments:    []Instrument{instr},
		Sections:       []Section{{Id: 0, Rows: 64}},
	}

	// Initialize Playhead
	Head = NewPlayhead(0)

	// --- PHASE 0 ---
	// Tick 0: Should trigger Note 60
	Tick(0)
	select {
	case cmd := <-audio.CommandQueue:
		if cmd.Note != 60 {
			t.Errorf("Tick 0: Expected note 60, got %v", cmd.Note)
		}
	default:
		t.Fatal("Tick 0: Expected a command in queue, but got none")
	}

	// Tick 1: Should trigger Note 62
	Tick(1)
	select {
	case cmd := <-audio.CommandQueue:
		if cmd.Note != 62 {
			t.Errorf("Tick 1: Expected note 62, got %v", cmd.Note)
		}
	default:
		t.Fatal("Tick 1: Expected a command in queue, but got none")
	}

	// Exhaust phrase 0 (16 steps)
	for i := 2; i < 16; i++ {
		Tick(int64(i))
	}

	// --- PHASE 1 (First Play) ---
	// Tick 16: Should transition to Phrase 1 and trigger Note 64
	Tick(16)
	select {
	case cmd := <-audio.CommandQueue:
		if cmd.Note != 64 {
			t.Errorf("Tick 16: Expected note 64 from Phrase 1, got %v", cmd.Note)
		}
	default:
		t.Fatal("Tick 16: Expected a command in queue after phrase transition")
	}

	// Exhaust phrase 1 (first play)
	for i := 17; i < 32; i++ {
		Tick(int64(i))
	}

	// --- PHASE 1 (Repeat) ---
	// Tick 32: Should repeat Phrase 1 and trigger Note 64 again
	Tick(32)
	select {
	case cmd := <-audio.CommandQueue:
		if cmd.Note != 64 {
			t.Errorf("Tick 32: Expected note 64 from repeated Phrase 1, got %v", cmd.Note)
		}
	default:
		t.Fatal("Tick 32: Expected a command in queue during repeat")
	}
}

func TestSectionTransition(t *testing.T) {
	// Setup project with 2 sections of 16 rows each
	p0 := Phrase{ID: 0}
	p0.Steps[0].Notes[0] = Note(60)

	p1 := Phrase{ID: 1}
	p1.Steps[0].Notes[0] = Note(72)

	instr := Instrument{Id: 0, SampleSourceType: SampleSourceTypeFm}

	track := Track{
		Id:             0,
		InstrumentId:   0,
		Instrument:     &instr,
		PhraseIds:      [MaxSections][]int32{0: {0}, 1: {1}},
		Volume:         1.0,
	}

	CurrentProject = &Project{
		Tracks:      []Track{track},
		Phrases:     []Phrase{p0, p1},
		Instruments: []Instrument{instr},
		Sections: []Section{
			{Id: 0, Rows: 16},
			{Id: 1, Rows: 16},
		},
	}

	// Initialize Playhead at Section 0
	Head = NewPlayhead(0)

	// Simulate a playback loop as seen in player.go
	var tickCount int64 = 0
	
	// Process Section 0 (16 rows)
	for i := 0; i < 16; i++ {
		Tick(tickCount)

		// Drain command queue so we don't have leftover notes
		select {
		case <-audio.CommandQueue:
		default:
		}

		Head.Section.TotalRowsCounter++
		
		if Head.Section.TotalRowsCounter >= int32(CurrentProject.Sections[Head.CurrentSectionId].Rows) {
			Head.CurrentSectionId++
			if Head.CurrentSectionId >= uint8(len(CurrentProject.Sections)) {
				Head.CurrentSectionId = 0
			}
			Head = NewPlayhead(Head.CurrentSectionId)
			tickCount = -1 // As in player.go
		}
		tickCount++
	}

	if Head.CurrentSectionId != 1 {
		t.Errorf("Expected transition to Section 1, still at %d", Head.CurrentSectionId)
	}

	// The next Tick should trigger Note 72 from Phrase 1 (Section 1)
	Tick(tickCount)
	select {
	case cmd := <-audio.CommandQueue:
		if cmd.Note != 72 {
			t.Errorf("After transition: Expected note 72 from Section 1, got %v", cmd.Note)
		}
	default:
		t.Fatal("Expected command from Section 1 after transition")
	}
}
