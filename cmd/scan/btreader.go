package main

import (
	"github.com/everadaptive/mindlights/controller"
	"golang.org/x/sys/unix"
)

type BTReader struct {
	fd int
}

func NewBTReader(fd int) BTReader {
	return BTReader{
		fd: fd,
	}
}

func (b *BTReader) Read(p []byte) (n int, err error) {
	return unix.Read(b.fd, p)
}

func (b *BTReader) Write(p []byte) (n int, err error) {
	return unix.Write(b.fd, p)
}

func ParseMindflex(data []byte, events chan controller.MindflexEvent) {
	extendedCodeLevel := 0
	doneCodeLevel := true
	e := controller.MindflexEEGPower{}

	for n := 0; n < len(data); n++ {
		if !doneCodeLevel && data[n] == 0x55 {
			extendedCodeLevel++
		} else {
			doneCodeLevel = true
		}

		code := data[n]
		length := 0
		if code >= 0x80 {
			length = int(data[n+1])
			n = n + 1
		} else {
			length = 1
		}

		if length > 0 {
			// log.Printf("EXCODE level: %d, CODE: 0x%02X, length: %d", extendedCodeLevel, code, length)
			// log.Printf("Data values:")
			for k := 1; k <= length; k++ {
				// log.Printf(" 0x%02X", data[n+k]&0xFF)
				switch code {
				case controller.POOR_SIGNAL:
					log.Debugf("SIGNAL: %d", int(data[n+k]))
					events <- controller.MindflexEvent{
						Type:          controller.POOR_SIGNAL,
						SignalQuality: int(data[n+k]),
					}
				case controller.ATTENTION:
					log.Debugf("ATTENTION: %d", int(data[n+k]))
					events <- controller.MindflexEvent{
						Type:      controller.ATTENTION,
						Attention: int(data[n+k]),
					}
				case controller.MEDITATION:
					log.Debugf("MEDITATION: %d", int(data[n+k]))
					events <- controller.MindflexEvent{
						Type:       controller.MEDITATION,
						Meditation: int(data[n+k]),
					}
				case controller.EEG_POWER:
					switch k {
					case 1:
						log.Debugf("DELTA: %d", parse3ByteInteger(data[k:k+3]))
						e.Delta = parse3ByteInteger(data[k : k+3])
					case 4:
						log.Debugf("THETA: %d", parse3ByteInteger(data[k:k+3]))
						e.Theta = parse3ByteInteger(data[k : k+3])
					case 7:
						log.Debugf("LOW_ALPHA: %d", parse3ByteInteger(data[k:k+3]))
						e.Low_Alpha = parse3ByteInteger(data[k : k+3])
					case 10:
						log.Debugf("HIGH_ALPHA: %d", parse3ByteInteger(data[k:k+3]))
						e.High_Alpha = parse3ByteInteger(data[k : k+3])
					case 13:
						log.Debugf("LOW_BETA: %d", parse3ByteInteger(data[k:k+3]))
						e.Low_Beta = parse3ByteInteger(data[k : k+3])
					case 16:
						log.Debugf("HIGH_BETA: %d", parse3ByteInteger(data[k:k+3]))
						e.High_Beta = parse3ByteInteger(data[k : k+3])
					case 19:
						log.Debugf("LOW_GAMMA: %d", parse3ByteInteger(data[k:k+3]))
						e.Low_Gamma = parse3ByteInteger(data[k : k+3])
					case 22:
						log.Debugf("HIGH_GAMMA: %d", parse3ByteInteger(data[k:k+3]))
						e.High_Gamma = parse3ByteInteger(data[k : k+3])
						events <- controller.MindflexEvent{
							Type:     controller.EEG_POWER,
							EEGPower: e,
						}
					}
					k = k + 2
				}
			}
			n = n + length
		}
	}
}

func ScanMindflex(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF {
		return 0, nil, nil
	}

	if len(data) < 3 {
		return 0, nil, nil
	}

	packetStart := -1
	packetLength := 0
	syncCount := 0
	for n := 0; n < len(data)-1; n++ {
		if data[n] == controller.SYNC {
			syncCount++
		}

		if syncCount >= 2 {
			packetStart = n - 2

			// We might have an additional SYNC
			if data[n] == controller.SYNC {
				continue
			}

			// PLENGTH TO LARGE
			if data[n] > controller.SYNC {
				syncCount = 0
				continue
			}

			packetLength = int(data[n])
			// log.Printf("packet length: %d, sync: %d", packetLength, syncCount)
			break
		}
	}

	// log.Printf("start: %d, packet length: %d, data length: %d", packetStart, packetLength, len(data))
	if packetStart >= 0 && len(data) >= packetStart+packetLength+3 {
		ret := data[packetStart : packetStart+packetLength+3]
		return packetStart + packetLength + 3, ret, nil
	}

	return 0, nil, nil
}