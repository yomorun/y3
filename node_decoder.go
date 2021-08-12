package y3

import (
	"bytes"
	"errors"

	"github.com/yomorun/y3/encoding"
	"github.com/yomorun/y3/utils"
)

func parsePayload(b []byte) (consumedBytes int, ifNodePacket bool, np *NodePacket, pp *PrimitivePacket, err error) {
	if len(b) == 0 {
		return 0, false, nil, nil, errors.New("parsePacket params can not be nil")
	}

	pos := 0
	// NodePacket
	if ok := utils.IsNodePacket(b[pos]); ok {
		np = &NodePacket{}
		endPos, err := DecodeToNodePacket(b, np)
		return endPos, true, np, nil, err
	}

	pp = &PrimitivePacket{}
	state, err := DecodeToPrimitivePacket(b, pp)
	return state.ConsumedBytes, false, nil, pp, err
}

// DecodeNodePacket parse out whole buffer to a NodePacket
func DecodeToNodePacket(buf []byte, pct *NodePacket) (consumedBytes int, err error) {
	if len(buf) == 0 {
		return 0, errors.New("empty buf")
	}

	pct.basePacket = &basePacket{
		valbuf: buf,
		buf:    &bytes.Buffer{},
	}

	pct.NodePackets = map[byte]NodePacket{}
	pct.PrimitivePackets = map[byte]PrimitivePacket{}

	pos := 0

	// `Tag`
	tag := NewTag(buf[pos])
	pct.basePacket.tag = tag
	pct.buf.WriteByte(buf[pos])
	pos++

	// `Length`: the type is `varint`
	tmpBuf := buf[pos:]
	var vallen int32
	codec := encoding.VarCodec{}
	err = codec.DecodePVarInt32(tmpBuf, &vallen)
	if err != nil {
		return 0, err
	}
	pct.basePacket.length = int(vallen)
	pct.buf.Write(buf[pos : pos+codec.Size])
	pos += codec.Size
	// if `Length` is 0, means empty node packet
	if vallen == 0 {
		return pos, nil
	}

	// `Value`
	// `raw` is pct.Length() length
	vl := int(vallen)
	if vl < 0 {
		return pos, errors.New("found L of V smaller than 0")
	}
	endPos := pos + vl
	pct.basePacket.valbuf = buf[pos:endPos]
	pct.buf.Write(buf[pos:endPos])

	// Parse value to Packet
	for {
		if pos >= endPos || pos >= len(buf) {
			break
		}
		_p, isNode, np, pp, err := parsePayload(buf[pos:endPos])
		pos += _p
		if err != nil {
			return 0, err
		}
		if isNode {
			pct.NodePackets[np.basePacket.tag.SeqID()] = *np
		} else {
			pct.PrimitivePackets[byte(pp.SeqID())] = *pp
		}
	}

	consumedBytes = endPos
	return consumedBytes, nil
}
