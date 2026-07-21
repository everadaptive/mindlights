//go:build darwin

package neurosky

import (
	"bufio"
	"time"

	"go.bug.st/serial"
	"go.uber.org/zap"
)

// Neurosky connects via a Bluetooth serial port.
// On macOS, pair the headset and use the *call-out* serial device path
// (e.g. /dev/cu.HEADSET02-SerialPort) as the bluetoothAddress in config.
// Use the cu.* node, not tty.*: the tty.* node is the dial-in device and
// blocks on modem carrier-detect, so outbound reads may never see data.
type Neurosky struct {
	port       serial.Port
	name       string
	serialPath string
	scanning   bool
	EventsChan chan MindflexEvent
	log        *zap.SugaredLogger
}

func NewNeurosky(serialPath string, name string, log *zap.SugaredLogger) (*Neurosky, error) {
	mode := &serial.Mode{BaudRate: 57600}
	log.Infow("connecting to headset", "serial", serialPath, "name", name)
	port, err := serial.Open(serialPath, mode)
	if err != nil {
		return nil, err
	}
	log.Infow("connected to headset", "serial", serialPath, "name", name)

	n := &Neurosky{
		port:       port,
		name:       name,
		serialPath: serialPath,
		scanning:   false,
		log:        log,
	}

	events := n.Start()
	t1 := time.NewTimer(5 * time.Second)
	select {
	case timeout := <-t1.C:
		n.log.Info("timed out: ", timeout)
		n.Close()
		return NewNeurosky(serialPath, name, log)
	case <-events:
		n.log.Info("reading from headset")
	}

	return n, nil
}

func (b *Neurosky) Read(p []byte) (n int, err error) {
	return b.port.Read(p)
}

func (b *Neurosky) Write(p []byte) (n int, err error) {
	return b.port.Write(p)
}

func (b *Neurosky) Start() chan MindflexEvent {
	b.Write([]byte{0x02})
	b.EventsChan = make(chan MindflexEvent, 10)
	b.scanning = true
	go func() {
		scanner := bufio.NewScanner(b)
		scanner.Split(ScanMindflex)
		for b.scanning && scanner.Scan() {
			p := scanner.Bytes()
			b.ParseMindflex(p[3:], b.EventsChan)
		}
		b.scanning = false
	}()
	return b.EventsChan
}

func (b *Neurosky) Stop() {
	b.scanning = false
}

func (b *Neurosky) Close() {
	b.Stop()
	b.port.Close()
	time.Sleep(1 * time.Second)
}
