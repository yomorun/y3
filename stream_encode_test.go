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
	encoder.AddPacketBuffer([]byte{0x11, 0x02, 0x01, 0x02})

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

func TestStreamEncoderNilReader(t *testing.T) {
	encoder := NewStreamEncoder(0x10)

	//-> 0x11, 0x02, 0x01, 0x02,
	encoder.AddPacketBuffer([]byte{0x11, 0x02, 0x01, 0x02})

	assert.EqualValues(t, 0, encoder.GetLen())

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
	assert.Equal(t, []byte(nil), final.Bytes())
}

func TestStreamEncoderMultiplePackets(t *testing.T) {
	expected := []byte{
		0x10, 0x0C,
		0x11, 0x02, 0x01, 0x02,
		0x13, 0x01, 0x07,
		0x12, 0x03, 0x01, 0x02, 0x03}
	data := []byte{0x01, 0x02, 0x03}

	encoder := NewStreamEncoder(0x10)
	encoder.AddPacketBuffer([]byte{0x11, 0x02, 0x01, 0x02})
	encoder.AddPacketBuffer([]byte{0x13, 0x01, 0x07})
	encoder.AddStreamPacket(0x12, len(data), bytes.NewReader(data))

	assert.EqualValues(t, len(expected), encoder.GetLen())

	n, err := io.ReadAll(encoder.GetReader())
	assert.NoError(t, err)
	assert.Equal(t, expected, n)
}

func TestStreamEncoderSourceLength(t *testing.T) {
	// a source longer than the declared length is cut at the length
	encoder := NewStreamEncoder(0x10)
	encoder.AddStreamPacket(0x12, 3, bytes.NewReader([]byte{0x01, 0x02, 0x03, 0x04, 0x05}))
	n, err := io.ReadAll(encoder.GetReader())
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x10, 0x05, 0x12, 0x03, 0x01, 0x02, 0x03}, n)

	// a source shorter than the declared length is an error
	encoder = NewStreamEncoder(0x10)
	encoder.AddStreamPacket(0x12, 5, bytes.NewReader([]byte{0x01, 0x02}))
	_, err = io.ReadAll(encoder.GetReader())
	assert.Equal(t, io.ErrUnexpectedEOF, err)
}
