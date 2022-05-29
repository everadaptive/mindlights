package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"go.uber.org/zap"
	"golang.org/x/sys/unix"
)

var log *zap.SugaredLogger

func main() {
	address := "98:D3:31:80:7B:3D"

	logger, _ := zap.NewDevelopment()
	defer logger.Sync() // flushes buffer, if any
	log = logger.Sugar()

	mac := str2ba(address) // YOUR BLUETOOTH MAC ADDRESS HERE

	fd, err := unix.Socket(syscall.AF_BLUETOOTH, syscall.SOCK_STREAM, unix.BTPROTO_RFCOMM)
	if err != nil {
		log.Fatal(err)
	}
	defer unix.Close(fd)

	addr := &unix.SockaddrRFCOMM{Addr: mac, Channel: 1}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)

	go func() {
		sig := <-c
		fmt.Printf("Got %s signal. Aborting...\n", sig)
		unix.Close(fd)
	}()

	log.Infow("connecting...", "mac", address)
	err = unix.Connect(fd, addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Infow("connected...", "mac", address)

	btReader := NewBTReader(fd)
	btReader.Write([]byte{0x02})

	scanner := bufio.NewScanner(&btReader)
	scanner.Split(ScanMindflex)

	events := make(chan MindflexEvent)

	go func() {
		defer close(events)

		for {
			select {
			case v, ok := <-events:
				if !ok {
					return
				}

				switch v.Type {
				case POOR_SIGNAL:
					poorSignalHandler(v.SignalQuality)
				case ATTENTION:
					attentionHandler(v.Attention)
				case MEDITATION:
					meditationHandler(v.Meditation)
				case EEG_POWER:
					eegPowerHandler(v.EEGPower)
				}
			}
		}
	}()

	for scanner.Scan() {
		p := scanner.Bytes()
		if len(p) > 7 {
			log.Infow("received packet", "length", len(p), "data", p)
			ParseMindflex(p[3:])
		}
	}
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
