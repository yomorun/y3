package y3

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/yomorun/y3/encoding"
)

type decodeState struct {
	ConsumedBytes int
	SizeL         int
}

// DecodePrimitivePacket parse out whole buffer to a PrimitivePacket
//
// Examples:
// [0x01, 0x01, 0x01] -> Key=0x01, Value=0x01
// [0x41, 0x06, 0x03, 0x01, 0x61, 0x04, 0x01, 0x62] -> key=0x03, value=0x61; key=0x04, value=0x62
func DecodeToPrimitivePacket(buf []byte, p *PrimitivePacket) (decodeState, error) {
	decoder := decodeState{
		ConsumedBytes: 0,
		SizeL:         0,
	}

	if buf == nil || len(buf) < primitivePacketBufferMinimalLength {
		return decoder, errors.New("invalid y3 packet minimal size")
	}

	p.basePacket = &basePacket{
		valbuf: buf,
		buf:    &bytes.Buffer{},
	}

	var pos = 0
	// first byte is `Tag`
	p.tag = NewTag(buf[pos])
	p.buf.WriteByte(buf[pos])
	pos += 1
	decoder.ConsumedBytes = pos

	// read `Varint` from buf for `Length of value`
	tmpBuf := buf[pos:]
	var bufLen int32
	codec := encoding.VarCodec{}
	err := codec.DecodePVarInt32(tmpBuf, &bufLen)
	if err != nil {
		return decoder, err
	}
	if codec.Size < 1 {
		return decoder, errors.New("malformed, size of Length can not smaller than 1")
	}

	// codec.Size describes how many bytes used to represent `Length`
	p.buf.Write(buf[pos : pos+codec.Size])
	pos += codec.Size

	decoder.ConsumedBytes = pos
	decoder.SizeL = codec.Size

	// the length of value
	p.length = int(bufLen)
	if p.length == 0 {
		p.valbuf = []byte{}
		return decoder, nil
	}

	// the next `p.length` bytes store value
	endPos := pos + p.length

	if pos > endPos || endPos > len(buf) || pos > len(buf) {
		return decoder, fmt.Errorf("beyond the boundary, pos=%v, endPos=%v", pos, endPos)
	}
	p.valbuf = buf[pos:endPos]
	p.buf.Write(buf[pos:endPos])

	decoder.ConsumedBytes = endPos
	return decoder, nil
}
