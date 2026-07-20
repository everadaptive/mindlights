package handler

import (
	"github.com/everadaptive/mindlights/eeg/neurosky"
	"github.com/hypebeast/go-osc/osc"
	"go.uber.org/zap"
)

type OSCSenderHandler struct {
	client *osc.Client
	log    *zap.SugaredLogger
}

func NewOSCSenderHandler(host string, port int, log *zap.SugaredLogger) *OSCSenderHandler {
	return &OSCSenderHandler{
		client: osc.NewClient(host, port),
		log:    log,
	}
}

func (h *OSCSenderHandler) Start() {}
func (h *OSCSenderHandler) Stop()  {}

func (h *OSCSenderHandler) PoorSignal(v neurosky.MindflexEvent) {
	msg := osc.NewMessage("/mindlights/" + v.Source + "/signal")
	msg.Append(int32(v.SignalQuality))
	h.client.Send(msg)
}

func (h *OSCSenderHandler) Attention(v neurosky.MindflexEvent) {
	msg := osc.NewMessage("/mindlights/" + v.Source + "/attention")
	msg.Append(int32(v.Attention))
	h.client.Send(msg)
}

func (h *OSCSenderHandler) Meditation(v neurosky.MindflexEvent) {
	msg := osc.NewMessage("/mindlights/" + v.Source + "/meditation")
	msg.Append(int32(v.Meditation))
	h.client.Send(msg)
}

func (h *OSCSenderHandler) Any(v neurosky.MindflexEvent) {
	if v.Type != neurosky.EEG_POWER {
		return
	}
	prefix := "/mindlights/" + v.Source + "/eeg/"
	send := func(addr string, val int) {
		msg := osc.NewMessage(addr)
		msg.Append(int32(val))
		h.client.Send(msg)
	}
	send(prefix+"delta", v.EEGPower.Delta)
	send(prefix+"theta", v.EEGPower.Theta)
	send(prefix+"lowAlpha", v.EEGPower.Low_Alpha)
	send(prefix+"highAlpha", v.EEGPower.High_Alpha)
	send(prefix+"lowBeta", v.EEGPower.Low_Beta)
	send(prefix+"highBeta", v.EEGPower.High_Beta)
	send(prefix+"lowGamma", v.EEGPower.Low_Gamma)
	send(prefix+"highGamma", v.EEGPower.High_Gamma)
}
