package y3

import (
	"fmt"
)

// basePacket is the base type of the NodePacket and PrimitivePacket
type basePacket struct {
	tag    *Tag
	length int
	valbuf []byte
}

func (bp *basePacket) Length() int {
	return bp.length
}

// SeqID returns Tag Key
func (bp *basePacket) SeqID() int {
	return int(bp.tag.SeqID())
}

// IsSlice determine if the current node is a Slice
func (bp *basePacket) IsSlice() bool {
	return bp.tag.IsSlice()
}

// GetValBuf get raw buffer of NodePacket
func (bp *basePacket) GetValBuf() []byte {
	return bp.valbuf
}

// String prints debug info
func (p *PrimitivePacket) String() string {
	return fmt.Sprintf("Tag=%#x, Length=%v, RawDataLength=%v, Raw=[%#x]",
		p.SeqID(), p.length, len(p.valbuf), p.valbuf)
}
