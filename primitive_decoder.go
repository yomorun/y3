package y3

import (
	"errors"
	"fmt"

	"github.com/yomorun/y3/encoding"
)

type decodeState struct {
	EndPos int
	SizeL  int
}

// DecodePrimitivePacket parse out whole buffer to a PrimitivePacket
//
// Examples:
// [0x01, 0x01, 0x01] -> Key=0x01, Value=0x01
// [0x41, 0x06, 0x03, 0x01, 0x61, 0x04, 0x01, 0x62] -> key=0x03, value=0x61; key=0x04, value=0x62
func DecodeToPrimitivePacket(buf []byte, p *PrimitivePacket) (decodeState, error) {
	decoder := decodeState{
		EndPos: 0,
		SizeL:  0,
	}

	if buf == nil || len(buf) < primitivePacketBufferMinimalLength {
		return decoder, errors.New("invalid y3 packet minimal size")
	}

	p.basePacket = &basePacket{
		valbuf: buf,
	}

	var pos = 0
	// first byte is `Tag`
	p.tag = NewTag(int(buf[pos]))
	pos += 1

	// read `Varint` from buf for `Length of value`
	tmpBuf := buf[pos:]
	var bufLen int32
	codec := encoding.VarCodec{}
	err := codec.DecodePVarInt32(tmpBuf, &bufLen)
	if err != nil {
		return decoder, err
	}
	if codec.Size < 1 {
		return decodeState{EndPos: pos, SizeL: codec.Size}, errors.New("malformed, size of Length can not smaller than 1")
	}

	// the length of value
	p.length = int(bufLen)
	if p.length == 0 {
		p.valbuf = []byte{}
		return decodeState{EndPos: pos, SizeL: codec.Size}, nil
	}

	// codec.Size describes how many bytes used to represent `Length`
	pos += codec.Size

	// the next `p.length` bytes store value
	endPos := pos + p.length

	if pos > endPos || endPos > len(buf) || pos > len(buf) {
		return decodeState{EndPos: endPos, SizeL: codec.Size}, fmt.Errorf("beyond the boundary, pos=%v, endPos=%v", pos, endPos)
	}
	p.valbuf = buf[pos:endPos]

	return decodeState{EndPos: endPos - 1, SizeL: codec.Size}, nil
}

// // DecodePrimitivePacket parse out whole buffer to a PrimitivePacket
// //
// // Examples:
// // [0x01, 0x01, 0x01] -> Key=0x01, Value=0x01
// // [0x41, 0x06, 0x03, 0x01, 0x61, 0x04, 0x01, 0x62] -> key=0x03, value=0x61; key=0x04, value=0x62
// func DecodePrimitivePacket(buf []byte) (packet *PrimitivePacket, endPos int, sizeL int, err error) {
// 	if buf == nil || len(buf) < primitivePacketBufferMinimalLength {
// 		return nil, 0, 0, errors.New("invalid y3 packet minimal size")
// 	}

// 	p := &PrimitivePacket{
// 		basePacket: &basePacket{
// 			valbuf: buf,
// 		},
// 	}

// 	var pos = 0
// 	// first byte is `Tag`
// 	p.tag = NewTag(int(buf[pos]))
// 	pos++

// 	// read `Varint` from buf for `Length of value`
// 	tmpBuf := buf[pos:]
// 	var bufLen int32
// 	codec := encoding.VarCodec{}
// 	err = codec.DecodePVarInt32(tmpBuf, &bufLen)
// 	if err != nil {
// 		return nil, 0, 0, err
// 	}
// 	sizeL = codec.Size

// 	if sizeL < 1 {
// 		return nil, 0, sizeL, errors.New("malformed, size of Length can not smaller than 1")
// 	}

// 	p.length = uint32(bufLen)
// 	pos += sizeL

// 	endPos = pos + int(p.length)

// 	// logger.Debugf(">>> sizeL=%v, length=%v, pos=%v, endPos=%v", sizeL, p.length, pos, endPos)

// 	if pos > endPos || endPos > len(buf) || pos > len(buf) {
// 		return nil, 0, sizeL, fmt.Errorf("beyond the boundary, pos=%v, endPos=%v", pos, endPos)
// 	}
// 	p.valbuf = buf[pos:endPos]
// 	// logger.Debugf("valBuf = %#X", p.valBuf)

// 	return p, endPos, sizeL, nil
// }
