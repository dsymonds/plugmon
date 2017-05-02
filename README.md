# Plugmon

This is a monitoring exporter for a smart plug I own.

It works with the
[Watts Clever INPLUG](https://www.wattsclever.com.au/products/inplug-wifi-app-controlled-socket),
and exports data for use by [Prometheus](https://prometheus.io/).

## The protocol

Here are some notes about the protocol that I have reverse engineered so far.

(https://github.com/Diagonactic/Ankuoo was a good starting point.)

It operates using a custom protocol via UDP on port 80. My best guess why
they use that port for a non-HTTP protocol is to make it more likely to
get past firewalls.

If you send a broadcast UDP message with 48 bytes that look like this:
```
	// These first 32 bytes will be echoed back.
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x0a, 0x00, 0x00, 0x00, 0xe1, 0x07, 0x05, 0x0c,
	0x01, 0x06, 0x1d, 0x04, 0x00, 0x00, 0x00, 0x00,
	10, 0, 0, 2,  // your local IP
	portLo, portHi,  // your local UDP port, little endian
	0x00, 0x00,

	// dunno what the rest is for
	0x8c, 0xc1, 0x00, 0x00, 0x00, 0x00,
	0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
```

then INPLUGs on the network will respond back. See `ping.go` for some
initial work on decoding that.
