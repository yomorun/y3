package y3

import (
	"bytes"

	"github.com/yomorun/y3/encoding"
)

// Encoder will encode object to Y3 encoding
type encoder struct {
	seqID    byte
	valbuf   []byte
	isNode   bool
	isArray  bool
	buf      *bytes.Buffer
	complete bool
}

type iEncoder interface {
	Encode() []byte
}

func (enc *encoder) GetValBuf() []byte {
	return enc.valbuf
}

func (enc *encoder) IsEmpty() bool {
	return len(enc.valbuf) == 0
}

func (enc *encoder) AddBytes(buf []byte) {
	enc.valbuf = append(enc.valbuf, buf...)
}

func (enc *encoder) addRawPacket(en iEncoder) {
	enc.valbuf = append(enc.valbuf, en.Encode()...)
}

// setTag write tag as seqID
func (enc *encoder) writeTag() {
	if enc.seqID > 0x3F {
		panic("sid should be in [0..0x7F]")
	}
	if enc.isNode {
		enc.seqID = enc.seqID | 0x80
	}
	if enc.isArray {
		enc.seqID = enc.seqID | 0x40
	}
	enc.buf.WriteByte(enc.seqID)
}

func (enc *encoder) writeLengthBuf() {
	// vallen := enc.valBuf.Len()
	vallen := len(enc.valbuf)
	if vallen < 0 {
		panic("length must greater than 0")
	}

	size := encoding.SizeOfPVarInt32(int32(vallen))
	codec := encoding.VarCodec{Size: size}
	tmp := make([]byte, size)
	err := codec.EncodePVarInt32(tmp, int32(vallen))
	if err != nil {
		panic(err)
	}
	enc.buf.Write(tmp)
}

// Encode returns a final Y3 encoded byte slice
func (enc *encoder) Encode() []byte {
	if !enc.complete {
		// Tag
		enc.writeTag()
		// Len
		enc.writeLengthBuf()
		// Val
		enc.buf.Write(enc.valbuf)
		enc.complete = true
	}
	return enc.buf.Bytes()
}
