package y3

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLackLengthPrimitivePacket(t *testing.T) {
	buf := []byte{0x01}
	p := &PrimitivePacket{}
	state, err := DecodeToPrimitivePacket(buf, p)
	assert.Error(t, err)
	assert.EqualValues(t, 0, state.EndPos)
	assert.EqualValues(t, 0, state.SizeL)

	p = &PrimitivePacket{}
	state, err = DecodeToPrimitivePacket([]byte{0x01, 0x00}, p)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, state.EndPos)
	assert.EqualValues(t, 1, state.SizeL)
}

// test for { 0x04: -1 }
func TestPacketRead(t *testing.T) {
	buf := []byte{0x04, 0x01, 0x7F}
	expectedTag := byte(0x04)
	expectedValue := []byte{0x7F}
	p := &PrimitivePacket{}

	state, err := DecodeToPrimitivePacket(buf, p)
	assert.NoError(t, err)
	assert.Equal(t, expectedTag, p.tag.SeqID())
	assert.Equal(t, 1, p.length)
	assert.EqualValues(t, expectedValue, p.valbuf)
	assert.Equal(t, 2, state.EndPos)
	assert.EqualValues(t, 1, state.SizeL)
}

// test for { 0x0A: 2 }
func TestParseInt32(t *testing.T) {
	buf := []byte{0x0A, 0x02, 0x81, 0x7F}
	p := &PrimitivePacket{}

	state, err := DecodeToPrimitivePacket(buf, p)
	assert.NoError(t, err)
	assert.EqualValues(t, 3, state.EndPos)
	assert.EqualValues(t, 1, state.SizeL)

	target, err := p.ToInt()
	assert.NoError(t, err)
	assert.Equal(t, 255, target)
}

// test for { 0x0B: "C" }
func TestParseString(t *testing.T) {
	buf := []byte{0x0B, 0x01, 0x43}
	expectedValue := "C"
	p := &PrimitivePacket{}

	state, err := DecodeToPrimitivePacket(buf, p)
	assert.NoError(t, err)
	assert.EqualValues(t, 2, state.EndPos)
	assert.EqualValues(t, 1, state.SizeL)

	target, err := p.ToUTF8String()
	assert.NoError(t, err)
	assert.Equal(t, expectedValue, target)
}

// test for { 0x04: nil }
func TestEmptyPrimitivePacket(t *testing.T) {
	buf := []byte{0x04, 0x00, 0x03}
	p := &PrimitivePacket{}
	state, err := DecodeToPrimitivePacket(buf, p)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, state.EndPos)
	assert.EqualValues(t, 1, state.SizeL)
	assert.Equal(t, 0, len(p.valbuf))
}

// test for { 0x0C: "" }
func TestParseEmptyString(t *testing.T) {
	buf := []byte{0x0C, 0x00}
	expectedValue := ""
	p := &PrimitivePacket{}

	state, err := DecodeToPrimitivePacket(buf, p)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, state.EndPos)
	assert.EqualValues(t, 1, state.SizeL)

	target, err := p.ToUTF8String()
	assert.NoError(t, err)
	assert.Equal(t, expectedValue, target)
}
