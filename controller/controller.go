package controller

import (
	"encoding/csv"

	"github.com/everadaptive/mindlights/eeg/neurosky"
	"github.com/everadaptive/mindlights/handler"
	"github.com/lucasb-eyer/go-colorful"
	"go.uber.org/zap"
)

type Controller struct {
	handler   handler.EEGHandler
	events    chan neurosky.MindflexEvent
	csvWriter *csv.Writer
	log       *zap.SugaredLogger
}

func NewController(handler handler.EEGHandler, events chan neurosky.MindflexEvent, csvWriter *csv.Writer, palette []colorful.Color, log *zap.SugaredLogger) Controller {
	return Controller{
		handler:   handler,
		events:    events,
		csvWriter: csvWriter,
		log:       log,
	}
}

func (c *Controller) Start(displayOffset int) {
	if c.csvWriter != nil {
		d2 := neurosky.EEGFullData{}
		if err := c.csvWriter.Write(d2.GetHeaders()); err != nil {
			c.log.Fatal("error writing headers to csv:", err)
		}
		c.csvWriter.Flush()
	}

	c.handler.Start()

	for v := range c.events {
		c.DoWork(v)
	}
}

func (c *Controller) DoWork(v neurosky.MindflexEvent) {
	// c.log.Debugw("packet received", "source", v.Source, "signal", v.SignalQuality, "attention", v.Attention, "meditation", v.Meditation)

	switch v.Type {
	case neurosky.POOR_SIGNAL:
		c.handler.PoorSignal(v)
		break
	case neurosky.ATTENTION:
		c.handler.Attention(v)
	case neurosky.MEDITATION:
		c.handler.Meditation(v)
	case neurosky.EEG_POWER:
		break
	}
}
