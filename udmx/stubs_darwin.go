//go:build darwin

package udmx

import "log"

type UDmxDevice struct{}

func (d *UDmxDevice) Open() {
	log.Fatal("uDMX display not supported on macOS; use --display=dummy or --visualization=osc")
}
func (d *UDmxDevice) Close()                                {}
func (d *UDmxDevice) SetChannelColor(channel, value uint16) {}
func (d *UDmxDevice) SetMultiple(values []byte)             {}
func (d *UDmxDevice) Render()                               {}

type FTDIDMXDevice struct{}

func (d *FTDIDMXDevice) Open() {
	log.Fatal("FTDI DMX display not supported on macOS; use --display=dummy or --visualization=osc")
}
func (d *FTDIDMXDevice) Close()                                {}
func (d *FTDIDMXDevice) SetChannelColor(channel, value uint16) {}
func (d *FTDIDMXDevice) SetMultiple(values []byte)             {}
func (d *FTDIDMXDevice) Render()                               {}

type SerialDMXDevice struct{}

func (d *SerialDMXDevice) Open() {
	log.Fatal("Serial DMX display not supported on macOS; use --display=dummy or --visualization=osc")
}
func (d *SerialDMXDevice) Close()                                {}
func (d *SerialDMXDevice) SetChannelColor(channel, value uint16) {}
func (d *SerialDMXDevice) SetMultiple(values []byte)             {}
func (d *SerialDMXDevice) Render()                               {}
