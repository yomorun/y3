package y3

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStreamParser1(t *testing.T) {
	data := []byte{0x01, 0x03, 0x01, 0x02, 0x03}
	reader := &pr{buf: data}

	p, err := ReadPacket(reader)
	assert.NoError(t, err)
	assert.Equal(t, data, p)
}

func TestStreamParser2(t *testing.T) {
	data := []byte{0x01, 0x03, 0x01, 0x02, 0x03, 0x04}
	reader := &pr{buf: data}

	p, err := ReadPacket(reader)
	assert.NoError(t, err)
	assert.Equal(t, data[:5], p)
}

func TestStreamParser3(t *testing.T) {
	data := []byte{0x01, 0x03, 0x01, 0x02}
	reader := &pr{buf: data}

	p, err := ReadPacket(reader)
	assert.ErrorIs(t, err, ErrMalformed)
	assert.Equal(t, []byte(nil), p)
}

func TestStreamParser4(t *testing.T) {
	data := []byte{}
	reader := &pr{buf: data}

	p, err := ReadPacket(reader)
	assert.ErrorIs(t, err, ErrMalformed)
	assert.Equal(t, []byte(nil), p)
}

func TestStreamParser5(t *testing.T) {
	data := []byte{0x01}
	reader := &pr{buf: data}

	p, err := ReadPacket(reader)
	assert.ErrorIs(t, err, ErrMalformed)
	assert.Equal(t, []byte(nil), p)
}

type pr struct {
	buf []byte
	off int
}

func (pr *pr) Read(buf []byte) (int, error) {
	if pr.off >= len(pr.buf) {
		return 0, io.EOF
	}

	copy(buf, []byte{pr.buf[pr.off]})
	pr.off++
	return 1, nil
}
