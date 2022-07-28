package handler

import (
	"time"

	"github.com/everadaptive/mindlights/display"
	"github.com/everadaptive/mindlights/eeg/neurosky"
	"github.com/lucasb-eyer/go-colorful"
	"go.uber.org/zap"
)

type AttentionLightHandler struct {
	poorSignal    bool
	log           *zap.SugaredLogger
	display       display.ColorDisplay
	palette       []colorful.Color
	colors        []colorful.Color
	displayOffset int
}

func NewAttentionLightHandler(display display.ColorDisplay, log *zap.SugaredLogger, palette []colorful.Color, displayOffset int) *AttentionLightHandler {
	return &AttentionLightHandler{
		poorSignal:    false,
		log:           log,
		display:       display,
		palette:       palette,
		displayOffset: displayOffset,
	}
}

func (h *AttentionLightHandler) Start() {
	h.colors = make([]colorful.Color, h.display.DisplaySize())
}

func (h *AttentionLightHandler) PoorSignal(v neurosky.MindflexEvent) {
	if v.SignalQuality > 0 {
		h.poorSignal = true
		h.log.Infow("low signal quality", "signal_quality", v.SignalQuality)
		h.display.SetSingleColor(colorful.Color{
			R: 255,
			G: 0,
			B: 0,
		})

		h.display.Render()

		time.Sleep(400 * time.Millisecond)
		h.display.SetSingleColor(colorful.Color{
			R: 0,
			G: 0,
			B: 0,
		})

		h.display.Render()
	} else {
		h.poorSignal = false
	}
}

func (h *AttentionLightHandler) Attention(v neurosky.MindflexEvent) {
	if h.poorSignal {
		return
	}
	if v.Attention == 0 {
		v.Attention = 1
	}
	if v.Attention == 100 {
		v.Attention = 99
	}

	h.colors = append(h.colors, h.palette[v.Attention])[1:]

	for i := 0; i < h.display.DisplaySize(); i++ {
		h.display.SetColor(i, h.colors[i])
	}

	h.display.Render()
}

func (h *AttentionLightHandler) Meditation(v neurosky.MindflexEvent) {
	if h.poorSignal {
		return
	}
}
