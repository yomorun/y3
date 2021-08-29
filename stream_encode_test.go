package y3

import (
	"bufio"
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStreamEncoder(t *testing.T) {
	expected := []byte{
		0x10, 0x0B,
		0x11, 0x02, 0x01, 0x02,
		0x12, 0x05, 0x01, 0x02, 0x03, 0x04, 0x05}
	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	s := new(bytes.Buffer)
	s.Write(data)

	encoder := NewStreamEncoder(0x10)

	//-> 0x11, 0x02, 0x01, 0x02,
	n11 := NewPrimitivePacketEncoder(0x11)
	n11.AddBytes([]byte{0x01, 0x02})
	encoder.AddPacket(n11)

	// -> 0x12, 0x05, 0x01, 0x02, 0x03, 0x04, 0x05
	encoder.AddStreamPacket(0x12, len(data), s)

	assert.EqualValues(t, len(expected), encoder.GetLen())

	n, err := io.ReadAll(encoder.GetReader())
	assert.NoError(t, err)
	assert.Equal(t, expected, n[:encoder.GetLen()])
}

func TestStreamEncoder3BytesBatch(t *testing.T) {
	expected := []byte{
		0x10, 0x0B,
		0x11, 0x02, 0x01, 0x02,
		0x12, 0x05, 0x01, 0x02, 0x03, 0x04, 0x05}
	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	s := new(bytes.Buffer)
	s.Write(data)

	encoder := NewStreamEncoder(0x10)

	//-> 0x11, 0x02, 0x01, 0x02,
	n11 := NewPrimitivePacketEncoder(0x11)
	n11.AddBytes([]byte{0x01, 0x02})
	encoder.AddPacket(n11)

	// -> 0x12, 0x05, 0x01, 0x02, 0x03, 0x04, 0x05
	encoder.AddStreamPacket(0x12, len(data), s)

	assert.EqualValues(t, len(expected), encoder.GetLen())

	final := new(bytes.Buffer)
	buf := make([]byte, 3)
	r := bufio.NewReader(encoder.GetReader())
	for {
		v, err := r.Read(buf)
		t.Logf("-->v=%d, err=%v", v, err)
		if err != nil {
			if err == io.EOF {
				final.Write(buf[:v])
				break
			}
		}
		final.Write(buf[:v])
	}
	assert.Equal(t, expected, final.Bytes())
}
