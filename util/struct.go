package main

// EEGData is the EEG raw data returned from the Mindwave
type EEGData struct {
	Delta    int
	Theta    int
	LoAlpha  int
	HiAlpha  int
	LoBeta   int
	HiBeta   int
	LoGamma  int
	MidGamma int
}

// FullData is all of the data returned from the Mindwave
type FullData struct {
	EEGData
	Signal     byte
	Attention  byte
	Meditation byte
	Blink      byte
	Wave       int16
}

const (
	// BTSync is the sync code
	BTSync byte = 0xAA

	// CodeEx Extended code
	CodeEx byte = 0x55

	// CodeSignalQuality POOR_SIGNAL quality 0-255
	CodeSignalQuality byte = 0x02

	// CodeAttention ATTENTION eSense 0-100
	CodeAttention byte = 0x04

	// CodeMeditation MEDITATION eSense 0-100
	CodeMeditation byte = 0x05

	// CodeBlink BLINK strength 0-255
	CodeBlink byte = 0x16

	// CodeWave RAW wave value: 2-byte big-endian 2s-complement
	CodeWave byte = 0x80

	// CodeAsicEEG ASIC EEG POWER 8 3-byte big-endian integers
	CodeAsicEEG byte = 0x83

	// Extended event
	Extended = "extended"

	// Signal event
	Signal = "signal"

	// Attention event
	Attention = "attention"

	// Meditation event
	Meditation = "meditation"

	// Blink event
	Blink = "blink"

	// Wave event
	Wave = "wave"

	// EEG event
	EEG = "eeg"

	// Error event
	Error = "error"

	// All event
	All = "all"
)
