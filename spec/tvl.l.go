package spec

import (
	"bytes"
	"errors"
	"io"

	"github.com/yomorun/y3/encoding"
)

// L is the Length in a TLV structure
type L struct {
	buf  []byte
	size int
	len  int
}

// NewL will take an int type len as parameter and return L to
// represent the sieze of V in a TLV. an integer will be encode as
// a PVarInt32 type to represent the value.
func NewL(len int) (L, error) {
	var l = L{}
	if len < -1 {
		return l, errors.New("y3.L: len can't less than -1")
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

// Raw will return the raw bytes of L.
func (l L) Bytes() []byte {
	return l.buf
}

// Size returns how many bytes used to represent this L.
func (l L) Size() int {
	return l.size
}

// Value returns the size of V.
func (l L) VSize() int {
	return int(l.len)
}

// ReadL read L from bufio.Reader
func ReadL(r io.Reader) (*L, error) {
	lenbuf := bytes.Buffer{}
	for {
		b, err := readByte(r)
		if err != nil {
			return nil, err
		}
		lenbuf.WriteByte(b)
		if b&msb != msb {
			break
		}
	}

	buf := lenbuf.Bytes()

	// decode to L
	length, err := decodeL(buf)
	if err != nil {
		return nil, err
	}

	return &L{
		buf:  buf,
		len:  int(length),
		size: len(buf),
	}, nil
}

func decodeL(buf []byte) (length int32, err error) {
	codec := encoding.VarCodec{}
	err = codec.DecodePVarInt32(buf, &length)
	return length, err
}
