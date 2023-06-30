package handler

import (
	"time"

	"github.com/everadaptive/mindlights/display"
	"github.com/everadaptive/mindlights/eeg/neurosky"
	"github.com/lucasb-eyer/go-colorful"
	"go.uber.org/zap"
)

type MovingHeadHandler struct {
	poorSignal    bool
	log           *zap.SugaredLogger
	display       display.ColorDisplay
	palette       []colorful.Color
	colors        []colorful.Color
	displayOffset int
}

func NewAMovingHeadHandler(display display.ColorDisplay, log *zap.SugaredLogger, palette []colorful.Color, displayOffset int) *MovingHeadHandler {
	return &MovingHeadHandler{
		poorSignal:    false,
		log:           log,
		display:       display,
		palette:       palette,
		displayOffset: displayOffset,
	}
}

func (h *MovingHeadHandler) Stop()                        {}
func (h *MovingHeadHandler) Any(v neurosky.MindflexEvent) {}

func (h *MovingHeadHandler) Start() {
	h.colors = make([]colorful.Color, h.display.DisplaySize())
}

func (h *MovingHeadHandler) PoorSignal(v neurosky.MindflexEvent) {
	if v.SignalQuality > 0 {
		h.poorSignal = true
		h.log.Infow("low signal quality", "signal_quality", v.SignalQuality)
		h.display.SetSingleColor(colorful.Color{
			R: 255,
			G: 0,
			B: 0,
		})

		h.display.Render()

		time.Sleep(100 * time.Millisecond)
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

func (h *MovingHeadHandler) Attention(v neurosky.MindflexEvent) {
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

	// h.display.Render()

	h.display.SetChannel(1, uint16(80-v.Attention))
	h.display.Render()

	// red := 2.5 * float64(v.Attention)
	// h.display.SetSingleColor(colorful.Color{
	// 	R: red,
	// 	G: 0,
	// 	B: 0,
	// })
}

func (h *MovingHeadHandler) Meditation(v neurosky.MindflexEvent) {
	if h.poorSignal {
		return
	}
	if v.Meditation == 0 {
		v.Meditation = 1
	}
	if v.Meditation == 100 {
		v.Meditation = 99
	}

	h.display.SetChannel(3, uint16(v.Meditation))
	h.display.Render()
	// h.colors = append(h.colors, h.palette[v.Meditation])[1:]

	// for i := 0; i < h.display.DisplaySize(); i++ {
	// 	h.display.SetColor(h.displayOffset+i, h.colors[i])
	// }

	// h.display.Render()
	// red := 2.5 * float64(v.Meditation)
	// h.display.SetSingleColor(colorful.Color{
	// 	R: 0,
	// 	G: 0,
	// 	B: red,
	// })
}
