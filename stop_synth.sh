#!/bin/bash

if pgrep fluidsynth > /dev/null; then
    echo "Stopping FluidSynth..."
    pkill fluidsynth
    echo "FluidSynth stopped successfully."
else
    echo "FluidSynth is not running."
fi
