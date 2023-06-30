package handler

import "github.com/everadaptive/mindlights/eeg/neurosky"

type EEGHandler interface {
	Start()
	Stop()
	PoorSignal(v neurosky.MindflexEvent)
	Attention(v neurosky.MindflexEvent)
	Meditation(v neurosky.MindflexEvent)
	Any(v neurosky.MindflexEvent)
}
