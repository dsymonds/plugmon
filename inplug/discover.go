package inplug

// This file contains the bits for switch discovery via UDP broadcast.

import (
	"fmt"
	"net"
	"time"
)

// DiscoveryRequest represents a broadcast request
// to discover switches.
type DiscoveryRequest struct {
	// Now is the time the request is made.
	Now time.Time
	// SourceIP and SourcePort are the return address,
	// and should usually be the broadcast sender.
	SourceIP   net.IP
	SourcePort int
}

func (dr *DiscoveryRequest) MarshalBinary() (data []byte, err error) {
	// The full request ends up being 48 bytes.
	data = make([]byte, 0, 48)

	// Start with a 12 byte static header.
	data = append(data, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0x0A, 0, 0, 0}...)

	y, mon, d := dr.Now.Date()
	h, min, s := dr.Now.Clock()

	// Two bytes for a year, in little endian.
	data = append(data, byte(y&0xFF), byte(y>>8))

	// Four bytes: seconds, minutes, then ?!?
	// 2017/05/04 18:46:51 -> 08 04
	// 2017/05/04 19:05:06 -> 09 04
	_ = h
	data = append(data, byte(s), byte(min), 8, 4)

	// Two bytes for day and month.
	data = append(data, byte(d), byte(mon))

	// Some more zero bytes.
	data = append(data, []byte{0, 0, 0, 0}...)

	// Source IP and port (little endian).
	ip := dr.SourceIP.To4()
	if ip == nil {
		return nil, fmt.Errorf("source IP (%v) must be IPv4", dr.SourceIP)
	}
	data = append(data, ip...)
	data = append(data, byte(dr.SourcePort&0xFF), byte(dr.SourcePort>>8))

	// A couple more zero bytes.
	data = append(data, []byte{0, 0}...)

	// This is unknown. 6C used to work, but now 8C is required.
	// 7B and 7C work too.
	data = append(data, 0x8C)

	// The remainder is more static bytes.
	data = append(data, 0xC1, 0, 0, 0, 0, 0x06, 0, 0, 0, 0, 0, 0, 0, 0, 0)

	if len(data) != 48 {
		return nil, fmt.Errorf("inplug internal error: len(data)=%d", len(data))
	}
	return data, nil
}
