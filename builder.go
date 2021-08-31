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
	Size() int
	Raw() []byte
	Reader() io.Reader
}

// Packet implementation
type ChunkPacket struct {
	t      T
	l      L
	valbuf []byte
	r      io.Reader
}

func (p *ChunkPacket) SeqID() int { return p.t.Sid() }

func (p *ChunkPacket) Size() int { return p.l.Size() }

func (p *ChunkPacket) Raw() []byte { return p.valbuf }

func (p *ChunkPacket) Reader() io.Reader { return p.r }

var _ Packet = &ChunkPacket{}

// TLV_T
type T byte

func NewT(seqID int, isNode bool) (T, error) {
	if seqID < 0 || seqID > maxSeqID {
		return 0, ErrInvalidSeqID
	}

	if isNode {
		seqID |= flagBitNode
	}

	return T(seqID), nil
}

func (t T) Sid() int {
	return int(t & wipeFlagBits)
}

func (t T) Bytes() []byte {
	return []byte{byte(t)}
}

// TLV_L
type L struct {
	buf  []byte
	size int
}

func NewL(len int) (L, error) {
	var l = L{}
	if len < -1 {
		return l, errors.New("y3.Len: len can't less than -1")
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
	return l, nil
}

func (l L) Bytes() []byte {
	return l.buf
}

func (l L) Size() int {
	return l.size
}

// Builder
type Builder struct {
	tag       T
	len       *L
	valReader io.Reader
	nodes     map[int]Packet
	state     int
	size      int32
	isChunked bool
	valbuf    *bytes.Buffer
	done      bool
}

func (b *Builder) Size() int {
	return int(b.size)
}

func (b *Builder) SetSeqID(seqID int, isNode bool) error {
	t, err := NewT(seqID, isNode)
	if err != nil {
		return err
	}
	b.tag = t
	b.state |= 0x01
	b.valbuf = new(bytes.Buffer)
	b.nodes = make(map[int]Packet)
	return nil
}

func (b *Builder) SetSize(size int) {
	b.size = int32(size)
}

func (b *Builder) setLen(length int) error {
	l, err := NewL(length)
	if err != nil {
		return err
	}
	b.len = &l
	b.state |= 0x02
	return nil
}

func (b *Builder) SetValReader(r io.Reader) {
	b.isChunked = true
	b.valReader = r
	b.state |= 0x04
}

func (b *Builder) AddValBytes(buf []byte) {
	b.size += int32(len(buf))
	b.valbuf.Write(buf)
	b.isChunked = false
	b.state |= 0x04
}

func (b *Builder) Packet() (Packet, error) {
	var err error
	if b.state&0x02 != 0x02 {
		err = b.setLen(int(b.size))
		if err != nil {
			return nil, err
		}
	}

	if b.state != 0x07 {
		return nil, ErrBuildIncomplete
	}

	if b.isChunked {
		return &ChunkPacket{
			t: b.tag,
			l: *b.len,
			r: b.valReader,
		}, err
	} else {
		return &ChunkPacket{
			t:      b.tag,
			l:      *b.len,
			valbuf: b.valbuf.Bytes(),
		}, err
	}
}

func (b *Builder) AddPacket(child Packet) error {
	if b.done {
		return errors.New("y3.Builder: can not add Packet after ChunkPacket")
	}
	b.nodes[child.SeqID()] = child
	buf := child.Raw()
	b.AddValBytes(buf)
	return nil
}

func (b *Builder) AddChunkPacket(child Packet) {
	b.done = true
	b.size += int32(child.Size())
	b.valReader = child.Reader()
}

func (b *Builder) Reader() io.Reader {
	return b.valReader
}
