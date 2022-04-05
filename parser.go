package y3

import (
	"bytes"
	"errors"
	"io"

	"github.com/yomorun/y3/encoding"
)

var (
	ErrMalformed = errors.New("y3.ReadPacket: malformed")
)

// ReadPacket will try to read a Y3 encoded packet from the reader
func ReadPacket(reader io.Reader) ([]byte, error) {
	// the first byte is y3.Tag
	tag, err := readByte(reader)
	if err != nil {
		if err == io.EOF {
			return nil, ErrMalformed
		}
		return nil, err
	}

	// buf will hold this packet
	buf := bytes.Buffer{}

	// write y3.Tag
	buf.WriteByte(tag)

	// read y3.Length bytes, a varint format
	lenbuf := bytes.Buffer{}
	for {
		b, err := readByte(reader)
		if err != nil {
			if err == io.EOF {
				return nil, ErrMalformed
			}
			return nil, err
		}
		lenbuf.WriteByte(b)
		// if the last byte is not 0x80, it is the last byte of the length
		if b&0x80 != 0x80 {
			break
		}
	}

	// parse to y3.Length
	var length int32
	codec := encoding.VarCodec{}
	err = codec.DecodePVarInt32(lenbuf.Bytes(), &length)
	if err != nil {
		return nil, ErrMalformed
	}

	// validate len decoded from stream
	if length < 0 {
		return nil, ErrMalformed
	}

	// write y3.Length bytes
	buf.Write(lenbuf.Bytes())

	// read y3.Val bytes
	var valbuf = make([]byte, length)
	m, err := io.ReadFull(reader, valbuf)
	if err != nil {
		return nil, ErrMalformed
	}

	if m != int(length) {
		return nil, ErrMalformed
	}

	buf.Write(valbuf)

	return buf.Bytes(), nil
}

func readByte(reader io.Reader) (byte, error) {
	var b [1]byte
	_, err := reader.Read(b[:])
	return b[0], err
}
