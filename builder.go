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
	// SeqID returns the sequence ID of this packet.
	SeqID() int
	// Size returns the size of whole packet.
	Size() int
	// VSize returns the size of V.
	VSize() int
	// Raw returns the whole bytes.
	Raw() []byte
	// Reader returns an io.Reader which returns who bytes.
	Reader() io.Reader
	// GetValReader returns an io.Reader which holds V.
	GetValReader() io.Reader
}

// Packet implementation
type StreamPacket struct {
	t             T
	l             L
	valbuf        []byte
	r             io.Reader
	chunkMode     bool
	chunkSize     int
	hasChunkChild bool
}

func (p *StreamPacket) SeqID() int { return p.t.Sid() }

// Size returns the size of whole packet.
func (p *StreamPacket) Size() int {
	// T.Size() + L.Size() + V.Size()
	return p.t.Size() + p.l.Size() + p.l.len
}

// VSize returns the size of V.
func (p *StreamPacket) VSize() int { return p.l.len }

func (p *StreamPacket) Raw() []byte {
	buf := new(bytes.Buffer)
	p.writeTL(buf)

	//if p.chunkSize < 1 {
	buf.Write(p.valbuf)
	//}

	return buf.Bytes()
}

func (p *StreamPacket) GetValReader() io.Reader {
	return p.r
}

func (p *StreamPacket) Reader() io.Reader {
	if p.chunkSize <= 0 {
		return nil
	}

	buf := new(bytes.Buffer)
	p.writeTL(buf)
	// child T/L is in buf
	lenTL := len(buf.Bytes())
	buf.Write(p.valbuf)

	return &yR{
		buf:    buf,
		src:    p.r,
		length: lenTL + p.l.len,
		slen:   p.chunkSize,
	}
}

func (p *StreamPacket) writeTL(buf *bytes.Buffer) {
	buf.Write(p.t.Raw())
	buf.Write(p.l.Raw())
}

var _ Packet = &StreamPacket{}

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

func (t T) Raw() []byte {
	return []byte{byte(t)}
}

func (t T) IsNode() bool {
	return t&flagBitNode == flagBitNode
}

func (t T) Size() int {
	return 1
}

// TLV_L
type L struct {
	buf  []byte
	size int
	len  int
}

func NewL(len int) (L, error) {
	var l = L{}
	if len < -1 {
		return l, errors.New("y3.Len: len can't less than -1")
	}

	vallen := int32(len)
	l.size = encoding.SizeOfPVarInt32(vallen)
	codec := encoding.VarCodec{Size: l.size}
	tmp := make([]byte, l.size)
	err := codec.EncodePVarInt32(tmp, vallen)
	if err != nil {
		panic(err)
	}
	l.buf = make([]byte, l.size)
	copy(l.buf, tmp)
	l.len = len
	return l, nil
}

func (l L) Raw() []byte {
	return l.buf
}

// Size returns how many bytes used to represent this L
func (l L) Size() int {
	return l.size
}

// Value returns the value this L represents
func (l L) Value() int {
	return l.len
}

// Builder
type Builder struct {
	tag           T
	len           *L
	valReader     io.Reader
	valReaderSize int
	nodes         map[int]Packet
	state         int
	size          int32 // size of value
	isChunked     bool
	valbuf        *bytes.Buffer
	done          bool
	seqID         int
	isNode        bool
	hasChunkChild bool
}

// Size returns the size of V.
func (b *Builder) Size() int {
	return int(b.size)
}

func (b *Builder) generateTag() error {
	t, err := NewT(b.seqID, b.isNode)
	if err != nil {
		return err
	}
	b.tag = t
	b.state |= 0x01
	return nil
}

// SetSeqID set sequence ID of a y3 packet.
// isNode
func (b *Builder) SetSeqID(seqID int, isNode bool) {
	b.seqID = seqID
	b.isNode = isNode
	// init
	b.valbuf = new(bytes.Buffer)
	b.nodes = make(map[int]Packet)
}

func (b *Builder) generateSize() error {
	l, err := NewL(int(b.size))
	if err != nil {
		return err
	}
	b.len = &l
	b.state |= 0x02
	return nil
}

func (b *Builder) SetValReader(r io.Reader, size int) {
	b.isChunked = true
	b.valReader = r
	b.state |= 0x04
	b.size += int32(size)
	b.valReaderSize = size
}

func (b *Builder) AddValBytes(buf []byte) {
	b.size += int32(len(buf))
	b.valbuf.Write(buf)
	b.isChunked = false
	b.state |= 0x04
}

func (b *Builder) Packet() (Packet, error) {
	err := b.generateTag()
	if err != nil {
		return nil, err
	}

	err = b.generateSize()
	if err != nil {
		return nil, err
	}

	if b.state != 0x07 {
		return nil, ErrBuildIncomplete
	}

	if b.isChunked {
		return &StreamPacket{
			t:             b.tag,
			l:             *b.len,
			r:             b.valReader,
			chunkMode:     true,
			chunkSize:     b.valReaderSize,
			valbuf:        b.valbuf.Bytes(),
			hasChunkChild: true,
		}, err
	} else {
		return &StreamPacket{
			t:         b.tag,
			l:         *b.len,
			valbuf:    b.valbuf.Bytes(),
			chunkMode: false,
		}, err
	}
}

func (b *Builder) AddPacket(child Packet) error {
	if b.done {
		return errors.New("y3.Builder: can not add Packet after StreamPacket")
	}
	b.nodes[child.SeqID()] = child
	buf := child.Raw()
	b.AddValBytes(buf)
	return nil
}

func (b *Builder) AddStreamPacket(child Packet) {
	b.done = true
	b.valReader = child.GetValReader()
	b.valReaderSize = child.VSize()
	b.state |= 0x04
	b.nodes[child.SeqID()] = child
	// add V size
	b.size += int32(child.Size())
	// append the bytes of child
	buf := child.Raw()
	b.valbuf.Write(buf)
	b.isChunked = true
	b.hasChunkChild = true
}
