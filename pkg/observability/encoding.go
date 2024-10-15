package observability

import (
	"bytes"
	"encoding/binary"
	"math"
)

func encodeMessage(msgs []FluentBitMessage) []byte {
	out := bytes.Buffer{}
	for _, msg := range msgs {
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], math.Float64bits(msg.Date))
		out.Write(buf[:])
		for key, msgData := range msg.Data {
			out.Write([]byte(key))
			out.Write([]byte{0xff})
			out.Write([]byte(msgData))
			out.Write([]byte{0xff})
		}
		out.Write([]byte{0xff})
	}
	return out.Bytes()
}
