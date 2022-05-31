package udmx

import (
	"C"

	"log"

	"github.com/google/gousb"
)

type UDmxDevice struct {
	device *gousb.Device
	ctx    *gousb.Context
}

func (d *UDmxDevice) Open() {
	ctx := gousb.NewContext()

	// vid, pid := gousb.ID(0x16c0), gousb.ID(0x05dc)
	vid, pid := gousb.ID(0x0403), gousb.ID(0x6001)
	device, err := ctx.OpenDeviceWithVIDPID(vid, pid)
	if err != nil {
		log.Fatalf("OpenDeviceWithVIDPID(): %v", err)
	}

	log.Println("Opened device")
	d.device = device
	d.ctx = ctx
}

func (d *UDmxDevice) Close() {
	d.device.Close()
	d.ctx.Close()
}

func (d *UDmxDevice) SetChannelColor(channel uint16, value uint16) {
	log.Printf("Setting channel %d to %d", channel, value)

	d.device.Control(gousb.ControlVendor|gousb.ControlDevice|gousb.ControlOut,
		uint8(SET_SINGLE_CHANNEL), value, channel-1, nil)
}

func (d *UDmxDevice) SetMultiple(values []byte) {
	d.device.Control(gousb.ControlVendor|gousb.ControlDevice|gousb.ControlOut,
		uint8(SET_CHANNEL_RANGE), uint16(len(values)), 1, values)
}

func (d *UDmxDevice) Render() {

}
