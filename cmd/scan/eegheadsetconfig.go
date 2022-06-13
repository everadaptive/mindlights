package main

type displayConfig struct {
	Size       int    `json:"size"`
	Brightness int    `json:"brightness"`
	Start      int    `json:"start"`
	Order      string `json:"rgbOrder"`
}

type eegHeadsetConfig struct {
	Name             string        `json:"name"`
	BluetoothAddress string        `json:"bluetoothAddress"`
	Display          displayConfig `json:"display"`
}
