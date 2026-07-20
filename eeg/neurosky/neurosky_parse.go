package neurosky

func (b *Neurosky) ParseMindflex(data []byte, events chan MindflexEvent) {
	extendedCodeLevel := 0
	doneCodeLevel := true
	e := MindflexEEGPower{}

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
			for k := 1; k <= length; k++ {
				switch code {
				case RESET:
					break
				case POOR_SIGNAL:
					b.log.Debugw("packet parsed", "source", b.name, "signal", int(data[n+k]))
					events <- MindflexEvent{
						Type:          POOR_SIGNAL,
						Source:        b.name,
						SignalQuality: int(data[n+k]),
					}
				case ATTENTION:
					b.log.Debugw("packet parsed", "source", b.name, "attention", int(data[n+k]))
					events <- MindflexEvent{
						Type:      ATTENTION,
						Source:    b.name,
						Attention: int(data[n+k]),
					}
				case MEDITATION:
					b.log.Debugw("packet parsed", "source", b.name, "meditation", int(data[n+k]))
					events <- MindflexEvent{
						Type:       MEDITATION,
						Source:     b.name,
						Meditation: int(data[n+k]),
					}
				case EEG_RAW:
					raw := int(data[n+k])*256 + int(data[n+k+1])
					if raw >= 32768 {
						raw = raw - 65536
					}
					events <- MindflexEvent{
						Type:        EEG_RAW,
						Source:      b.name,
						EEGRawPower: int(raw),
					}
					k = k + 1
				case EEG_POWER:
					switch k {
					case 1:
						e.Delta = parse3ByteInteger(data[k : k+3])
					case 4:
						e.Theta = parse3ByteInteger(data[k : k+3])
					case 7:
						e.Low_Alpha = parse3ByteInteger(data[k : k+3])
					case 10:
						e.High_Alpha = parse3ByteInteger(data[k : k+3])
					case 13:
						e.Low_Beta = parse3ByteInteger(data[k : k+3])
					case 16:
						e.High_Beta = parse3ByteInteger(data[k : k+3])
					case 19:
						e.Low_Gamma = parse3ByteInteger(data[k : k+3])
					case 22:
						e.High_Gamma = parse3ByteInteger(data[k : k+3])
						b.log.Debugw("packet parsed", "source", b.name, "delta", e.Delta, "theta", e.Theta, "lowAlpha", e.Low_Alpha, "highAlpha", e.High_Alpha, "lowBeta", e.Low_Beta, "highBeta", e.High_Beta, "lowGamma", e.Low_Gamma, "highGamma", e.High_Gamma)
						events <- MindflexEvent{
							Type:     EEG_POWER,
							Source:   b.name,
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
		if data[n] == SYNC {
			syncCount++
		}

		if syncCount >= 2 {
			packetStart = n - 2

			if data[n] == SYNC {
				continue
			}

			if data[n] > SYNC {
				syncCount = 0
				continue
			}

			packetLength = int(data[n])
			break
		}
	}

	if packetStart >= 0 && len(data) >= packetStart+packetLength+3 {
		ret := data[packetStart : packetStart+packetLength+3]
		return packetStart + packetLength + 3, ret, nil
	}

	return 0, nil, nil
}

func parse3ByteInteger(data []byte) int {
	return ((int(data[0]) << 16) |
		(((1 << 16) - 1) & (int(data[1]) << 8)) |
		(((1 << 8) - 1) & int(data[2])))
}
