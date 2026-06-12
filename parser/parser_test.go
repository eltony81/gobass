package parser

import (
	"strings"
	"testing"
)

func TestStringToMidiNote(t *testing.T) {
	tests := []struct {
		strNum   int
		fret     int
		expected uint8
		wantErr  bool
	}{
		{1, 0, 43, false},
		{1, 5, 48, false},
		{3, 3, 36, false},
		{4, 0, 28, false},
		{5, 0, 23, false}, // String 5 is Low B (base 23)
		{5, 3, 26, false}, // String 5 fret 3 is D
		{6, 0, 0, true},   // String 6 is invalid
	}

	for _, tt := range tests {
		got, err := StringToMidiNote(tt.strNum, tt.fret)
		if (err != nil) != tt.wantErr {
			t.Errorf("StringToMidiNote(%d, %d) error = %v, wantErr %v", tt.strNum, tt.fret, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && got != tt.expected {
			t.Errorf("StringToMidiNote(%d, %d) = %d, want %d", tt.strNum, tt.fret, got, tt.expected)
		}
	}
}

func TestParseTablature(t *testing.T) {
	input := `
# A comment line
3|3|4
p|0|8
x|4|8
`
	bpm := 120.0
	// 120 BPM:
	// Quarter note (division 4) duration: 240000 / (120 * 4) = 500ms
	// Eighth note (division 8) duration: 240000 / (120 * 8) = 250ms

	r := strings.NewReader(input)
	events, err := ParseTablature(r, bpm)
	if err != nil {
		t.Fatalf("ParseTablature failed: %v", err)
	}

	// Should produce 4 events:
	// Note 1 (3|3|4): NoteOn at 0ms, NoteOff at 485ms (500 - 15)
	// Rest (p|0|8): No events, but advances currentMs from 500ms to 750ms
	// Note 2 (x|4|8): NoteOn at 750ms, NoteOff at 790ms (750 + Min(250*0.5, 40))
	if len(events) != 4 {
		t.Fatalf("Expected 4 events, got %d", len(events))
	}

	// Check note 1
	if events[0].Action != NoteOn || events[0].Pitch != 36 || events[0].Timestamp != 0 {
		t.Errorf("Event 0 mismatch: %+v", events[0])
	}
	if events[1].Action != NoteOff || events[1].Pitch != 36 || events[1].Timestamp != 485 {
		t.Errorf("Event 1 mismatch: %+v", events[1])
	}

	// Check ghost note starting after the rest (at 750ms)
	if events[2].Action != NoteOn || events[2].Pitch != 28 || events[2].Timestamp != 750 || events[2].Velocity != 40 {
		t.Errorf("Event 2 mismatch: %+v", events[2])
	}
	if events[3].Action != NoteOff || events[3].Pitch != 28 || events[3].Timestamp != 790 {
		t.Errorf("Event 3 mismatch: %+v", events[3])
	}
}
