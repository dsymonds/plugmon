package main

import (
	"bytes"
	"log"
	"net"
	"strings"
	"time"

	"github.com/dsymonds/plugmon/inplug"
)

func main() {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{})
	if err != nil {
		log.Fatalf("net.ListenUDP: %v", err)
	}
	laddr := conn.LocalAddr().(*net.UDPAddr)
	log.Printf("Listening for UDP responses on port %d", laddr.Port)

	discReq := &inplug.DiscoveryRequest{
		Now:        time.Now(),
		SourceIP:   laddr.IP,
		SourcePort: laddr.Port,
	}
	msg, err := discReq.MarshalBinary()
	if err != nil {
		log.Fatalf("Encoding discovery request: %v", err)
	}

	dst := &net.UDPAddr{
		IP:   net.IPv4(255, 255, 255, 255),
		Port: 80,
	}
	log.Printf("sending %d byte message: %x", len(msg), msg)
	if _, err := conn.WriteToUDP(msg, dst); err != nil {
		log.Fatalf("conn.WriteToUDP: %v", err)
	}

	// Wait for any responses over the next 2s.
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var scratch [1 << 10]byte
	var nresp int
	for {
		nb, raddr, err := conn.ReadFrom(scratch[:])
		if err != nil {
			if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
				break
			}
			log.Fatalf("conn.ReadFrom: %v", err)
		}
		b := scratch[:nb]
		log.Printf("got back %d bytes from %s: %x", nb, raddr, b)

		d := decode(msg, b)
		log.Printf("* %q (MAC %s, IP %s)", d.name, d.mac, d.ip)
		log.Printf("  unknown1: %x", d.unknown1)
		log.Printf("  unknown2: %x", d.unknown2)
		nresp++
	}
	log.Printf("Received %d responses.", nresp)
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

	d.unknown1 = next(4)  // offset=32
	d.unknown2 = next(18) // offset=36

	// Next 4 are the reversed IP of the switch.
	d.ip = rev(next(4)) // offset=54

	// Next 6 are the reversed MAC.
	d.mac = rev(next(6)) // offset=60

	// The remainder is the switch's name,
	// padded with zero bytes.
	d.name = strings.TrimRight(string(b), "\x00")

	return d
}

type data struct {
	unknown1 []byte
	unknown2 []byte
	ip       net.IP
	mac      net.HardwareAddr
	name     string
}
