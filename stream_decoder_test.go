package y3

import (
	"bytes"
	"io"
	"testing"

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
