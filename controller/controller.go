package controller

import (
	"encoding/csv"
	"fmt"
	"log"
	"time"

	"github.com/everadaptive/mindlights/display"
	"github.com/lucasb-eyer/go-colorful"
)

type Controller struct {
	display   display.ColorDisplay
	events    chan MindflexEvent
	csvWriter *csv.Writer
	palette   []colorful.Color
}

func NewController(display display.ColorDisplay, events chan MindflexEvent, csvWriter *csv.Writer, palette []colorful.Color) Controller {
	return Controller{
		display:   display,
		events:    events,
		csvWriter: csvWriter,
		palette:   palette,
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

	go func() {
		defer close(c.events)

		for {
			select {
			case v, ok := <-c.events:
				if !ok {
					return
				}

				c.DoWork(v, colors)
			}
		}
	}()
}

func (c Controller) DoWork(v MindflexEvent, colors []colorful.Color) {
	switch v.Type {
	case POOR_SIGNAL:
		fmt.Println("Signal", v.SignalQuality)
	case ATTENTION:
		fmt.Printf("Attention=%d\n", v.Attention)
		if v.Attention == 0 {
			v.Attention = 1
		}
		if v.Attention == 100 {
			v.Attention = 99
		}

		colors = colors[1:]
		colors = append(colors, c.palette[v.Attention])

		for i := 0; i < c.display.DisplaySize(); i++ {
			c.display.SetColor(i, colors[i])
		}

		c.display.Render()

		// red := 2.5 * float64(v.Attention)
		// c.display.SetSingleColor(colorful.Color{
		// 	R: red,
		// 	G: 0,
		// 	B: 0,
		// })
	case MEDITATION:
		fmt.Printf("Meditation=%d\n", v.Meditation)
		if v.Meditation == 0 {
			v.Meditation = 1
		}
		if v.Meditation == 100 {
			v.Meditation = 99
		}

		// red := 2.5 * float64(v.Meditation)
		// c.display.SetSingleColor(colorful.Color{
		// 	R: 0,
		// 	G: 0,
		// 	B: red,
		// })
	case EEG_POWER:
		fmt.Println("Meditation", v.EEGPower)
	}
}

func (c Controller) DisplayTest() {
	colors := make([]colorful.Color, c.display.DisplaySize())
	fmt.Println("step", c.display.DisplaySize())

	col, _ := colorful.Hex("000000")
	for i := 0; i < c.display.DisplaySize(); i++ {
		colors = colors[1:]
		colors = append(colors, col)
		fmt.Println(colors)
	}

	for i := 0; i <= 100; i++ {
		fmt.Println("step", i, colors)

		colors = colors[1:]
		colors = append(colors, c.palette[i])

		for k := 0; k < c.display.DisplaySize(); k++ {
			c.display.SetColor(k, colors[k])
		}

		c.display.Render()

		time.Sleep(5 * time.Millisecond)
	}
}
