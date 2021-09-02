package spec

import (
	"io"
)

// T is the Tag in a TLV structure
type T byte

// NewT returns a T with sequenceID. If this packet contains other
// packets, this packet will be a "node packet", the T of this packet
// will set MSB to T.
func NewT(seqID int) (T, error) {
	if seqID < 0 || seqID > maxSeqID {
		return 0, errInvalidSeqID
	}

	return T(seqID), nil
}

// Sid returns the sequenceID of this packet.
func (t T) Sid() int {
	return int(t & wipeFlagBits)
}

// Bytes returns raw bytes of T.
func (t T) Bytes() []byte {
	return []byte{byte(t)}
}

// IsNode will return true if this packet contains other packets.
// Otherwise return flase.
func (t T) IsNodeMode() bool {
	return t&flagBitNode == flagBitNode
}

// SetIsNodeMode will set T to indicates this packet contains
// other packets.
func (t T) SetNodeMode(flag bool) {
	if flag {
		t |= flagBitNode
	}
}

// Size return the size of T raw bytes.
func (t T) Size() int {
	return 1
}

// ReadT read T from a bufio.Reader
func ReadT(rd io.Reader) (T, error) {
	b, err := readByte(rd)
	return T(b), err
}
