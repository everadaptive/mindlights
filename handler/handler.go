package handler

import "github.com/everadaptive/mindlights/eeg/neurosky"

type EEGHandler interface {
	Start()
	PoorSignal(v neurosky.MindflexEvent)
	Attention(v neurosky.MindflexEvent)
	Meditation(v neurosky.MindflexEvent)
}
