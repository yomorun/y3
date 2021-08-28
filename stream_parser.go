package y3

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/yomorun/y3/encoding"
)

// streamReader read an Y3 packet from a io.Reader, and return
// the ValReader after decode out Tag and Len
type streamReader struct {
	src   io.Reader
	chTag byte
	chLen int
	chVal io.Reader
}

// NewStreamReader create a new y3 streamReader
func NewStreamParser(reader io.Reader) *streamReader {
	return &streamReader{
		src: reader,
	}
}

// GetTag will block until Tag arrived
func (sr *streamReader) GetTag() byte {
	return sr.chTag
}

// GenLen will block until Len arrived
func (sr *streamReader) GetLen() int {
	return sr.chLen
}

func (sr *streamReader) GetValReader() io.Reader {
	return sr.chVal
}

// Do must run in a goroutine
func (sr *streamReader) Do() error {
	if sr.src == nil {
		return errors.New("y3: nil source reader")
	}

	tag, err := readByte(sr.src)
	if err != nil {
		// TODO: determine io.EOF
		if err == io.EOF {

		}
		return err
	}

	// the first byte is y3.Tag
	sr.chTag = tag

	// read y3.Length bytes, a varint format
	lenbuf := bytes.Buffer{}
	for {
		b, err := readByte(sr.src)
		if err != nil {
			return err
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
		return err
	}

	// validate len decoded from stream
	if length < 0 {
		return fmt.Errorf("y3: streamParse() get lenbuf=(%# x), decode len=(%v)", lenbuf.Bytes(), length)
	}

	sr.chLen = int(length)

	// read next {len} bytes as y3.Value
	sr.chVal = &valR{
		length: int(length),
		src:    sr.src,
	}

	return nil
}

type valR struct {
	length int
	off    int
	src    io.Reader
}

func (r *valR) Read(p []byte) (n int, err error) {
	if r.src == nil {
		return 0, nil
	}

	if r.off >= r.length {
		return 0, io.EOF
	}

	bound := len(p)
	if len(p) > r.length-r.off {
		bound = r.length - r.off
	}
	// update readed
	r.off, err = r.src.Read(p[0:bound])
	return r.off, err
}

func StreamReadPacket(reader io.Reader) (*streamReader, error) {
	sp := NewStreamParser(reader)
	err := sp.Do()
	return sp, err
}
