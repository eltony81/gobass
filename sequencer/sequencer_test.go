package sequencer

import (
	"context"
	"sync"
	"testing"
	"time"

	"gobass/parser"
)

type MockMIDIWriter struct {
	mu     sync.Mutex
	events []parser.Event
}

func (m *MockMIDIWriter) SendNoteOn(channel uint8, pitch uint8, velocity uint8) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, parser.Event{
		Action:   parser.NoteOn,
		Pitch:    pitch,
		Velocity: velocity,
	})
	return nil
}

func (m *MockMIDIWriter) SendNoteOff(channel uint8, pitch uint8) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, parser.Event{
		Action: parser.NoteOff,
		Pitch:  pitch,
	})
	return nil
}

func (m *MockMIDIWriter) GetEvents() []parser.Event {
	m.mu.Lock()
	defer m.mu.Unlock()
	res := make([]parser.Event, len(m.events))
	copy(res, m.events)
	return res
}

func TestSequencerPlay(t *testing.T) {
	events := []parser.Event{
		{Timestamp: 10, Action: parser.NoteOn, Pitch: 36, Velocity: 100},
		{Timestamp: 60, Action: parser.NoteOff, Pitch: 36, Velocity: 0},
	}

	mockWriter := &MockMIDIWriter{}
	seq := NewSequencer(events, mockWriter)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err := seq.Play(ctx)
	if err != nil {
		t.Fatalf("sequencer failed: %v", err)
	}

	outEvents := mockWriter.GetEvents()
	if len(outEvents) != 2 {
		t.Fatalf("expected 2 dispatched events, got %d", len(outEvents))
	}

	if outEvents[0].Action != parser.NoteOn || outEvents[0].Pitch != 36 {
		t.Errorf("first event should be NoteOn 36, got %+v", outEvents[0])
	}
	if outEvents[1].Action != parser.NoteOff || outEvents[1].Pitch != 36 {
		t.Errorf("second event should be NoteOff 36, got %+v", outEvents[1])
	}
}
