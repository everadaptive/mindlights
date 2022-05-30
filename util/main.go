package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"golang.org/x/sys/unix"
)

func main() {
	mac := str2ba("98:D3:31:80:7B:3D") // YOUR BLUETOOTH MAC ADDRESS HERE

	fd, err := unix.Socket(syscall.AF_BLUETOOTH, syscall.SOCK_STREAM, unix.BTPROTO_RFCOMM)
	if err != nil {
		log.Fatal(err)
	}

	addr := &unix.SockaddrRFCOMM{Addr: mac, Channel: 1}

	// _ = unix.Bind(fd, &unix.SockaddrRFCOMM{
	// 	Channel: 1,
	// 	Addr:    [6]uint8{0, 0, 0, 0, 0, 0}, // BDADDR_ANY or 00:00:00:00:00:00
	// })
	// _ = unix.Listen(fd, 1)
	// nfd, sa, _ := unix.Accept(fd)
	// fmt.Printf("conn addr=%v fd=%d", sa.(*unix.SockaddrRFCOMM).Addr, nfd)

	log.Print("connecting...")
	err = unix.Connect(fd, addr)
	if err != nil {
		log.Fatal(err)
	}

	defer unix.Close(fd)
	log.Println("done")

	unix.Write(fd, []byte{0x02})

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)

	go func() {
		select {
		case sig := <-c:
			fmt.Printf("Got %s signal. Aborting...\n", sig)
			unix.Close(fd)
		}
	}()

	// set headset to raw mode
	// port.Write([]byte{0x00, 0xF8, 0x00, 0x00, 0x00, 0xE0})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		buf := []byte{}
		buffer := bytes.NewBuffer(buf)
		for {
			log.Println("reading data")
			bRead := make([]byte, 20)

			_, err := unix.Read(fd, bRead)
			if err != nil {
				log.Fatal(err)
			}

			buffer.Write(bRead)

			b0 := byte(0)
			b1 := byte(0)
			for buffer.Len() > 2 {
				fullData := FullData{}

				b1, _ = buffer.ReadByte()

				if b1 == byte(0) {
					continue
				}

				if b0 == 0xAA && b1 == 0xAA {
					log.Println("Found packet")

					for buffer.Len() > 0 {
						b, _ := buffer.ReadByte()

						switch b {
						case CodeEx:
						case CodeSignalQuality:
							ret, _ := buffer.ReadByte()
							fullData.Signal = ret
						case CodeAttention:
							ret, _ := buffer.ReadByte()
							fullData.Attention = ret
						case CodeMeditation:
							ret, _ := buffer.ReadByte()
							fullData.Meditation = ret
						case CodeBlink:
							ret, _ := buffer.ReadByte()
							fullData.Blink = ret
						case CodeWave:
							buffer.Next(1)
							var ret = make([]byte, 2)
							buffer.Read(ret)
							fullData.Wave = int16(ret[0])<<8 | int16(ret[1])
						case CodeAsicEEG:
							ret := make([]byte, 25)
							i, _ := buffer.Read(ret)
							if i == 25 {
								eegData := parseEEG(ret)
								fullData.Delta = eegData.Delta
								fullData.Theta = eegData.Theta
								fullData.LoAlpha = eegData.LoAlpha
								fullData.HiAlpha = eegData.HiAlpha
								fullData.LoBeta = eegData.LoBeta
								fullData.HiBeta = eegData.HiBeta
								fullData.LoGamma = eegData.LoGamma
								fullData.MidGamma = eegData.MidGamma
							}

							goto nextPacket
						}
					}

					log.Printf("%+v\n", fullData)
				}

			nextPacket:
				b0 = b1
			}
		}
	}()
	wg.Wait()
}

// parseEEG returns data converted into EEG map
func parseEEG(data []byte) EEGData {
	return EEGData{
		Delta:    parse3ByteInteger(data[0:3]),
		Theta:    parse3ByteInteger(data[3:6]),
		LoAlpha:  parse3ByteInteger(data[6:9]),
		HiAlpha:  parse3ByteInteger(data[9:12]),
		LoBeta:   parse3ByteInteger(data[12:15]),
		HiBeta:   parse3ByteInteger(data[15:18]),
		LoGamma:  parse3ByteInteger(data[18:21]),
		MidGamma: parse3ByteInteger(data[21:25]),
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
