package trkr

import (
	"os"
	"testing"
)

func TestProjectPersistence(t *testing.T) {
	// 1. Create a project with specific data
	p := &Project{
		Filename: "test_project.json",
		BPM:      123,
	}
	p.Sections = []Section{{Id: 0, Name: "Test", Rows: 32}}
	p.Phrases = []Phrase{{ID: 0}}
	p.Phrases[0].Steps[0].Notes[0] = Note(60)
	
	CurrentProject = p
	BeatsPerMinute = 123

	// 2. Save it
	err := SaveProject()
	if err != nil {
		t.Fatalf("Failed to save project: %v", err)
	}
	defer os.Remove("test_project.json")
	defer os.Remove("test_project.syx")

	// 3. Reset globals and load it back
	loadedProject := &Project{}
	err = LoadProject("test_project.json", loadedProject)
	if err != nil {
		t.Fatalf("Failed to load project: %v", err)
	}

	// 4. Verify data
	if loadedProject.BPM != 123 {
		t.Errorf("Expected BPM 123, got %d", loadedProject.BPM)
	}
	if len(loadedProject.Sections) != 1 || loadedProject.Sections[0].Name != "Test" {
		t.Errorf("Section data mismatch")
	}
	if loadedProject.Phrases[0].Steps[0].Notes[0] != 60 {
		t.Errorf("Note data mismatch")
	}
}
