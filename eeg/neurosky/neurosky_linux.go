//go:build linux

package neurosky

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
	"syscall"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sys/unix"
)

type Neurosky struct {
	fd               int
	name             string
	bluetoothAddress string
	scanning         bool
	EventsChan       chan MindflexEvent
	log              *zap.SugaredLogger
}

func NewNeurosky(bluetoothAddress string, name string, log *zap.SugaredLogger) (*Neurosky, error) {
	fd, err := unix.Socket(syscall.AF_BLUETOOTH, syscall.SOCK_STREAM, unix.BTPROTO_RFCOMM)
	if err != nil {
		return nil, err
	}

	mac := str2ba(bluetoothAddress)
	addr := &unix.SockaddrRFCOMM{Addr: mac, Channel: 1}

	log.Infow("connecting to headset", "mac", bluetoothAddress, "name", name)
	err = unix.Connect(fd, addr)
	if err != nil {
		return nil, err
	}
	log.Infow("connected to headset", "mac", bluetoothAddress, "name", name)

	n := Neurosky{
		fd:               fd,
		name:             name,
		bluetoothAddress: bluetoothAddress,
		scanning:         false,
		log:              log,
	}
	events := n.Start()
	t1 := time.NewTimer(5 * time.Second)
	select {
	case timeout := <-t1.C:
		n.log.Info("timed out: ", timeout)
		n.Close()
		return NewNeurosky(bluetoothAddress, name, log)
	case <-events:
		n.log.Info("reading from headset")
	}

	return &n, nil
}

func (b *Neurosky) Read(p []byte) (n int, err error) {
	return unix.Read(b.fd, p)
}

func (b *Neurosky) Write(p []byte) (n int, err error) {
	return unix.Write(b.fd, p)
}

func (b *Neurosky) Start() (events chan MindflexEvent) {
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
	unix.Close(b.fd)
	time.Sleep(1 * time.Second)
}

func str2ba(addr string) [6]byte {
	a := strings.Split(addr, ":")
	var b [6]byte
	for i, tmp := range a {
		u, _ := strconv.ParseUint(tmp, 16, 8)
		b[len(b)-1-i] = byte(u)
	}
	return b
}

func ba2str(addr [6]byte) string {
	return fmt.Sprintf("%2.2X:%2.2X:%2.2X:%2.2X:%2.2X:%2.2X",
		addr[5], addr[4], addr[3], addr[2], addr[1], addr[0])
}
