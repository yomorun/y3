package y3

import (
	"bytes"
)

// basePacket is the base type of the NodePacket and PrimitivePacket
type basePacket struct {
	tag    *Tag
	length int
	valbuf []byte
	buf    *bytes.Buffer
}

// GetRawBytes get all raw bytes of this packet
func (bp *basePacket) GetRawBytes() []byte {
	return bp.buf.Bytes()
}

// Length return the length of Val this packet
func (bp *basePacket) Length() int {
	return bp.length
}

// SeqID returns Tag of this packet
func (bp *basePacket) SeqID() byte {
	return bp.tag.SeqID()
}

// IsSlice determine if the current node is a Slice
func (bp *basePacket) IsSlice() bool {
	return bp.tag.IsSlice()
}

// GetValBuf get raw buffer of Val of this packet
func (bp *basePacket) GetValBuf() []byte {
	return bp.valbuf
}
