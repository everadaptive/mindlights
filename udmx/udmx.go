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

const SET_SINGLE_CHANNEL = 1
const SET_CHANNEL_RANGE = 2

func (udmx *UDmxDevice) Open() {
	ctx := gousb.NewContext()

	vid, pid := gousb.ID(0x16c0), gousb.ID(0x05dc)
	device, err := ctx.OpenDeviceWithVIDPID(vid, pid)
	if err != nil {
		log.Fatalf("OpenDeviceWithVIDPID(): %v", err)
	}

	log.Println("Opened device")
	udmx.device = device
	udmx.ctx = ctx
}

func (udmx *UDmxDevice) Close() {
	udmx.device.Close()
	udmx.ctx.Close()
}

func (udmx *UDmxDevice) SetChannelColor(channel uint16, value uint16) {
	log.Printf("Setting channel %d to %d", channel, value)

	udmx.device.Control(gousb.ControlVendor|gousb.ControlDevice|gousb.ControlOut,
		uint8(SET_SINGLE_CHANNEL), value, channel-1, nil)
}

func (udmx *UDmxDevice) SetMultiple(values []byte) {
	udmx.device.Control(gousb.ControlVendor|gousb.ControlDevice|gousb.ControlOut,
		uint8(SET_CHANNEL_RANGE), uint16(len(values)), 1, values)
}
