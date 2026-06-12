package parser

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

// ActionType represents whether the MIDI event starts or stops a note.
type ActionType int

const (
	NoteOn ActionType = iota
	NoteOff
)

func (a ActionType) String() string {
	if a == NoteOn {
		return "NoteOn"
	}
	return "NoteOff"
}

// Event represents a scheduled MIDI event.
type Event struct {
	Timestamp int64      // Absolute time in milliseconds from start
	Action    ActionType // NoteOn or NoteOff
	Pitch     uint8      // MIDI note number (0-127)
	Velocity  uint8      // MIDI velocity (0-127)
}

// StringToMidiNote maps a string number (1-5) and a fret number to a MIDI pitch.
// String 1 = G (base 43)
// String 2 = D (base 38)
// String 3 = A (base 33)
// String 4 = E (base 28)
// String 5 = B (base 23)
func StringToMidiNote(strNum int, fret int) (uint8, error) {
	var base uint8
	switch strNum {
	case 1:
		base = 43 // G
	case 2:
		base = 38 // D
	case 3:
		base = 33 // A
	case 4:
		base = 28 // E
	case 5:
		base = 23 // Low B / Si
	default:
		return 0, fmt.Errorf("invalid string number: %d (must be 1-5)", strNum)
	}
	return base + uint8(fret), nil
}

// ParseTablature reads a tab file and generates a list of absolute-timed MIDI events.
func ParseTablature(r io.Reader, bpm float64) ([]Event, error) {
	if bpm <= 0 {
		return nil, fmt.Errorf("invalid BPM: %f (must be > 0)", bpm)
	}

	var events []Event
	scanner := bufio.NewScanner(r)
	var currentMs float64 = 0
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) != 3 {
			return nil, fmt.Errorf("line %d: invalid format %q (expected Fret|String|Duration)", lineNum, line)
		}

		fretStr := strings.TrimSpace(parts[0])
		strStr := strings.TrimSpace(parts[1])
		durStr := strings.TrimSpace(parts[2])

		// Parse string number
		strNum, err := strconv.Atoi(strStr)
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid string %q: %w", lineNum, strStr, err)
		}

		// Parse duration division
		durationDiv, err := strconv.Atoi(durStr)
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid duration division %q: %w", lineNum, durStr, err)
		}
		if durationDiv <= 0 {
			return nil, fmt.Errorf("line %d: duration division must be positive: %d", lineNum, durationDiv)
		}

		// Calculate duration in milliseconds: (240000 / BPM) / division
		durationMs := 240000.0 / (bpm * float64(durationDiv))

		// If it's a rest (pausa), just advance the time cursor and skip MIDI event generation
		if fretStr == "p" || fretStr == "P" || fretStr == "r" || fretStr == "R" || fretStr == "-" {
			currentMs += durationMs
			continue
		}

		// Parse fret
		var pitch uint8
		var velocity uint8 = 100
		var noteDurationMs = durationMs

		if fretStr == "x" || fretStr == "X" {
			// Ghost note
			// Ghost notes are muted, we play them at lower velocity and with a shorter duration (e.g. half of the step duration or max 40ms)
			p, err := StringToMidiNote(strNum, 0) // Treat open string as reference pitch for ghost note
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", lineNum, err)
			}
			pitch = p
			velocity = 40
			noteDurationMs = math.Min(durationMs*0.5, 40.0) // very short duration
		} else {
			fret, err := strconv.Atoi(fretStr)
			if err != nil {
				return nil, fmt.Errorf("line %d: invalid fret %q: %w", lineNum, fretStr, err)
			}
			p, err := StringToMidiNote(strNum, fret)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", lineNum, err)
			}
			pitch = p
			// Subtract 15ms from duration to create a natural legato gap and prevent timestamp collisions
			noteDurationMs = math.Max(durationMs-15.0, 20.0)
		}

		// Generate NoteOn
		events = append(events, Event{
			Timestamp: int64(math.Round(currentMs)),
			Action:    NoteOn,
			Pitch:     pitch,
			Velocity:  velocity,
		})

		// Generate NoteOff
		events = append(events, Event{
			Timestamp: int64(math.Round(currentMs + noteDurationMs)),
			Action:    NoteOff,
			Pitch:     pitch,
			Velocity:  0,
		})

		// Advance the playback cursor by the full step duration
		currentMs += durationMs
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading tablature: %w", err)
	}

	return events, nil
}
