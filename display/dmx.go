package display

import (
	"log"

	"github.com/everadaptive/mindlights/udmx"
	"github.com/lucasb-eyer/go-colorful"
)

type RGBOrder int

const (
	RGB  RGBOrder = iota
	RGBW RGBOrder = iota
)

type DmxChannels struct {
	MasterBrightness int
	FirstRGBChannel  int
	RGBOrder         RGBOrder
}

type UDmxDisplay struct {
	Size     int
	device   udmx.DmxDevice
	Channels DmxChannels

	r int
	g int
	b int
	w int
}

func NewUDmxDisplay(size int, device udmx.DmxDevice, channels DmxChannels) UDmxDisplay {
	log.Printf("creating new display size: %d rgborder: %d", size, channels.RGBOrder)
	d := UDmxDisplay{
		Size:     size,
		device:   device,
		Channels: channels,
	}

	d.r = 0
	d.g = 1
	d.b = 2

	switch channels.RGBOrder {
	case RGBW:
		d.r = 0
		d.g = 1
		d.b = 2
		d.w = 3
	case RGB:
	default:
		log.Print("RGB Order")
		d.r = 0
		d.g = 1
		d.b = 2
		d.w = -1
	}

	log.Println(d)
	return d
}

func (d UDmxDisplay) SetColor(id int, color colorful.Color) error {
	r, g, b := color.RGB255()
	if d.Channels.MasterBrightness > 0 {
		d.device.SetChannelColor(uint16(d.Channels.MasterBrightness), 255)
	}

	// log.Printf("id: %d first: %d r: %d g: %d b: %d, r: %d, g: %d, b: %d", id, d.Channels.FirstRGBChannel, d.r, d.g, d.b, r, g, b)
	d.device.SetChannelColor(uint16(3*id+d.Channels.FirstRGBChannel+d.r), uint16(r))
	d.device.SetChannelColor(uint16(3*id+d.Channels.FirstRGBChannel+d.g), uint16(g))
	d.device.SetChannelColor(uint16(3*id+d.Channels.FirstRGBChannel+d.b), uint16(b))

	return nil
}

func (d UDmxDisplay) SetSingleColor(color colorful.Color) error {
	for i := 0; i < d.Size; i++ {
		r, g, b := color.RGB255()
		if d.Channels.MasterBrightness > 0 {
			d.device.SetChannelColor(uint16(d.Channels.MasterBrightness), 255)
		}

		// if r != 0 {
		d.device.SetChannelColor(uint16(3*i+d.Channels.FirstRGBChannel+d.r), uint16(r))
		// }
		// if g != 0 {
		d.device.SetChannelColor(uint16(3*i+d.Channels.FirstRGBChannel+d.g), uint16(g))
		// }
		// if b != 0 {
		d.device.SetChannelColor(uint16(3*i+d.Channels.FirstRGBChannel+d.b), uint16(b))
		// }
	}

	return nil
}

func (d UDmxDisplay) SetChannel(channel uint16, value uint16) error {
	d.device.SetChannelColor(channel, value)
	return nil
}

func (d UDmxDisplay) DisplaySize() int {
	return d.Size
}

func (d UDmxDisplay) Render() {
	d.device.Render()
}
