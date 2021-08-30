package y3

import (
	"bytes"
	"errors"
	"io"

	"github.com/yomorun/y3/encoding"
)

const maxSeqID = 0x3F
const flagBitNode = 0x80
const wipeFlagBits = 0x3F
const flagBitSlice = 0x40

var (
	ErrInvalidSeqID    = errors.New("y3.Builder: SeqID should >= 0 and =< 0x3F")
	ErrBuildIncomplete = errors.New("y3.Builder: invalid structure of packet")
)

// Packet
type Packet interface {
	SeqID() int
}

// Packet implementation
type ChunkPacket struct {
	t   T
	l   L
	sid int
}

func (p *ChunkPacket) SeqID() int { return p.sid }

var _ Packet = &ChunkPacket{}

// TLV
type T byte

func NewT(seqID int, isNode bool, isSlice bool) (T, error) {
	if seqID < 0 || seqID > maxSeqID {
		return 0, ErrInvalidSeqID
	}

	if isNode {
		seqID |= flagBitNode
	}

	if isSlice {
		seqID |= flagBitSlice
	}

	return T(seqID), nil
}

func (t T) Sid() int {
	return int(t & wipeFlagBits)
}

type L struct {
	buf  []byte
	size int
}

func (l L) Parse(len int) error {
	if len < -1 {
		return errors.New("y3.Len: len can't less than -1")
	}

	var vallen int32
	l.size = encoding.SizeOfPVarInt32(vallen)
	codec := encoding.VarCodec{Size: l.size}
	tmp := make([]byte, l.size)
	err := codec.EncodePVarInt32(tmp, vallen)
	if err != nil {
		panic(err)
	}
	copy(l.buf, tmp)
	return nil
}

// Builder
type Builder struct {
	buf       *bytes.Buffer
	tag       T
	len       *L
	valReader io.Reader
	nodes     map[int]Packet
	state     int
}

func (b *Builder) SetSeqID(seqID int, isNode bool, isSlice bool) error {
	t, err := NewT(seqID, isNode, isSlice)
	if err != nil {
		return err
	}
	b.tag = t
	b.state |= 0x01
	return nil
}

func (b *Builder) SetLen(length int) error {
	var l L
	err := l.Parse(length)
	if err != nil {
		return err
	}
	b.len = &l
	b.state |= 0x02
	return nil
}

func (b *Builder) SetVal(buf []byte) {
	b.buf.Write(buf)
	b.state |= 0x04
}

func (b *Builder) SetValReader(r io.Reader) {
	b.valReader = r
	b.state |= 0x01
}

func (b *Builder) Packet() (Packet, error) {
	if b.state != 0x07 {
		return nil, ErrBuildIncomplete
	}
	return &ChunkPacket{
		t: b.tag,
		l: *b.len,
	}, nil
}

func (b *Builder) AddPacket(child Packet) {
	b.nodes[child.SeqID()] = child
}
