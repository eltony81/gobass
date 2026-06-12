package sequencer

import (
	"context"
	"fmt"
	"sort"
	"time"

	"gobass/parser"
)

// MIDIWriter is an interface for sending MIDI messages.
// This allows mocking the MIDI output in tests.
type MIDIWriter interface {
	SendNoteOn(channel uint8, pitch uint8, velocity uint8) error
	SendNoteOff(channel uint8, pitch uint8) error
}

// Sequencer schedules and plays MIDI events.
type Sequencer struct {
	events []parser.Event
	writer MIDIWriter
}

// NewSequencer creates a new Sequencer with sorted events.
func NewSequencer(events []parser.Event, writer MIDIWriter) *Sequencer {
	// Sort events by timestamp.
	// If timestamps are equal, we sort NoteOff (value 1) before NoteOn (value 0)
	// to prevent NoteOff from cutting off a NoteOn starting at the exact same millisecond.
	sortedEvents := make([]parser.Event, len(events))
	copy(sortedEvents, events)

	sort.SliceStable(sortedEvents, func(i, j int) bool {
		if sortedEvents[i].Timestamp == sortedEvents[j].Timestamp {
			// NoteOff (1) should come before NoteOn (0)
			return sortedEvents[i].Action == parser.NoteOff && sortedEvents[j].Action == parser.NoteOn
		}
		return sortedEvents[i].Timestamp < sortedEvents[j].Timestamp
	})

	return &Sequencer{
		events: sortedEvents,
		writer: writer,
	}
}

// Play runs the look-ahead event loop. It blocks until all events are played
// or the context is cancelled.
func (s *Sequencer) Play(ctx context.Context) error {
	if len(s.events) == 0 {
		return nil
	}

	fmt.Printf("Starting playback of %d events...\n", len(s.events))
	startTime := time.Now()
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()

	nextEventIdx := 0
	const lookaheadWindowMs = 25 // 25ms look-ahead window

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			elapsed := time.Since(startTime).Milliseconds()

			// Look-ahead: schedule any events scheduled within the lookahead window
			lookaheadLimit := elapsed + lookaheadWindowMs

			for nextEventIdx < len(s.events) {
				ev := s.events[nextEventIdx]
				if ev.Timestamp > lookaheadLimit {
					break // This event and subsequent ones are too far in the future
				}

				// Calculate exact delay from right now to when the event should trigger
				delayMs := ev.Timestamp - time.Since(startTime).Milliseconds()

				if delayMs <= 0 {
					// Event is already due or overdue, run it immediately
					s.dispatchEvent(ev)
				} else {
					// Schedule event with precise delay using time.AfterFunc
					eventToDispatch := ev
					time.AfterFunc(time.Duration(delayMs)*time.Millisecond, func() {
						s.dispatchEvent(eventToDispatch)
					})
				}

				nextEventIdx++
			}

			// If we've processed all events, we can stop the loop once the last event has had time to execute.
			if nextEventIdx >= len(s.events) {
				// Wait for the final NoteOff to finish playing
				lastEvent := s.events[len(s.events)-1]
				finalWait := lastEvent.Timestamp - elapsed
				if finalWait < 0 {
					finalWait = 0
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(time.Duration(finalWait+50) * time.Millisecond):
					fmt.Println("Playback finished successfully.")
					return nil
				}
			}
		}
	}
}

// dispatchEvent sends the MIDI message using the MIDIWriter.
func (s *Sequencer) dispatchEvent(ev parser.Event) {
	var err error
	channel := uint8(0) // Default MIDI channel (usually channel 1 in MIDI terms is 0)

	switch ev.Action {
	case parser.NoteOn:
		err = s.writer.SendNoteOn(channel, ev.Pitch, ev.Velocity)
	case parser.NoteOff:
		err = s.writer.SendNoteOff(channel, ev.Pitch)
	}

	if err != nil {
		fmt.Printf("[Error] Failed to send MIDI event %+v: %v\n", ev, err)
	}
}
