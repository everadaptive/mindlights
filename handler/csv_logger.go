package handler

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"github.com/everadaptive/mindlights/eeg/neurosky"
	"go.uber.org/zap"
)

type CSVLoggerHandler struct {
	poorSignal  bool
	log         *zap.SugaredLogger
	logFilename string
	logFile     *os.File
	csvWriter   *csv.Writer
}

func NewCSVLoggerHandler(log *zap.SugaredLogger, logFilename string) *CSVLoggerHandler {
	return &CSVLoggerHandler{
		poorSignal:  false,
		log:         log,
		logFilename: logFilename,
	}
}

func (h *CSVLoggerHandler) Stop() {
	h.csvWriter.Flush()
	h.logFile.Close()
}

func (h *CSVLoggerHandler) Start() {
	f, err := os.OpenFile(h.logFilename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	h.logFile = f
	h.csvWriter = csv.NewWriter(f)
	headers := []string{
		"timestamp",
		"source",
		"attention",
		"meditation",
		"signal_quality",
		"raw",
		"low_alpha",
		"high_alpha",
		"low_beta",
		"high_beta",
		"low_gamma",
		"high_gamma",
		"delta",
		"theta",
	}
	if err := h.csvWriter.Write(headers); err != nil {
		//write failed do something
	}
}

func (h *CSVLoggerHandler) Any(v neurosky.MindflexEvent) {
	values := []string{
		fmt.Sprintf("%d", time.Now().UnixMilli()),
		v.Source,
		fmt.Sprintf("%d", v.Attention),
		fmt.Sprintf("%d", v.Meditation),
		fmt.Sprintf("%d", v.SignalQuality),
		fmt.Sprintf("%d", v.EEGRawPower),
		fmt.Sprintf("%d", v.EEGPower.Low_Alpha),
		fmt.Sprintf("%d", v.EEGPower.High_Alpha),
		fmt.Sprintf("%d", v.EEGPower.Low_Beta),
		fmt.Sprintf("%d", v.EEGPower.High_Beta),
		fmt.Sprintf("%d", v.EEGPower.Low_Gamma),
		fmt.Sprintf("%d", v.EEGPower.High_Gamma),
		fmt.Sprintf("%d", v.EEGPower.Delta),
		fmt.Sprintf("%d", v.EEGPower.Theta),
	}
	if err := h.csvWriter.Write(values); err != nil {
		//write failed do something
	}
}

func (h *CSVLoggerHandler) PoorSignal(v neurosky.MindflexEvent) {
}

func (h *CSVLoggerHandler) Meditation(v neurosky.MindflexEvent) {
}

func (h *CSVLoggerHandler) Attention(v neurosky.MindflexEvent) {
}

func (h *CSVLoggerHandler) EEGRaw(v neurosky.MindflexEvent) {

}
