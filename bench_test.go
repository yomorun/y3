package y3

import (
	"bytes"
	"testing"
)

func BenchmarkStreamParser(b *testing.B) {
	for i := 0; i < b.N; i++ {
		data := []byte{0x01, 0x03, 0x01, 0x02, 0x03, 0x04}
		reader := bytes.NewReader(data)

		ReadPacket(reader)
	}
}