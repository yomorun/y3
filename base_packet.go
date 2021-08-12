package y3

import (
	"bytes"
)

// basePacket is the base type of the NodePacket and PrimitivePacket
type basePacket struct {
	tag *Tag
	// tagbuf []byte
	length int
	// lenbuf []byte
	valbuf []byte
	buf    *bytes.Buffer
}

// func (bp *basePacket) buildBuf() {
// 	bp.buf = append(bp.tagbuf, bp.lenbuf...)
// 	bp.buf = append(bp.buf, bp.valbuf...)
// }

// GetRawBytes get raw bytes of this packet
func (bp *basePacket) GetRawBytes() []byte {
	return bp.buf.Bytes()
}

func (bp *basePacket) Length() int {
	return bp.length
}

// SeqID returns Tag Key
func (bp *basePacket) SeqID() byte {
	return bp.tag.SeqID()
}

// IsSlice determine if the current node is a Slice
func (bp *basePacket) IsSlice() bool {
	return bp.tag.IsSlice()
}

// GetValBuf get raw buffer of NodePacket
func (bp *basePacket) GetValBuf() []byte {
	return bp.valbuf
}
