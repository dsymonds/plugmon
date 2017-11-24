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

	// The time is in UTC.
	now := dr.Now.UTC()
	y, mon, d := now.Date()
	h, min, s := now.Clock()

	// Two bytes for a year, in little endian.
	data = append(data, byte(y&0xFF), byte(y>>8))

	// Three bytes: seconds, minutes, hour
	data = append(data, byte(s), byte(min), byte(h))

	// I don't know what this byte is for. I've seen it be 4 or 5.
	data = append(data, 4)

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
	// 7B, 7C, 8E, 91 work too.
	data = append(data, 0x8C)

	// The remainder is more static bytes.
	data = append(data, 0xC1, 0, 0, 0, 0, 0x06, 0, 0, 0, 0, 0, 0, 0, 0, 0)

	if len(data) != 48 {
		return nil, fmt.Errorf("inplug internal error: len(data)=%d", len(data))
	}
	return data, nil
}
