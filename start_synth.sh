#!/bin/bash

SF2_FILE="/home/tony/Projects/gobass/FluidR3_GM.sf2"

if [ ! -f "$SF2_FILE" ]; then
    echo "Error: SoundFont file not found at $SF2_FILE"
    echo "Please run 'bash run_with_sound.sh' once to download it."
    exit 1
fi

# Check if fluidsynth is already running
if pgrep fluidsynth > /dev/null; then
    echo "FluidSynth is already running."
    exit 0
fi

echo "Starting FluidSynth in background..."
# Start fluidsynth using nohup and redirecting stdin with high gain volume
nohup fluidsynth -g 1.5 -a pipewire -m alsa_seq -s -i "$SF2_FILE" < /dev/null > fluidsynth.log 2>&1 &

# Wait a few seconds for initialization
echo "Waiting 4 seconds for audio/MIDI buffer synchronization..."
sleep 4

echo "FluidSynth started successfully!"
echo "You can now run your sequencer:"
echo "  ./gobass -bpm 100 -file tab.txt -port \"Synth\""
