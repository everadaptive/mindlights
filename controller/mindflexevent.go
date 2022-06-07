package controller

type MindflexType byte

const (
	SYNC        = 0xAA
	POOR_SIGNAL = 0x02
	ATTENTION   = 0x04
	MEDITATION  = 0x05
	EEG_RAW     = 0x80
	EEG_POWER   = 0x83

	// RESET means the headset needs to be reset to receive faster data
	RESET = 0x82
)

type MindflexEEGPower struct {
	Delta      int `json:"delta"`
	Theta      int `json:"theta"`
	Low_Alpha  int `json:"lowAlpha"`
	High_Alpha int `json:"highAlpha"`
	Low_Beta   int `json:"lowBeta"`
	High_Beta  int `json:"highBeta"`
	Low_Gamma  int `json:"lowGamma"`
	High_Gamma int `json:"highGamma"`
}

type MindflexEvent struct {
	Type MindflexType `json:"type"`

	Source string `json:"source"`

	SignalQuality  int              `json:"signalQuality"`
	Attention      int              `json:"attention"`
	Meditation     int              `json:"meditation"`
	EEGPower       MindflexEEGPower `json:"eegPower"`
	EEGRawPower    int              `json:"eegRawPower"`
	EEGRawPowerFFT []ComplexValue   `json:"eegRawPowerFft"`
}

type ComplexValue struct {
	Real      float64 `json:"real"`
	Imaginary float64 `json:"imaginary"`
}

type MindflexEventHandlerFunc func(v MindflexEvent)
