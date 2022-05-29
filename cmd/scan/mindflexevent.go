package main

type MindflexType byte

const (
	POOR_SIGNAL = 0x02
	ATTENTION   = 0x04
	MEDITATION  = 0x05
	EEG_POWER   = 0x83
)

type MindflexEEGPower struct {
	Delta      int
	Theta      int
	Low_Alpha  int
	High_Alpha int
	Low_Beta   int
	High_Beta  int
	Low_Gamma  int
	High_Gamma int
}

type MindflexEvent struct {
	Type MindflexType

	SignalQuality int
	Attention     int
	Meditation    int
	EEGPower      MindflexEEGPower
}
