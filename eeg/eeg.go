package eeg

import "github.com/everadaptive/mindlights/eeg/neurosky"

type EEGHeadset interface {
	Start() chan neurosky.MindflexEvent
	Stop()
	Close()
}
