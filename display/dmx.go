package display

import (
	"fmt"

	"github.com/everadaptive/mindlights/udmx"
	"github.com/lucasb-eyer/go-colorful"
)

type UDmxDisplay struct {
	Size   int
	device udmx.DmxDevice
	red    int
	green  int
	blue   int
}

func NewUDmxDisplay(size int, device udmx.DmxDevice) UDmxDisplay {
	return UDmxDisplay{
		Size:   size,
		device: device,
		red:    8,
		green:  9,
		blue:   10,
	}
}

func (d UDmxDisplay) SetColor(id int, color colorful.Color) error {
	if id < 0 || id > d.Size {
		return fmt.Errorf("error setting color, index is out of bounds")
	}

	r, g, b := color.RGB255()

	d.device.SetChannelColor(uint16(3*id+d.red), uint16(r))
	d.device.SetChannelColor(uint16(3*id+d.green), uint16(g))
	d.device.SetChannelColor(uint16(3*id+d.blue), uint16(b))

	return nil
}

func (d UDmxDisplay) SetSingleColor(color colorful.Color) error {
	for i := 0; i < d.Size; i++ {
		r, g, b := color.RGB255()

		if r != 0 {
			d.device.SetChannelColor(uint16(3*i+d.red), uint16(r))
		}
		if g != 0 {
			d.device.SetChannelColor(uint16(3*i+d.green), uint16(g))
		}
		if b != 0 {
			d.device.SetChannelColor(uint16(3*i+d.blue), uint16(b))
		}
	}

	return nil
}

func (d UDmxDisplay) DisplaySize() int {
	return d.Size
}

func (d UDmxDisplay) Render() {
	d.device.Render()
}
