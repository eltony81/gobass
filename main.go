package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gobass/midi"
	"gobass/parser"
	"gobass/sequencer"
)

func main() {
	// Parse CLI flags
	bpmFlag := flag.Float64("bpm", 100.0, "Beats per minute (BPM)")
	portFlag := flag.String("port", "fluidsynth", "MIDI output port name substring match")
	fileFlag := flag.String("file", "tab.txt", "Path to the tablature text file")
	transposeFlag := flag.Int("transpose", 0, "Semitones to transpose (e.g. +12 or +24 for laptop speakers)")
	programFlag := flag.Int("program", 33, "MIDI program/instrument number (0-127, e.g. 33 for electric bass, 36 for slap bass, 0 for piano)")
	flag.Parse()

	fmt.Printf("--- Go MIDI Bass Sequencer ---\n")
	fmt.Printf("Config: BPM=%.1f, File=%s, Port=%q, Transpose=%d, Program=%d\n", *bpmFlag, *fileFlag, *portFlag, *transposeFlag, *programFlag)

	// Check if the tablature file exists; if not, print a message and exit
	if _, err := os.Stat(*fileFlag); os.IsNotExist(err) {
		fmt.Printf("Error: Tablature file %q not found.\n", *fileFlag)
		fmt.Println("Please create one or use a valid path. Example format per line:")
		fmt.Println("  3|3|4   <- Fret 3, String 3, Quarter note")
		fmt.Println("  x|4|8   <- Ghost note, String 4, Eighth note")
		os.Exit(1)
	}

	// Open the tablature file
	file, err := os.Open(*fileFlag)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Parse the events
	events, err := parser.ParseTablature(file, *bpmFlag)
	if err != nil {
		fmt.Printf("Error parsing tablature: %v\n", err)
		os.Exit(1)
	}

	// Apply transposition
	if *transposeFlag != 0 {
		for i := range events {
			newPitch := int(events[i].Pitch) + *transposeFlag
			if newPitch < 0 {
				newPitch = 0
			} else if newPitch > 127 {
				newPitch = 127
			}
			events[i].Pitch = uint8(newPitch)
		}
	}

	if len(events) == 0 {
		fmt.Println("No events parsed from file. Exiting.")
		os.Exit(0)
	}

	// Connect to FluidSynth MIDI port
	midiClient, cleanup, err := midi.NewClient(*portFlag)
	if err != nil {
		fmt.Printf("Error connecting to MIDI: %v\n", err)
		fmt.Println("\nMake sure FluidSynth is running in background. For example:")
		fmt.Println("  fluidsynth -a alsa -m alsa_seq -s -i /usr/share/soundfonts/FluidR3_GM.sf2")
		os.Exit(1)
	}
	defer cleanup()

	// Give the MIDI driver and target port a brief moment to stabilize
	time.Sleep(800 * time.Millisecond)

	// Set instrument program
	err = midiClient.SendProgramChange(0, uint8(*programFlag))
	if err != nil {
		fmt.Printf("Warning: failed to set instrument: %v\n", err)
	}

	// Handle graceful shutdown via Ctrl+C
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nInterrupt received, shutting down sequencer...")
		cancel()
	}()

	// Initialize and run the sequencer
	seq := sequencer.NewSequencer(events, midiClient)
	if err := seq.Play(ctx); err != nil {
		if ctx.Err() != nil {
			fmt.Println("Playback cancelled by user.")
		} else {
			fmt.Printf("Playback error: %v\n", err)
		}
	}
}
