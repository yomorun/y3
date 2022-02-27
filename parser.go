package y3

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/yomorun/y3/encoding"
)

var (
	ErrMalformed = errors.New("y3.ReadPacket: malformed")
)

// ReadPacket will try to read a Y3 encoded packet from the reader
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

	// read y3.Length bytes, a varint format
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
	var length int32
	codec := encoding.VarCodec{}
	err = codec.DecodePVarInt32(lenbuf.Bytes(), &length)
	if err != nil {
		return nil, ErrMalformed
	}

	// validate len decoded from stream
	if length < 0 {
		return nil, fmt.Errorf("y3.ReadPacket() get lenbuf=(%# x), decode len=(%v)", lenbuf.Bytes(), length)
	}

	// write y3.Length bytes
	buf.Write(lenbuf.Bytes())

	// read next {len} bytes as y3.Value
	valbuf := bytes.Buffer{}

	// every batch read 512 bytes, if next reads < 512, read
	var count int
	for {
		batchReadSize := 1024 * 1024
		var tmpbuf = []byte{}
		if int(length)-count < batchReadSize {
			tmpbuf = make([]byte, int(length)-count)
		} else {
			tmpbuf = make([]byte, batchReadSize)
		}
		p, err := reader.Read(tmpbuf)
		count += p
		if err != nil {
			if err == io.EOF {
				valbuf.Write(tmpbuf[:p])
				break
			}
			return nil, fmt.Errorf("y3 parse valbuf error: %v", err)
		}
		valbuf.Write(tmpbuf[:p])
		if count == int(length) {
			break
		}
	}

	if count < int(length) {
		// return nil, fmt.Errorf("[y3] p should == len when getting y3 value buffer, len=%d, p=%d", length, count)
		return nil, ErrMalformed
	}
	// write y3.Value bytes
	buf.Write(valbuf.Bytes())

	return buf.Bytes(), nil
}

func readByte(reader io.Reader) (byte, error) {
	var b [1]byte
	_, err := reader.Read(b[:])
	return b[0], err
}
