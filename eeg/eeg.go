package eeg

import "github.com/everadaptive/mindlights/controller"

type EEGHeadset interface {
	Start() chan controller.MindflexEvent
	Stop()
	Close()
}
