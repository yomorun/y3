package spec

import (
	"errors"
	"io"
)

const (
	maxSeqID     = 0x3F
	flagBitNode  = 0x80
	wipeFlagBits = 0x3F
	msb          = 0x80
)

var (
	errInvalidSeqID = errors.New("y3.Builder: SeqID should >= 0 and =< 0x3F")
)

func readByte(reader io.Reader) (byte, error) {
	var b [1]byte
	n, err := reader.Read(b[:])
	if n == 0 {
		return 0x00, err
	}
	return b[0], err
}
