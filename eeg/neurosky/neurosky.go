package neurosky

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
	"syscall"

	"github.com/everadaptive/mindlights/controller"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
)

type Neurosky struct {
	fd               int
	name             string
	bluetoothAddress string
	scanning         bool
}

var (
	log *zap.SugaredLogger
)

func init() {
	c := zap.NewDevelopmentConfig()
	c.Sampling = &zap.SamplingConfig{
		Initial:    1,
		Thereafter: 200,
	}

	logger, _ := c.Build()
	defer logger.Sync() // flushes buffer, if any
	log = logger.Sugar()
}

func NewNeurosky(bluetoothAddress string, name string) (*Neurosky, error) {
	fd, err := unix.Socket(syscall.AF_BLUETOOTH, syscall.SOCK_STREAM, unix.BTPROTO_RFCOMM)
	if err != nil {
		return nil, err
	}

	mac := str2ba(bluetoothAddress) // YOUR BLUETOOTH MAC ADDRESS HERE
	addr := &unix.SockaddrRFCOMM{Addr: mac, Channel: 1}

	log.Infow("connecting to headset", "mac", bluetoothAddress, "name", name)
	err = unix.Connect(fd, addr)
	if err != nil {
		return nil, err
	}
	log.Infow("connected to headset", "mac", bluetoothAddress, "name", name)

	return &Neurosky{
		fd:               fd,
		name:             name,
		bluetoothAddress: bluetoothAddress,
		scanning:         false,
	}, nil
}

func (b *Neurosky) Read(p []byte) (n int, err error) {
	return unix.Read(b.fd, p)
}

func (b *Neurosky) Write(p []byte) (n int, err error) {
	return unix.Write(b.fd, p)
}

func (b *Neurosky) Start() (events chan controller.MindflexEvent) {
	b.Write([]byte{0x02})

	events = make(chan controller.MindflexEvent, 10)

	b.scanning = true
	go func() {
		scanner := bufio.NewScanner(b)
		scanner.Split(ScanMindflex)

		for b.scanning && scanner.Scan() {
			p := scanner.Bytes()
			b.ParseMindflex(p[3:], events)
		}
		b.scanning = false
	}()

	return events
}

func (b *Neurosky) Stop() {
	b.scanning = false
}

func (b *Neurosky) Close() {
	b.Stop()
	unix.Close(b.fd)
}

func (b *Neurosky) ParseMindflex(data []byte, events chan controller.MindflexEvent) {
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

		log.Infow("received packet", "mac", b.bluetoothAddress, "name", b.name, "length", len(data), "type", fmt.Sprintf("0x%02x", code))

		if length > 0 {
			// log.Printf("EXCODE level: %d, CODE: 0x%02X, length: %d", extendedCodeLevel, code, length)
			// log.Printf("Data values:")
			for k := 1; k <= length; k++ {
				// log.Printf(" 0x%02X", data[n+k]&0xFF)
				switch code {
				case controller.RESET:
					// b.Stop()
					// time.Sleep(500 * time.Millisecond)
					// b.Start()
					break
				case controller.POOR_SIGNAL:
					log.Debugw("packet parsed", "source", b.name, "signal", int(data[n+k]))
					events <- controller.MindflexEvent{
						Type:          controller.POOR_SIGNAL,
						Source:        b.name,
						SignalQuality: int(data[n+k]),
					}
				case controller.ATTENTION:
					log.Debugw("packet parsed", "source", b.name, "attention", int(data[n+k]))
					events <- controller.MindflexEvent{
						Type:      controller.ATTENTION,
						Source:    b.name,
						Attention: int(data[n+k]),
					}
				case controller.MEDITATION:
					log.Debugw("packet parsed", "source", b.name, "meditation", int(data[n+k]))
					events <- controller.MindflexEvent{
						Type:       controller.MEDITATION,
						Source:     b.name,
						Meditation: int(data[n+k]),
					}
				case controller.EEG_RAW:
					raw := int(data[n+k])*256 + int(data[n+k+1])
					if raw >= 32768 {
						raw = raw - 65536
					}
					// log.Debugw("packet parsed", "source", b.name, "raw", raw)
					events <- controller.MindflexEvent{
						Type:        controller.EEG_RAW,
						Source:      b.name,
						EEGRawPower: int(raw),
					}
					k = k + 1
				case controller.EEG_POWER:
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
						log.Debugw("packet parsed", "source", b.name, "delta", e.Delta, "theta", e.Theta, "lowAlpha", e.Low_Alpha, "highAlpha", e.High_Alpha, "lowBeta", e.Low_Beta, "highBeta", e.High_Beta, "lowGamma", e.Low_Gamma, "highGamma", e.High_Gamma)
						events <- controller.MindflexEvent{
							Type:     controller.EEG_POWER,
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

func parse3ByteInteger(data []byte) int {
	return ((int(data[0]) << 16) |
		(((1 << 16) - 1) & (int(data[1]) << 8)) |
		(((1 << 8) - 1) & int(data[2])))
}

// str2ba converts MAC address string representation to little-endian byte array
func str2ba(addr string) [6]byte {
	a := strings.Split(addr, ":")
	var b [6]byte
	for i, tmp := range a {
		u, _ := strconv.ParseUint(tmp, 16, 8)
		b[len(b)-1-i] = byte(u)
	}
	return b
}

// ba2str converts MAC address little-endian byte array to string representation
func ba2str(addr [6]byte) string {
	return fmt.Sprintf("%2.2X:%2.2X:%2.2X:%2.2X:%2.2X:%2.2X",
		addr[5], addr[4], addr[3], addr[2], addr[1], addr[0])
}