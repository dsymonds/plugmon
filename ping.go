package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"strings"
)

func main() {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{})
	if err != nil {
		log.Fatalf("net.ListenUDP: %v", err)
	}
	laddr := conn.LocalAddr().(*net.UDPAddr)
	msg := []byte{
		// These first 32 bytes will be echoed back.
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x0a, 0x00, 0x00, 0x00, 0xe1, 0x07, 0x05, 0x0c,
		0x01, 0x06, 0x1d, 0x04, 0x00, 0x00, 0x00, 0x00,
		laddr.IP[0], laddr.IP[1], laddr.IP[2], laddr.IP[3],
		byte(laddr.Port & 0xff), byte(laddr.Port >> 8),
		0x00, 0x00,

		// dunno what the rest is for
		0x6c, 0xc1, 0x00, 0x00, 0x00, 0x00, // 6c -> 108 (dec)
		0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	dst := &net.UDPAddr{
		IP:   net.IPv4(255, 255, 255, 255),
		Port: 80,
	}
	log.Printf("sending %d byte message: %x", len(msg), msg)
	if _, err := conn.WriteToUDP(msg, dst); err != nil {
		log.Fatalf("conn.WriteToUDP: %v", err)
	}
	var scratch [1 << 10]byte
	n, raddr, err := conn.ReadFrom(scratch[:])
	if err != nil {
		log.Fatalf("conn.ReadFrom: %v", err)
	}
	b := scratch[:n]
	log.Printf("got back %d bytes from %s: %x", n, raddr, b)

	d := decode(msg, b)
	log.Printf("* %q (MAC %s, IP %s, ID %q)", d.name, d.mac, d.ip, d.id)
	log.Printf("  unknown1: %x", d.unknown1)
}

func decode(req, b []byte) *data {
	if len(b) != 128 {
		log.Fatalf("bad response length %d (want 128)", len(b))
	}
	if !bytes.Equal(req[:32], b[:32]) {
		log.Fatalf("first 32 bytes of response isn't an echo\n req[:32] = %x\nresp[:32] = %x", req[:32], b[:32])
	}

	// next returns the next n bytes from b.
	next := func(n int) []byte {
		x := b[:n]
		b = b[n:]
		return x
	}
	rev := func(b []byte) []byte {
		for i, j := 0, len(b)-1; i < len(b)/2; i, j = i+1, j-1 {
			b[i], b[j] = b[j], b[i]
		}
		return b
	}

	d := new(data)

	// Drop the first 32 bytes. They echo the request.
	next(32)

	d.unknown1 = next(20)

	// Next 2 are the device ID.
	switch x := fmt.Sprintf("0x%X", next(2)); x {
	case "0x1627":
		d.id = "Neo Power/INPLUG"
	default:
		d.id = x
	}

	// Next 4 are the reversed IP of the switch.
	d.ip = rev(next(4))

	// Next 6 are the reversed MAC.
	d.mac = rev(next(6))

	// The remainder is the switch's name,
	// padded with zero bytes.
	d.name = strings.TrimRight(string(b), "\x00")

	return d
}

type data struct {
	unknown1 []byte
	id       string
	ip       net.IP
	mac      net.HardwareAddr
	name     string
}
