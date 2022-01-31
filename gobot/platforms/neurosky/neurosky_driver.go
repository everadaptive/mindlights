package neurosky

import (
	"bytes"
	"fmt"

	"gobot.io/x/gobot"
)

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

// Driver is the Gobot Driver for the Mindwave
type Driver struct {
	name       string
	connection gobot.Connection
	gobot.Eventer
}

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

// NewDriver creates a Neurosky Driver
// and adds the following events:
//
//   extended - user's current extended level
//   signal - shows signal strength
//   attention - user's current attention level
//   meditation - user's current meditation level
//   blink - user's current blink level
//   wave - shows wave data
//   eeg - showing eeg data
func NewDriver(a *Adaptor) *Driver {
	n := &Driver{
		name:       "Neurosky",
		connection: a,
		Eventer:    gobot.NewEventer(),
	}

	n.AddEvent(Extended)
	n.AddEvent(Signal)
	n.AddEvent(Attention)
	n.AddEvent(Meditation)
	n.AddEvent(Blink)
	n.AddEvent(Wave)
	n.AddEvent(EEG)
	n.AddEvent(Error)
	n.AddEvent(All)

	return n
}

// Connection returns the Driver's connection
func (n *Driver) Connection() gobot.Connection { return n.connection }

// Name returns the Driver name
func (n *Driver) Name() string { return n.name }

// SetName sets the Driver name
func (n *Driver) SetName(name string) { n.name = name }

// adaptor returns neurosky adaptor
func (n *Driver) adaptor() *Adaptor {
	return n.Connection().(*Adaptor)
}

// Start creates a go routine to listen from serial port
// and parse buffer readings
func (n *Driver) Start() (err error) {
	go func() {
		buffFull := make([]byte, 0)

		for {
			buff := make([]byte, 1024)
			c, err := n.adaptor().sp.Read(buff[:])
			// fmt.Printf("Read bytes: %d len: %d\n", c, len(buffFull))
			if c > 0 {
				// fmt.Printf("c-1 %+v\n", buff[:c-1])
				// fmt.Printf("  c %+v\n", buff[:c])
				// fmt.Printf("c+1c %+v\n", buff[:c+1])
				buffFull = append(buffFull, buff[:c]...)
			}

			if err != nil {
				n.Publish(n.Event("error"), err)
			} else {
				nRead := n.parse(bytes.NewBuffer(buffFull))
				// fmt.Printf("pre %+v\n", buffFull)
				if nRead > 0 {
					buffFull = buffFull[nRead:]
					// fmt.Printf("pos %+v\n", buffFull)
				}
			}

			if len(buffFull) > 64 {
				fmt.Printf("Buffer too full, resetting. len: %d\n", len(buffFull))
				buffFull = make([]byte, 0)
			}
		}
	}()
	return
}

// Halt stops neurosky driver (void)
func (n *Driver) Halt() (err error) { return }

// parse converts bytes buffer into packets until no more data is present
func (n *Driver) parse(buf *bytes.Buffer) int {
	count := 0
	b1 := byte(0)

	for buf.Len() > 2 {
		if count > 100 {
			return 90
		}

		if count == 0 {
			count++
			b1, _ = buf.ReadByte()
		}

		count++
		b2, _ := buf.ReadByte()
		if b1 == BTSync && b2 == BTSync {
			length, _ := buf.ReadByte()
			payload := make([]byte, length)
			nRead, err := buf.Read(payload)

			if err != nil {
				n.Publish(n.Event("error"), err)
			}

			if nRead == int(length) {
				//checksum, _ := buf.ReadByte()
				count = count + len(buf.Next(1))
				n.parsePacket(bytes.NewBuffer(payload))
				return count + int(length)
			}
		} else {
			b1 = b2
		}
	}

	return 0
}

// parsePacket publishes event according to data parsed
func (n *Driver) parsePacket(buf *bytes.Buffer) {
	fullData := FullData{}

	for buf.Len() > 0 {
		b, _ := buf.ReadByte()
		switch b {
		case CodeEx:
			n.Publish(n.Event("extended"), nil)
		case CodeSignalQuality:
			ret, _ := buf.ReadByte()
			n.Publish(n.Event("signal"), ret)
			fullData.Signal = ret
		case CodeAttention:
			ret, _ := buf.ReadByte()
			n.Publish(n.Event("attention"), ret)
			fullData.Attention = ret
		case CodeMeditation:
			ret, _ := buf.ReadByte()
			n.Publish(n.Event("meditation"), ret)
			fullData.Meditation = ret
		case CodeBlink:
			ret, _ := buf.ReadByte()
			n.Publish(n.Event("blink"), ret)
			fullData.Blink = ret
		case CodeWave:
			buf.Next(1)
			var ret = make([]byte, 2)
			buf.Read(ret)
			n.Publish(n.Event("wave"), int16(ret[0])<<8|int16(ret[1]))
			fullData.Wave = int16(ret[0])<<8 | int16(ret[1])
		case CodeAsicEEG:
			ret := make([]byte, 25)
			i, _ := buf.Read(ret)
			if i == 25 {
				eegData := n.parseEEG(ret)
				n.Publish(n.Event("eeg"), eegData)
				fullData.Delta = eegData.Delta
				fullData.Theta = eegData.Theta
				fullData.LoAlpha = eegData.LoAlpha
				fullData.HiAlpha = eegData.HiAlpha
				fullData.LoBeta = eegData.LoBeta
				fullData.HiBeta = eegData.HiBeta
				fullData.LoGamma = eegData.LoGamma
				fullData.MidGamma = eegData.MidGamma
			}
		}
	}

	n.Publish(n.Event("all"), fullData)
}

// parseEEG returns data converted into EEG map
func (n *Driver) parseEEG(data []byte) EEGData {
	return EEGData{
		Delta:    n.parse3ByteInteger(data[0:3]),
		Theta:    n.parse3ByteInteger(data[3:6]),
		LoAlpha:  n.parse3ByteInteger(data[6:9]),
		HiAlpha:  n.parse3ByteInteger(data[9:12]),
		LoBeta:   n.parse3ByteInteger(data[12:15]),
		HiBeta:   n.parse3ByteInteger(data[15:18]),
		LoGamma:  n.parse3ByteInteger(data[18:21]),
		MidGamma: n.parse3ByteInteger(data[21:25]),
	}
}

func (n *Driver) parse3ByteInteger(data []byte) int {
	return ((int(data[0]) << 16) |
		(((1 << 16) - 1) & (int(data[1]) << 8)) |
		(((1 << 8) - 1) & int(data[2])))
}
