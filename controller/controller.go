package controller

import (
	"encoding/csv"
	"time"

	"github.com/everadaptive/mindlights/display"
	"github.com/lucasb-eyer/go-colorful"
	"go.uber.org/zap"
)

type Controller struct {
	display   display.ColorDisplay
	events    chan MindflexEvent
	csvWriter *csv.Writer
	palette   []colorful.Color
	colors    []colorful.Color
}

var (
	log *zap.SugaredLogger
)

func init() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync() // flushes buffer, if any
	log = logger.Sugar()
}

func NewController(display display.ColorDisplay, events chan MindflexEvent, csvWriter *csv.Writer, palette []colorful.Color) Controller {
	return Controller{
		display:   display,
		events:    events,
		csvWriter: csvWriter,
		palette:   palette,
	}
}

func (c *Controller) Start() {
	if c.csvWriter != nil {
		d2 := EEGFullData{}
		if err := c.csvWriter.Write(d2.GetHeaders()); err != nil {
			log.Fatal("error writing headers to csv:", err)
		}
		c.csvWriter.Flush()
	}

	c.colors = make([]colorful.Color, c.display.DisplaySize())

	for v := range c.events {
		c.DoWork(v)
	}
}

func (c *Controller) DoWork(v MindflexEvent) {
	log.Debugw("packet received", "source", v.Source, "signal", v.SignalQuality, "attention", v.Attention, "meditation", v.Meditation)

	switch v.Type {
	case POOR_SIGNAL:
		break
	case ATTENTION:
		if v.Attention == 0 {
			v.Attention = 1
		}
		if v.Attention == 100 {
			v.Attention = 99
		}

		// c.colors = append(c.colors, c.palette[v.Attention])[1:]

		// for i := 0; i < c.display.DisplaySize(); i++ {
		// 	c.display.SetColor(i, c.colors[i])
		// }

		// c.display.Render()

		// red := 2.5 * float64(v.Attention)
		// c.display.SetSingleColor(colorful.Color{
		// 	R: red,
		// 	G: 0,
		// 	B: 0,
		// })
	case MEDITATION:
		if v.Meditation == 0 {
			v.Meditation = 1
		}
		if v.Meditation == 100 {
			v.Meditation = 99
		}

		c.colors = append(c.colors, c.palette[v.Meditation])[1:]

		for i := 0; i < c.display.DisplaySize(); i++ {
			c.display.SetColor(i, c.colors[i])
		}

		c.display.Render()
		// red := 2.5 * float64(v.Meditation)
		// c.display.SetSingleColor(colorful.Color{
		// 	R: 0,
		// 	G: 0,
		// 	B: red,
		// })
	case EEG_POWER:
		break
	}
}

func (c *Controller) DisplayTest(total_steps int) {
	colors := make([]colorful.Color, c.display.DisplaySize())
	col, _ := colorful.Hex("000000")
	for i := 0; i < c.display.DisplaySize(); i++ {
		colors = colors[1:]
		colors = append(colors, col)
	}

	for i := 0; i <= total_steps; i++ {
		log.Infow("step", i, "total_steps", total_steps, "percent", i/total_steps, "colors", colors)

		colors = colors[1:]
		colors = append(colors, c.palette[i])

		for k := 0; k < c.display.DisplaySize(); k++ {
			c.display.SetColor(k, colors[k])
		}

		c.display.Render()

		time.Sleep(50 * time.Millisecond)
	}
}
