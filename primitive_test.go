package y3

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNilLengthPrimitivePacket(t *testing.T) {
	buf := []byte{0x01}
	p := &PrimitivePacket{}
	state, err := DecodeToPrimitivePacket(buf, p)
	assert.Error(t, err, "invalid y3 packet minimal size")
	assert.EqualValues(t, 0, state.ConsumedBytes)
	assert.EqualValues(t, 0, state.SizeL)
}

// test for { 0x04: nil }
func TestZeroLengthPrimitivePacket(t *testing.T) {
	buf := []byte{0x04, 0x00, 0x03}
	p := &PrimitivePacket{}
	state, err := DecodeToPrimitivePacket(buf, p)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x04, 0x00}, p.GetRawBytes())

	assert.EqualValues(t, 2, state.ConsumedBytes)
	assert.EqualValues(t, 1, state.SizeL)
	assert.Equal(t, 0, len(p.valbuf))

	p = &PrimitivePacket{}
	state, err = DecodeToPrimitivePacket([]byte{0x04, 0x00}, p)
	assert.NoError(t, err)
	assert.EqualValues(t, 2, state.ConsumedBytes)
	assert.EqualValues(t, 1, state.SizeL)
	assert.EqualValues(t, false, p.IsSlice())
}

func TestNagetiveLengthPrimitivePacket(t *testing.T) {
	buf := []byte{0x04, 0x74, 0x01, 0x01}
	p := &PrimitivePacket{}
	state, err := DecodeToPrimitivePacket(buf, p)
	assert.Error(t, err, "invalid y3 packet, negative length")
	assert.EqualValues(t, 2, state.ConsumedBytes)
	assert.EqualValues(t, 1, state.SizeL)
	assert.Equal(t, []byte{0x04, 0x74}, p.GetRawBytes())
	assert.Equal(t, []byte{}, p.GetValBuf())
}

func TestWrongLenPrimitivePacket(t *testing.T) {
	buf := []byte{0x0A, 0x70, 0x01, 0x02}
	p := &PrimitivePacket{}
	state, err := DecodeToPrimitivePacket(buf, p)
	assert.Error(t, err, "invalid y3 packet, negative length")
	assert.EqualValues(t, 2, state.ConsumedBytes)
	assert.EqualValues(t, 1, state.SizeL)
	assert.Equal(t, []byte{0x0A, 0x70}, p.GetRawBytes())
	assert.Equal(t, []byte{}, p.GetValBuf())
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

// test for { 0x0A: 255 }
func TestInt32(t *testing.T) {
	v := 255
	p := NewPrimitivePacketEncoder(0x0A)
	p.SetInt32Value(int32(v))
	buf := p.Encode()

	assert.Equal(t, []byte{0x0A, 0x02, 0x00, 0xFF}, buf)

	packet := &PrimitivePacket{}
	state, err := DecodeToPrimitivePacket(buf, packet)
	assert.NoError(t, err)
	assert.Equal(t, buf, packet.GetRawBytes())
	f, e := packet.ToInt32()
	assert.NoError(t, e)
	assert.EqualValues(t, v, f)
	assert.EqualValues(t, 4, state.ConsumedBytes)
	assert.EqualValues(t, 1, state.SizeL)
}

// test for { 0x0A: 255 }
func TestInt64(t *testing.T) {
	v := 255
	p := NewPrimitivePacketEncoder(0x0A)
	p.SetInt64Value(int64(v))
	buf := p.Encode()

	assert.Equal(t, []byte{0x0A, 0x02, 0x00, 0xFF}, buf)

	packet := &PrimitivePacket{}
	state, err := DecodeToPrimitivePacket(buf, packet)
	assert.NoError(t, err)
	assert.Equal(t, buf, packet.GetRawBytes())
	f, e := packet.ToInt64()
	assert.NoError(t, e)
	assert.EqualValues(t, v, f)
	assert.EqualValues(t, 4, state.ConsumedBytes)
	assert.EqualValues(t, 1, state.SizeL)
}

// test for { 0x0A: 255 }
func TestUInt32(t *testing.T) {
	v := 255
	p := NewPrimitivePacketEncoder(0x0A)
	p.SetUInt32Value(uint32(v))
	buf := p.Encode()

	assert.Equal(t, []byte{0x0A, 0x02, 0x00, 0xFF}, buf)

	packet := &PrimitivePacket{}
	state, err := DecodeToPrimitivePacket(buf, packet)
	assert.NoError(t, err)
	assert.Equal(t, buf, packet.GetRawBytes())
	f, e := packet.ToUInt32()
	assert.NoError(t, e)
	assert.EqualValues(t, v, f)
	assert.EqualValues(t, 4, state.ConsumedBytes)
	assert.EqualValues(t, 1, state.SizeL)
}

// test for { 0x0A: 255 }
func TestUInt64(t *testing.T) {
	v := 255
	p := NewPrimitivePacketEncoder(0x0A)
	p.SetUInt64Value(uint64(v))
	buf := p.Encode()

	assert.Equal(t, []byte{0x0A, 0x02, 0x00, 0xFF}, buf)

	packet := &PrimitivePacket{}
	state, err := DecodeToPrimitivePacket(buf, packet)
	assert.NoError(t, err)
	assert.Equal(t, buf, packet.GetRawBytes())
	f, e := packet.ToUInt64()
	assert.NoError(t, e)
	assert.EqualValues(t, v, f)
	assert.EqualValues(t, 4, state.ConsumedBytes)
	assert.EqualValues(t, 1, state.SizeL)
}

// test for { 0x0A: 1 }
func TestFloat32(t *testing.T) {
	var v float32 = 1
	expect := []byte{0x0A, 0x02, 0x3F, 0x80}
	p := NewPrimitivePacketEncoder(0x0A)
	p.SetFloat32Value(float32(v))
	buf := p.Encode()
	assert.Equal(t, expect, buf)

	packet := &PrimitivePacket{}
	state, err := DecodeToPrimitivePacket(buf, packet)
	assert.NoError(t, err)
	f, e := packet.ToFloat32()
	assert.NoError(t, e)
	assert.EqualValues(t, v, f)
	assert.Equal(t, buf, packet.GetRawBytes())
	assert.EqualValues(t, 4, state.ConsumedBytes)
	assert.EqualValues(t, 1, state.SizeL)
}

// test for { 0x0A: 255 }
func TestFloat64(t *testing.T) {
	var v float64 = 1
	expect := []byte{0x0A, 0x02, 0x3F, 0xF0}
	p := NewPrimitivePacketEncoder(0x0A)
	p.SetFloat64Value(float64(v))
	buf := p.Encode()
	assert.Equal(t, expect, buf)

	packet := &PrimitivePacket{}
	state, err := DecodeToPrimitivePacket(buf, packet)
	assert.NoError(t, err)
	f, e := packet.ToFloat64()
	assert.NoError(t, e)
	assert.EqualValues(t, v, f)
	assert.Equal(t, buf, packet.GetRawBytes())
	assert.EqualValues(t, 4, state.ConsumedBytes)
	assert.EqualValues(t, 1, state.SizeL)
}

// test for { 0x0B: "yomo" }
func TestString(t *testing.T) {
	expect := []byte{0x0B, 0x04, 0x79, 0x6F, 0x6D, 0x6F}
	v := "yomo"
	p := NewPrimitivePacketEncoder(0x0B)
	p.SetStringValue(v)
	buf := p.Encode()
	assert.Equal(t, expect, buf)

	packet := &PrimitivePacket{}

	state, err := DecodeToPrimitivePacket(buf, packet)
	assert.NoError(t, err)
	assert.Equal(t, buf, packet.GetRawBytes())
	assert.Equal(t, []byte{0x79, 0x6F, 0x6D, 0x6F}, p.GetValBuf())
	target, err := packet.ToUTF8String()
	assert.NoError(t, err)
	assert.Equal(t, v, target)
	assert.EqualValues(t, 6, state.ConsumedBytes)
	assert.EqualValues(t, 1, state.SizeL)
}

// test for { 0x0C: "" }
func TestParseEmptyString(t *testing.T) {
	expect := []byte{0x0C, 0x00}
	v := ""
	p := NewPrimitivePacketEncoder(0x0C)
	p.SetStringValue(v)
	buf := p.Encode()
	assert.Equal(t, expect, buf)

	packet := &PrimitivePacket{}

	state, err := DecodeToPrimitivePacket(buf, packet)
	assert.NoError(t, err)
	assert.Equal(t, buf, packet.GetRawBytes())
	target, err := packet.ToUTF8String()
	assert.NoError(t, err)
	assert.Equal(t, v, target)
	assert.EqualValues(t, 2, state.ConsumedBytes)
	assert.EqualValues(t, 1, state.SizeL)
}

// test for { 0x2C: true }
func TestPrimitivePacketBool(t *testing.T) {
	expect := []byte{0x2C, 0x01, 0x01}
	v := true
	p := NewPrimitivePacketEncoder(0x2C)
	p.SetBoolValue(v)
	buf := p.Encode()
	assert.Equal(t, expect, buf)

	packet := &PrimitivePacket{}

	state, err := DecodeToPrimitivePacket(buf, packet)
	assert.NoError(t, err)
	assert.Equal(t, buf, packet.GetRawBytes())
	target, err := packet.ToBool()
	assert.NoError(t, err)
	assert.Equal(t, v, target)
	assert.EqualValues(t, 3, state.ConsumedBytes)
	assert.EqualValues(t, 1, state.SizeL)
}

// test for { 0x2C: false }
func TestPrimitivePacketBool2(t *testing.T) {
	buf := []byte{0x2C, 0x00}
	packet := &PrimitivePacket{}

	state, err := DecodeToPrimitivePacket(buf, packet)
	assert.NoError(t, err)
	assert.Equal(t, buf, packet.GetRawBytes())
	target, err := packet.ToBool()
	assert.NoError(t, err)
	assert.Equal(t, false, target)
	assert.EqualValues(t, 2, state.ConsumedBytes)
	assert.EqualValues(t, 1, state.SizeL)
}

// test for { 0x2C: false }
func TestPrimitivePacketBool3(t *testing.T) {
	buf := []byte{0x2C, 0x01, 0x08}
	packet := &PrimitivePacket{}

	state, err := DecodeToPrimitivePacket(buf, packet)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x2C, 0x01, 0x08}, packet.GetRawBytes())
	target, err := packet.ToBool()
	assert.NoError(t, err)
	assert.Equal(t, false, target)
	assert.EqualValues(t, 3, state.ConsumedBytes)
	assert.EqualValues(t, 1, state.SizeL)
}

func TestPrimitivePacketBytes(t *testing.T) {
	v := make([]byte, 255)
	for i := 0; i < 255; i++ {
		v[i] = byte(i)
	}
	expect := append([]byte{0x01, 0x81, 0x7F}, v...)
	p := NewPrimitivePacketEncoder(0x01)
	p.SetBytesValue(v)
	buf := p.Encode()
	assert.Equal(t, expect, buf)

	packet := &PrimitivePacket{}
	state, err := DecodeToPrimitivePacket(buf, packet)
	assert.NoError(t, err)
	assert.Equal(t, buf, packet.GetRawBytes())
	target := packet.ToBytes()
	assert.NoError(t, err)
	assert.Equal(t, v, target)
	assert.EqualValues(t, 255+2+1, state.ConsumedBytes)
	assert.EqualValues(t, 2, state.SizeL)
}
