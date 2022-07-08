package main

import (
	"github.com/everadaptive/mindlights/controller"
	"github.com/everadaptive/mindlights/display"
	"github.com/everadaptive/mindlights/udmx"
	"go.uber.org/zap"
)

func main() {
	displaySize := 1
	displayType := "ftdidmx"
	var disp display.ColorDisplay

	// PARCAN Single Lamp
	channels := display.DmxChannels{
		MasterBrightness: 1,
		FirstRGBChannel:  3,
		RGBOrder:         display.RGB,
	}

	// Eurolite LED-144 Bar
	// channels := display.DmxChannels{
	// 	MasterBrightness: 0,
	// 	RGBOrder:         display.RGB,
	// }

	// LED Moving Head 11-Chan
	// channels := display.DmxChannels{
	// 	MasterBrightness: 6,
	// 	FirstRGBChannel:  8,
	// 	RGBOrder:         display.RGB,
	// }

	if displayType == "udmx" {
		dmxDevice := udmx.UDmxDevice{}

		dmxDevice.Open()
		defer dmxDevice.Close()

		disp = display.NewUDmxDisplay(displaySize, &dmxDevice, channels)
	} else if displayType == "serialdmx" {
		dmxDevice := udmx.SerialDMXDevice{}

		dmxDevice.Open()
		defer dmxDevice.Close()

		disp = display.NewUDmxDisplay(displaySize, &dmxDevice, channels)
	} else if displayType == "ftdidmx" {
		dmxDevice := udmx.FTDIDMXDevice{}

		dmxDevice.Open()
		defer dmxDevice.Close()
		disp = display.NewUDmxDisplay(displaySize, &dmxDevice, channels)
	} else if displayType == "dummy" {
		disp = display.NewDummyDisplay()
	}

	palette := controller.CustomPalette6()

	c := controller.NewController(disp, nil, nil, palette, &zap.SugaredLogger{})
	c.DisplayTest(100)
}
