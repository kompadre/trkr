package player

import (
	"testing"
	"time"
	"trkr"
	"trkr/internal/audio"
)

func TestCreatePlayerAndGetElem(t *testing.T) {
	CreatePlayer(nil)
	elem := GetElem()
	if elem == nil {
		t.Fatalf("expected non-nil ui element from GetElem()")
	}
	if elem.Visible {
		t.Errorf("expected player ui element to be invisible")
	}
}

func TestPlayheadAndRepeatsLeft(t *testing.T) {
	p0 := trkr.Phrase{ID: 0, Repeats: 3}
	p1 := trkr.Phrase{ID: 1, Repeats: 2}

	sec := trkr.Section{Id: 0, Name: "Intro", Rows: 64}

	track := trkr.Track{
		Id:             0,
		CurrentSection: 0,
		PhraseIds:      [trkr.MaxSections][]int32{0: {0, 1}},
		Phrases:        [trkr.MaxSections][]*trkr.Phrase{0: {&p0, &p1}},
	}

	project := &trkr.Project{
		CurrentSection: 0,
		Tracks:         []trkr.Track{track},
		Phrases:        []trkr.Phrase{p0, p1},
		Sections:       []trkr.Section{sec},
	}

	trkr.CurrentProject = project

	playhead := NewPlayhead(0)
	if playhead.CurrentSectionId != 0 {
		t.Errorf("expected CurrentSectionId == 0, got %d", playhead.CurrentSectionId)
	}
	if playhead.Section.Tracks[0].TrackId != 0 {
		t.Errorf("expected TrackId == 0, got %d", playhead.Section.Tracks[0].TrackId)
	}
	if playhead.Section.Tracks[0].CurrentPhraseId != 0 {
		t.Errorf("expected CurrentPhraseId == 0, got %d", playhead.Section.Tracks[0].CurrentPhraseId)
	}

	Head = playhead
	trRuntime := playhead.Section.Tracks[0]

	// Slot equal to PhraseCounter (0), RepeatsCounter = 0 -> min(3, 3 - 0) = 3
	if rem := trRuntime.RepeatsLeft(0); rem != 3 {
		t.Errorf("expected RepeatsLeft(0) == 3, got %d", rem)
	}

	// RepeatsCounter = 1 -> min(3, 3 - 1) = 2
	trRuntime.Phrase.RepeatsCounter = 1
	if rem := trRuntime.RepeatsLeft(0); rem != 2 {
		t.Errorf("expected RepeatsLeft(0) == 2, got %d", rem)
	}

	// Slot != PhraseCounter (1 != 0) -> phrase.Repeats = 2
	if rem := trRuntime.RepeatsLeft(1); rem != 2 {
		t.Errorf("expected RepeatsLeft(1) == 2, got %d", rem)
	}
}

func TestPlayAndStop(t *testing.T) {
	audio.InitVoiceMasterStream()

	p0 := trkr.Phrase{ID: 0, Repeats: 1}
	sec := trkr.Section{Id: 0, Name: "Intro", Rows: 16}
	track := trkr.Track{
		Id:             0,
		CurrentSection: 0,
		PhraseIds:      [trkr.MaxSections][]int32{0: {0}},
		Phrases:        [trkr.MaxSections][]*trkr.Phrase{0: {&p0}},
	}

	trkr.CurrentProject = &trkr.Project{
		CurrentSection: 0,
		Tracks:         []trkr.Track{track},
		Phrases:        []trkr.Phrase{p0},
		Sections:       []trkr.Section{sec},
	}

	go Play()

	// Wait briefly for playback to start
	time.Sleep(50 * time.Millisecond)
	if !IsPlaying {
		t.Errorf("expected IsPlaying to be true after Play()")
	}

	Stop()

	// Wait briefly for playback to stop
	time.Sleep(50 * time.Millisecond)
	if IsPlaying {
		t.Errorf("expected IsPlaying to be false after Stop()")
	}
}
