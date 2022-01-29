package controller

import (
	"encoding/csv"
	"fmt"
	"log"
	"time"

	"github.com/everadaptive/mindlights/display"
	"github.com/jinzhu/copier"
	"github.com/lucasb-eyer/go-colorful"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/neurosky"
)

type Controller struct {
	display      display.ColorDisplay
	neuroDriver  *neurosky.Driver
	neuroAdapter *neurosky.Adaptor
	csvWriter    *csv.Writer
	palette      []colorful.Color
}

func NewController(display display.ColorDisplay, neuroDriver *neurosky.Driver, neuroAdapter *neurosky.Adaptor, csvWriter *csv.Writer, palette []colorful.Color) Controller {
	return Controller{
		display:      display,
		neuroDriver:  neuroDriver,
		neuroAdapter: neuroAdapter,
		csvWriter:    csvWriter,
		palette:      palette,
	}
}

func (c Controller) Start() {
	if c.csvWriter != nil {
		d2 := EEGFullData{}
		if err := c.csvWriter.Write(d2.GetHeaders()); err != nil {
			log.Fatalln("error writing headers to csv:", err)
		}
		c.csvWriter.Flush()
	}

	colors := make([]colorful.Color, c.display.DisplaySize())

	robot := gobot.NewRobot("brainBot",
		[]gobot.Connection{c.neuroAdapter},
		[]gobot.Device{c.neuroDriver},
		func() {
			c.DoWork(colors)
		},
	)

	robot.Start()
}

func (c Controller) DoWork(colors []colorful.Color) {
	// c.neuroDriver.On(c.neuroDriver.Event("extended"), func(data interface{}) {
	// 	fmt.Println("Extended", data)
	// })
	c.neuroDriver.On(c.neuroDriver.Event("signal"), func(data interface{}) {
		fmt.Println("Signal", data)
	})
	c.neuroDriver.On(c.neuroDriver.Event("attention"), func(data interface{}) {
		fmt.Println("Attention", data)
		if data == 0 {
			data = 1
		}
		if data == 100 {
			data = 99
		}

		colors = colors[1:]
		colors = append(colors, c.palette[data.(uint8)])

		for i := 0; i < c.display.DisplaySize(); i++ {
			c.display.SetColor(i, colors[i])
		}
	})
	c.neuroDriver.On(c.neuroDriver.Event("meditation"), func(data interface{}) {
		fmt.Println("Meditation", data)
		if data == 0 {
			data = 1
		}
		if data == 100 {
			data = 99
		}
	})
	// c.neuroDriver.On(c.neuroDriver.Event("blink"), func(data interface{}) {
	// 	fmt.Println("Blink", data)
	// })
	// c.neuroDriver.On(c.neuroDriver.Event("wave"), func(data interface{}) {
	// 	fmt.Println("Wave", data)
	// })
	c.neuroDriver.On(c.neuroDriver.Event("eeg"), func(data interface{}) {
	})

	c.neuroDriver.On(c.neuroDriver.Event("all"), func(data interface{}) {
		fullData := EEGFullData{}
		copier.Copy(&fullData, data.(neurosky.FullData))
		fullData.Timestamp = time.Now().UTC().Format(time.RFC3339)
		v := fullData.ToSlice()
		log.Printf("%s", v)
		if c.csvWriter != nil {
			if err := c.csvWriter.Write(v); err != nil {
				log.Fatalln("error writing record to csv:", err)
			}

			// Write any buffered data to the underlying writer (standard output).
			c.csvWriter.Flush()
		}
	})
}

func (c Controller) DisplayTest() {
	colors := make([]colorful.Color, c.display.DisplaySize())

	col, _ := colorful.Hex("000000")
	for i := 0; i < c.display.DisplaySize(); i++ {
		c.display.DisplaySize()
		colors = append(colors, col)
	}

	for i := 0; i <= 100; i++ {
		fmt.Println("step", i)

		colors = colors[1:]
		colors = append(colors, c.palette[i])

		for k := 0; k < c.display.DisplaySize(); k++ {
			c.display.SetColor(k, colors[k])
		}

		time.Sleep(50 * time.Millisecond)
	}
}
