package y3

import (
	"bytes"
	"io"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
)

func TestStreamParser(t *testing.T) {
	data := []byte{
		0x11, 0x03, 0x01, 0x02, 0x03,
		0x12, 0x02, 0x01, 0x02}
	s := new(bytes.Buffer)
	s.Write(data)

	var i int
	for {
		sp, err := StreamReadPacket(s)
		if err != nil {
			assert.Error(t, io.EOF, err)
			break
		}
		switch i {
		case 0:
			assert.NoError(t, err)
			assert.EqualValues(t, 0x11, sp.Tag)
			assert.Equal(t, 3, sp.Len)
			all, err := io.ReadAll(sp.Val)
			assert.NoError(t, err)
			assert.Equal(t, data[2:5], all)
		case 1:
			assert.NoError(t, err)
			assert.EqualValues(t, 0x12, sp.Tag)
			assert.Equal(t, 2, sp.Len)
			all, err := io.ReadAll(sp.Val)
			assert.NoError(t, err)
			assert.Equal(t, data[7:9], all)
		default:
			assert.Error(t, io.EOF, err)
		}
		i++
	}
	assert.EqualValues(t, 2, i)
}

func TestStreamParserShortReads(t *testing.T) {
	data := []byte{
		0x11, 0x03, 0x01, 0x02, 0x03,
		0x12, 0x02, 0x01, 0x02}
	// the source returns one byte per Read, like a slow network stream
	s := iotest.OneByteReader(bytes.NewReader(data))

	sp, err := StreamReadPacket(s)
	assert.NoError(t, err)
	assert.Equal(t, 3, sp.Len)
	all, err := io.ReadAll(sp.Val)
	assert.NoError(t, err)
	assert.Equal(t, data[2:5], all)

	sp, err = StreamReadPacket(s)
	assert.NoError(t, err)
	assert.EqualValues(t, 0x12, sp.Tag)
	all, err = io.ReadAll(sp.Val)
	assert.NoError(t, err)
	assert.Equal(t, data[7:9], all)
}

func TestStreamParserTruncatedVal(t *testing.T) {
	data := []byte{0x11, 0x03, 0x01}
	s := bytes.NewReader(data)

	sp, err := StreamReadPacket(s)
	assert.NoError(t, err)
	assert.Equal(t, 3, sp.Len)
	_, err = io.ReadAll(sp.Val)
	assert.Equal(t, io.ErrUnexpectedEOF, err)
}
