package main

import (
	"fmt"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

var log *zap.SugaredLogger

func main() {
	rootCmd.Execute()
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
