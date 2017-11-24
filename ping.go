package main

import (
	"bytes"
	"log"
	"net"
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

		var resp inplug.DiscoveryResponse
		if err := resp.UnmarshalBinary(b); err != nil {
			log.Fatal(err)
		}
		if !bytes.Equal(msg[:32], b[:32]) {
			log.Fatalf("first 32 bytes of response isn't an echo\n req[:32] = %x\nresp[:32] = %x", msg[:32], b[:32])
		}

		log.Printf("* %q (MAC %s, IP %s)", resp.Name, resp.MAC, resp.IP)
		log.Printf("  unknown1: %x", resp.Unknown1)
		log.Printf("  unknown2: %x", resp.Unknown2)
		nresp++
	}
	log.Printf("Received %d responses.", nresp)
}
