package display

import (
	"fmt"

	"github.com/everadaptive/mindlights/udmx"
	"github.com/lucasb-eyer/go-colorful"
)

type DmxDisplay struct {
	Size   int
	device *udmx.UDmxDevice
	red    int
	green  int
	blue   int
}

func NewDmxDisplay(size int, device *udmx.UDmxDevice) DmxDisplay {
	return DmxDisplay{
		Size:   size,
		device: device,
		red:    1,
		green:  2,
		blue:   3,
	}
}

func (d DmxDisplay) SetColor(id int, color colorful.Color) error {
	if id < 0 || id > d.Size {
		return fmt.Errorf("error setting color, index is out of bounds")
	}

	r, g, b := color.RGB255()
	d.device.SetChannelColor(uint16(3*id+d.red), uint16(r))
	d.device.SetChannelColor(uint16(3*id+d.green), uint16(g))
	d.device.SetChannelColor(uint16(3*id+d.blue), uint16(b))

	return nil
}

func (d DmxDisplay) SetSingleColor(color colorful.Color) error {
	for i := 0; i < d.Size; i++ {
		r, g, b := color.RGB255()
		d.device.SetChannelColor(uint16(3*i+d.red), uint16(r))
		d.device.SetChannelColor(uint16(3*i+d.green), uint16(g))
		d.device.SetChannelColor(uint16(3*i+d.blue), uint16(b))
	}

	return nil
}

func (d DmxDisplay) DisplaySize() int {
	return d.Size
}
