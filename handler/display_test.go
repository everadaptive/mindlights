package handler

import (
	"time"

	"github.com/everadaptive/mindlights/display"
	"github.com/lucasb-eyer/go-colorful"
	"go.uber.org/zap"
)

type DisplayTestHandler struct {
	poorSignal    bool
	log           *zap.SugaredLogger
	display       display.ColorDisplay
	palette       []colorful.Color
	colors        []colorful.Color
	displayOffset int
	steps         int
}

func NewDisplayTestHandler(display display.ColorDisplay, log *zap.SugaredLogger, palette []colorful.Color, displayOffset int, steps int) DisplayTestHandler {
	return DisplayTestHandler{
		poorSignal:    false,
		log:           log,
		display:       display,
		palette:       palette,
		displayOffset: displayOffset,
		steps:         steps,
	}
}

func (h *DisplayTestHandler) Start() {
	h.colors = make([]colorful.Color, h.display.DisplaySize())
	col, _ := colorful.Hex("000000")
	for i := 0; i < h.display.DisplaySize(); i++ {
		h.colors = h.colors[1:]
		h.colors = append(h.colors, col)
	}

	for i := 0; i <= h.steps; i++ {
		h.log.Infow("step", i, "h.steps", h.steps, "percent", i/h.steps, "colors", h.colors)

		h.colors = h.colors[1:]
		h.colors = append(h.colors, h.palette[i])

		for k := 0; k < h.display.DisplaySize(); k++ {
			h.display.SetColor(k, h.colors[k])
		}

		h.display.Render()

		time.Sleep(50 * time.Millisecond)
	}
}
