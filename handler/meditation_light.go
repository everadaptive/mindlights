package handler

import (
	"time"

	"github.com/everadaptive/mindlights/display"
	"github.com/everadaptive/mindlights/eeg/neurosky"
	"github.com/lucasb-eyer/go-colorful"
	"go.uber.org/zap"
)

type MeditationLightHandler struct {
	poorSignal    bool
	log           *zap.SugaredLogger
	display       display.ColorDisplay
	palette       []colorful.Color
	colors        []colorful.Color
	displayOffset int
	displaySize   int
}

func NewMeditationLightHandler(display display.ColorDisplay, log *zap.SugaredLogger, palette []colorful.Color, displayOffset int, displaySize int) *MeditationLightHandler {
	return &MeditationLightHandler{
		poorSignal:    false,
		log:           log,
		display:       display,
		palette:       palette,
		displayOffset: displayOffset,
		displaySize:   displaySize,
	}
}

func (h *MeditationLightHandler) Start() {
	h.colors = make([]colorful.Color, h.displaySize)
}

func (h *MeditationLightHandler) PoorSignal(v neurosky.MindflexEvent) {
	if v.SignalQuality > 0 {
		h.poorSignal = true
		h.log.Infow("low signal quality", "signal_quality", v.SignalQuality)
		for i := 0; i < h.displaySize; i++ {
			h.display.SetColor(h.displayOffset+i, colorful.Color{
				R: 255,
				G: 0,
				B: 0,
			})
		}

		h.display.Render()

		time.Sleep(400 * time.Millisecond)
		for i := 0; i < h.displaySize; i++ {
			h.display.SetColor(h.displayOffset+i, colorful.Color{
				R: 0,
				G: 0,
				B: 0,
			})
		}

		h.display.Render()
	} else {
		h.poorSignal = false
	}
}

func (h *MeditationLightHandler) Attention(v neurosky.MindflexEvent) {
}

func (h *MeditationLightHandler) Meditation(v neurosky.MindflexEvent) {
	if h.poorSignal {
		return
	}

	if v.Meditation == 0 {
		v.Meditation = 1
	}
	if v.Meditation == 100 {
		v.Meditation = 99
	}

	h.colors = append(h.colors, h.palette[v.Meditation])[1:]

	for i := 0; i < h.displaySize; i++ {
		h.display.SetColor(h.displayOffset+i, h.colors[i])
	}

	h.display.Render()
}
