package y3

import (
	"bytes"
	"fmt"
)

// Encoder will encode object to Y3 encoding
type encoder struct {
	seqID   int
	valbuf  []byte
	isNode  bool
	isArray bool
	buf     *bytes.Buffer
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

func (enc *encoder) String() string {
	return fmt.Sprintf("Encoder: isNode=%v | seqID=%#x | valBuf=%#v | buf=%#v", enc.isNode, enc.seqID, enc.valbuf, enc.buf)
}
