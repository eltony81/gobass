#!/bin/bash

SF2_FILE="/home/tony/Projects/gobass/FluidR3_GM.sf2"

# 1. Download SoundFont if not present
if [ ! -f "$SF2_FILE" ] || [ $(stat -c%s "$SF2_FILE") -lt 1000 ]; then
    echo "Downloading FluidR3_GM.sf2 (approx. 141MB)..."
    curl -L -o "$SF2_FILE" "https://raw.githubusercontent.com/urish/cinto/master/media/FluidR3%20GM.sf2"
else
    echo "SoundFont already downloaded."
fi

# 2. Kill any existing fluidsynth process
echo "Cleaning up any old FluidSynth instances..."
pkill fluidsynth || true
sleep 1

# 3. Start FluidSynth in the background with the downloaded SoundFont
echo "Starting FluidSynth..."
fluidsynth -g 1.5 -a pipewire -m alsa_seq -s -i "$SF2_FILE" < /dev/null > fluidsynth.log 2>&1 &
FLUID_PID=$!

# Give FluidSynth a moment to initialize and expose the MIDI port
sleep 6

# 4. Run the sequencer to play the bassline
echo "Starting sequencer playback..."
./gobass -bpm 100 -file tab.txt -port "Synth"

# 5. Clean up FluidSynth process
echo "Stopping FluidSynth..."
kill $FLUID_PID || true
echo "Finished!"
