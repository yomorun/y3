package y3

import (
	"bytes"
	"errors"
	"io"

	"github.com/yomorun/y3/encoding"
)

func ReadPacket(reader io.Reader) ([]byte, error) {
	tag, err := readByte(reader)
	if err != nil {
		return nil, err
	}
	// buf will contain a complete y3 encoded handshakeFrame
	buf := bytes.Buffer{}

	// the first byte is y3.Tag
	// write y3.Tag bytes
	buf.WriteByte(tag)

	// read y3.Length bytes, a varint format.
	lenbuf := bytes.Buffer{}
	for {
		b, err := readByte(reader)
		if err != nil {
			return nil, err
		}
		lenbuf.WriteByte(b)
		if b&0x80 != 0x80 {
			break
		}
	}
	// parse to y3.Length
	var len int32
	codec := encoding.VarCodec{}
	err = codec.DecodePVarInt32(lenbuf.Bytes(), &len)
	if err != nil {
		return nil, err
	}
	// write y3.Length bytes
	buf.Write(lenbuf.Bytes())

	// read next {len} bytes as y3.Value
	valbuf := make([]byte, len)
	p, err := reader.Read(valbuf)
	if err != nil {
		return nil, err
	}
	if p < int(len) {
		return nil, errors.New("[y3] p should == len when getting y3 value buffer")
	}
	// write y3.Value bytes
	buf.Write(valbuf)

	return buf.Bytes(), nil
}

func readByte(reader io.Reader) (byte, error) {
	var b [1]byte
	_, err := reader.Read(b[:])
	return b[0], err
}
