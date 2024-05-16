//go:build linux
// +build linux

package wireguardlinux

import (
	"encoding/binary"
	"fmt"
	"unsafe"
)

// sockaddrPort interprets port as a big endian uint16 for use passing sockaddr
// structures to the kernel.
func sockaddrPort(port int) uint16 {
	return binary.BigEndian.Uint16(Uint16Bytes(uint16(port)))
}

// Uint16Bytes encodes a uint16 into a newly-allocated byte slice using the
// host machine's native endianness.  It is a shortcut for allocating a new
// byte slice and filling it using PutUint16.
func Uint16Bytes(v uint16) []byte {
	b := make([]byte, 2)
	PutUint16(b, v)
	return b
}

// PutUint16 encodes a uint16 into b using the host machine's native endianness.
// If b is not exactly 2 bytes in length, PutUint16 will panic.
func PutUint16(b []byte, v uint16) {
	if l := len(b); l != 2 {
		panic(fmt.Sprintf("PutUint16: unexpected byte slice length: %d", l))
	}

	*(*uint16)(unsafe.Pointer(&b[0])) = v
}
