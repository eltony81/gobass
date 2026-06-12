package midi

import (
	"fmt"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv" // Registers the rtmididrv driver
)

// Client handles sending messages to a MIDI output port.
type Client struct {
	out  drivers.Out
	send func(midi.Message) error
}

// NewClient initializes the MIDI driver and connects to the specified output port.
// If portName is empty, it connects to the first available output port.
func NewClient(portName string) (*Client, func(), error) {
	var targetPort drivers.Out
	var err error

	if portName != "" {
		targetPort, err = midi.FindOutPort(portName)
		if err != nil {
			// If not found, list available and return error
			var ports []string
			for _, p := range midi.GetOutPorts() {
				ports = append(ports, p.String())
			}
			return nil, nil, fmt.Errorf("could not find output port matching %q. Available ports: %v", portName, ports)
		}
	} else {
		// Use port 0 as default
		targetPort, err = midi.OutPort(0)
		if err != nil {
			return nil, nil, fmt.Errorf("no MIDI output ports found. Is FluidSynth running?")
		}
	}

	err = targetPort.Open()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open MIDI output port %s: %w", targetPort.String(), err)
	}

	fmt.Printf("Connected to MIDI output port: %s\n", targetPort.String())

	sendFn, err := midi.SendTo(targetPort)
	if err != nil {
		targetPort.Close()
		return nil, nil, fmt.Errorf("failed to initialize sender: %w", err)
	}

	cleanup := func() {}

	return &Client{
		out:  targetPort,
		send: sendFn,
	}, cleanup, nil
}

// SendNoteOn sends a MIDI Note On message on the given channel.
func (c *Client) SendNoteOn(channel uint8, pitch uint8, velocity uint8) error {
	msg := midi.NoteOn(channel, pitch, velocity)
	return c.send(msg)
}

// SendNoteOff sends a MIDI Note Off message on the given channel.
func (c *Client) SendNoteOff(channel uint8, pitch uint8) error {
	msg := midi.NoteOff(channel, pitch)
	return c.send(msg)
}

// SendProgramChange sends a MIDI Program Change message to select an instrument.
func (c *Client) SendProgramChange(channel uint8, program uint8) error {
	msg := midi.ProgramChange(channel, program)
	return c.send(msg)
}
