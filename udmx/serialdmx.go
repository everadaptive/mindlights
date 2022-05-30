package udmx

import (
	"C"

	"log"

	"github.com/akualab/dmx"
)

type SerialDMXDevice struct {
	device *dmx.DMX
}

const SET_SINGLE_CHANNEL = 1
const SET_CHANNEL_RANGE = 2

func (d *SerialDMXDevice) Open() {
	dmx, e := dmx.NewDMXConnection("/dev/ttyUSB1")
	if e != nil {
		log.Fatal(e)
	}

	d.device = dmx
}

func (d *SerialDMXDevice) Close() {
	d.device.Close()
}

func (d *SerialDMXDevice) SetChannelColor(channel uint16, value uint16) {
	// Set values for channels.
	err := d.device.SetChannel(int(channel), byte(value))
	if err != nil {
		log.Println(err)
	}

	log.Printf("Setting channel %d to %d", int(channel), byte(value))
}

func (d *SerialDMXDevice) SetMultiple(values []byte) {
	// udmx.device.Control(gousb.ControlVendor|gousb.ControlDevice|gousb.ControlOut,
	// uint8(SET_CHANNEL_RANGE), uint16(len(values)), 1, values)
}

func (d *SerialDMXDevice) Render() {
	// Send!
	log.Print("Rendering")
	d.device.Render()
}
