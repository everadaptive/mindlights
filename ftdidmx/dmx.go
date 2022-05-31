// Simple Go package to send DMX messages.
// Copyright (c) 2013 AKUALAB INC. All Rights Reserved.
// www.akualab.com - @akualab - info@akualab.com
//
// CREDITS:
// Ported from pySimpleDMX (https://github.com/c0z3n/pySimpleDMX)
// Written by Michael Dvorkin
//
// GNU General Public License v3.  http://www.gnu.org/licenses/
package ftdidmx

import (
	"fmt"
	"io"
	"log"
	"time"

	"github.com/ziutek/ftdi"
)

const (
	START_VAL       = 0x7E
	END_VAL         = 0xE7
	BAUD            = 250000
	TIMEOUT         = 1
	FRAME_SIZE      = 511
	FRAME_SIZE_LOW  = byte(FRAME_SIZE & 0xFF)
	FRAME_SIZE_HIGH = byte(FRAME_SIZE >> 8 & 0xFF)

	DMX_MAB   = 16
	DMX_BREAK = 110
)

var labels = map[string]byte{
	"GET_WIDGET_PARAMETERS": 3, // unused
	"SET_WIDGET_PARAMETERS": 4, // unused
	"RX_DMX_PACKET":         5, // unused
	"TX_DMX_PACKET":         6,
	"TX_RDM_PACKET_REQUEST": 7, // unused
	"RX_DMX_ON_CHANGE":      8, // unused
}

// A serial DMX connection.
type DMX struct {
	dev            *ftdi.Device
	frame          [FRAME_SIZE]byte
	packet         [FRAME_SIZE + 10]byte
	serial         io.ReadWriteCloser
	redChan        int
	blueChan       int
	greenChan      int
	brightnessChan int
}

// Creates a new DMX connection using a serial device.
func NewDMXConnection() (dmx *DMX, err error) {
	device, err := ftdi.OpenFirst(0x0403, 0x6001, ftdi.ChannelA)
	if err != nil {
		log.Fatal(err)
	}

	err = device.Reset()
	if err != nil {
		log.Fatal(err)
	}

	err = device.SetBaudrate(BAUD)
	if err != nil {
		log.Fatal(err)
	}

	err = device.SetLineProperties(ftdi.DataBits8, ftdi.StopBits2, ftdi.ParityNone)
	if err != nil {
		log.Fatal(err)
	}

	err = device.SetFlowControl(ftdi.FlowCtrlDisable)
	if err != nil {
		log.Fatal(err)
	}

	device.PurgeBuffers()
	device.SetRTS(0)

	dmx = &DMX{}
	dmx.dev = device
	dmx.serial = device

	id, _ := dmx.dev.ChipID()
	log.Printf("0x%08x", id)

	latency, _ := device.LatencyTimer()
	log.Printf("latency: %d", latency)
	log.Printf("Opened device [%04x:%04x].", dmx.dev.EEPROM().VendorId(), dmx.dev.EEPROM().ProductId())
	return
}

// Set channel level in the dmx frame to be rendered
// the next time Render() is called.
func (dmx *DMX) SetChannel(channel int, val byte) error {

	checkChannelID(channel)
	dmx.frame[channel] = val
	return nil
}

// Turn off a specific channel.
func (dmx *DMX) ClearChannel(channel int) error {

	checkChannelID(channel)
	dmx.frame[channel] = 0
	return nil
}

// Turn off all channels.
func (dmx *DMX) ClearAll() {

	for i, _ := range dmx.frame {
		dmx.frame[i] = 0
	}
}

// Send frame to serial device.
func (dmx *DMX) Render() error {
	dmx.dev.Reset()

	dmx.dev.SetLineProperties2(ftdi.DataBits8, ftdi.StopBits2, ftdi.ParityNone, ftdi.BreakOn)

	time.Sleep(DMX_BREAK * time.Nanosecond)

	dmx.dev.SetLineProperties2(ftdi.DataBits8, ftdi.StopBits2, ftdi.ParityNone, ftdi.BreakOff)

	time.Sleep(DMX_MAB * time.Nanosecond)

	fmt.Println(dmx.frame[0:])

	// Write dmx frame.
	dmx.serial.Write(dmx.frame[0:])

	// sleep until next frame

	return nil
}

// Send frame to serial device.
func (dmx *DMX) RenderLoop() error {
	m_frameTimeUs := 230000000 * time.Nanosecond
	for {
		dmx.dev.Reset()
		start := time.Now()

		err := dmx.dev.SetLineProperties2(ftdi.DataBits8, ftdi.StopBits2, ftdi.ParityNone, ftdi.BreakOn)
		if err != nil {
			// if error skip this frame
			goto framesleep
		}

		time.Sleep(DMX_BREAK * time.Nanosecond)

		err = dmx.dev.SetLineProperties2(ftdi.DataBits8, ftdi.StopBits2, ftdi.ParityNone, ftdi.BreakOff)
		if err != nil {
			// if error skip this frame
			goto framesleep
		}

		time.Sleep(DMX_MAB * time.Nanosecond)

		fmt.Println(dmx.frame[0:])

		// Write dmx frame.
		dmx.serial.Write(dmx.frame[0:])

		// sleep until next frame
	framesleep:
		for (time.Since(start) / m_frameTimeUs) > 0 {
			time.Sleep(1000 * time.Nanosecond)
		}
	}

	return nil
}

// Close serial port.
func (dmx *DMX) Close() error {
	return dmx.serial.Close()
}

// Convenience method to map colors and brightness to channels.
func (dmx *DMX) ChannelMap(brightness, red, green, blue int) {

	checkChannelID(brightness)
	checkChannelID(red)
	checkChannelID(green)
	checkChannelID(blue)

	dmx.brightnessChan = brightness
	dmx.redChan = red
	dmx.greenChan = green
	dmx.blueChan = blue
}

// Configures RGB+Brightness channels and renders the color.
// Call ChannelMap to configure the RGB channels before calling
// this method.
func (dmx *DMX) SendRGB(brightness, red, green, blue byte) (e error) {

	dmx.ClearAll()
	e = dmx.SetChannel(dmx.brightnessChan, brightness)
	if e != nil {
		return
	}
	e = dmx.SetChannel(dmx.redChan, red)
	if e != nil {
		return
	}
	e = dmx.SetChannel(dmx.greenChan, green)
	if e != nil {
		return
	}
	e = dmx.SetChannel(dmx.blueChan, blue)
	if e != nil {
		return
	}
	e = dmx.Render()
	if e != nil {
		return
	}
	return
}

func checkChannelID(id int) {
	if (id > 512) || (id < 1) {
		panic(fmt.Sprintf("Invalid channel [%d]", id))
	}
}
