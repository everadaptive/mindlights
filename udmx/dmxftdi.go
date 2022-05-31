package udmx

import (
	"C"

	"log"

	"github.com/everadaptive/mindlights/ftdidmx"
)

type FTDIDMXDevice struct {
	device *ftdidmx.DMX
}

func (d *FTDIDMXDevice) Open() {
	dmx, e := ftdidmx.NewDMXConnection()
	if e != nil {
		log.Fatal(e)
	}

	d.device = dmx
}

func (d *FTDIDMXDevice) Close() {
	d.device.Close()
}

func (d *FTDIDMXDevice) SetChannelColor(channel uint16, value uint16) {
	// Set values for channels.
	err := d.device.SetChannel(int(channel), byte(value))
	if err != nil {
		log.Println(err)
	}

	log.Printf("Setting channel %d to %d", int(channel), byte(value))
}

func (d *FTDIDMXDevice) SetMultiple(values []byte) {
	// udmx.device.Control(gousb.ControlVendor|gousb.ControlDevice|gousb.ControlOut,
	// uint8(SET_CHANNEL_RANGE), uint16(len(values)), 1, values)
}

func (d *FTDIDMXDevice) Render() {
	// Send!
	log.Print("Rendering")
	d.device.Render()
}
