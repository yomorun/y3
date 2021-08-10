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
	assert.EqualValues(t, 0, state.ConsumedBytes)
	assert.EqualValues(t, 0, state.SizeL)

	p = &PrimitivePacket{}
	state, err = DecodeToPrimitivePacket([]byte{0x01, 0x00}, p)
	assert.NoError(t, err)
	assert.EqualValues(t, 2, state.ConsumedBytes)
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
	assert.Equal(t, buf, p.GetRawBytes())

	assert.Equal(t, expectedTag, p.tag.SeqID())
	assert.Equal(t, 1, p.length)
	assert.EqualValues(t, expectedValue, p.valbuf)
	assert.Equal(t, 3, state.ConsumedBytes)
	assert.EqualValues(t, 1, state.SizeL)
}

// test for { 0x0A: [0x81, 0x7F] }
func TestParseInt32(t *testing.T) {
	buf := []byte{0x0A, 0x02, 0x81, 0x7F}
	p := &PrimitivePacket{}

	state, err := DecodeToPrimitivePacket(buf, p)
	assert.NoError(t, err)
	assert.Equal(t, buf, p.GetRawBytes())

	assert.EqualValues(t, 4, state.ConsumedBytes)
	assert.EqualValues(t, 1, state.SizeL)

	target, err := p.ToInt()
	assert.NoError(t, err)
	assert.EqualValues(t, 255, target)

	v1, err := p.ToUInt32()
	assert.NoError(t, err)
	assert.EqualValues(t, 255, v1)

	v2, err := p.ToInt64()
	assert.NoError(t, err)
	assert.EqualValues(t, 255, v2)

	v3, err := p.ToUInt64()
	assert.NoError(t, err)
	assert.EqualValues(t, 255, v3)

	v4, err := p.ToBool()
	assert.NoError(t, err)
	assert.EqualValues(t, false, v4)
}

// test for { 0x03: [0x3F, 0x9D, 0x70, 0xA4] }
func TestParseFloat32(t *testing.T) {
	buf := []byte{0x03, 0x04, 0x3F, 0x9D, 0x70, 0xA4}
	p := &PrimitivePacket{}

	_, err := DecodeToPrimitivePacket(buf, p)
	assert.NoError(t, err)
	assert.Equal(t, buf, p.GetRawBytes())

	v4, err := p.ToFloat32()
	assert.NoError(t, err)
	assert.EqualValues(t, 1.23, v4)
}

// test for { 0x03: [0x3F, 0xF3, 0xAE, 0x14, 0x7A, 0xE1, 0x47, 0xAE] }
func TestParseFloat64(t *testing.T) {
	buf := []byte{0x03, 0x08, 0x3F, 0xF3, 0xAE, 0x14, 0x7A, 0xE1, 0x47, 0xAE}
	p := &PrimitivePacket{}

	_, err := DecodeToPrimitivePacket(buf, p)
	assert.NoError(t, err)
	assert.Equal(t, buf, p.GetRawBytes())

	v4, err := p.ToFloat64()
	assert.NoError(t, err)
	assert.EqualValues(t, 1.23, v4)
}

// test for { 0x0B: "C" }
func TestParseString(t *testing.T) {
	buf := []byte{0x0B, 0x01, 0x43}
	expectedValue := "C"
	p := &PrimitivePacket{}

	state, err := DecodeToPrimitivePacket(buf, p)
	assert.NoError(t, err)
	assert.Equal(t, buf, p.GetRawBytes())

	assert.EqualValues(t, 3, state.ConsumedBytes)
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
	assert.Equal(t, []byte{0x04, 0x00}, p.GetRawBytes())

	assert.EqualValues(t, 2, state.ConsumedBytes)
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
	assert.Equal(t, buf, p.GetRawBytes())
	assert.Equal(t, false, p.IsSlice())

	assert.EqualValues(t, 2, state.ConsumedBytes)
	assert.EqualValues(t, 1, state.SizeL)

	target, err := p.ToUTF8String()
	assert.NoError(t, err)
	assert.Equal(t, expectedValue, target)
}

func TestPrimitivePacketEncodeBytesAndString(t *testing.T) {
	buf := []byte("yomo")
	p := NewPrimitivePacketEncoder(0x02)
	p.SetBytesValue(buf)
	res := p.Encode()
	assert.Equal(t, []byte{0x02, 0x04, 0x79, 0x6F, 0x6D, 0x6F}, res)
	assert.Equal(t, []byte{0x79, 0x6F, 0x6D, 0x6F}, p.GetValBuf())

	p = NewPrimitivePacketEncoder(0x02)
	p.SetStringValue("yomo")
	res = p.Encode()
	assert.Equal(t, []byte{0x02, 0x04, 0x79, 0x6F, 0x6D, 0x6F}, res)
	assert.Equal(t, []byte{0x79, 0x6F, 0x6D, 0x6F}, p.GetValBuf())
}

func TestPrimitivePacketEncodeNumbers(t *testing.T) {
	p := NewPrimitivePacketEncoder(0x03)
	p.SetInt32Value(-128)
	res := p.Encode()
	assert.Equal(t, []byte{0x03, 0x02, 0xFF, 0x00}, res)

	p = NewPrimitivePacketEncoder(0x03)
	p.SetUInt32Value(128)
	res = p.Encode()
	assert.Equal(t, []byte{0x03, 0x02, 0x81, 0x00}, res)

	p = NewPrimitivePacketEncoder(0x03)
	p.SetInt64Value(-128)
	res = p.Encode()
	assert.Equal(t, []byte{0x03, 0x02, 0xFF, 0x00}, res)

	p = NewPrimitivePacketEncoder(0x03)
	p.SetUInt64Value(128)
	res = p.Encode()
	assert.Equal(t, []byte{0x03, 0x02, 0x81, 0x00}, res)

	p = NewPrimitivePacketEncoder(0x03)
	p.SetFloat32Value(1.23)
	res = p.Encode()
	assert.Equal(t, []byte{0x03, 0x04, 0x3F, 0x9D, 0x70, 0xA4}, res)

	p = NewPrimitivePacketEncoder(0x03)
	p.SetFloat64Value(1.23)
	res = p.Encode()
	assert.Equal(t, []byte{0x03, 0x08, 0x3F, 0xF3, 0xAE, 0x14, 0x7A, 0xE1, 0x47, 0xAE}, res)

	p = NewPrimitivePacketEncoder(0x03)
	p.SetBoolValue(true)
	res = p.Encode()
	assert.Equal(t, []byte{0x03, 0x01, 0x01}, res)

	p = NewPrimitivePacketEncoder(0x03)
	p.SetBoolValue(false)
	res = p.Encode()
	assert.Equal(t, []byte{0x03, 0x01, 0x00}, res)
}
