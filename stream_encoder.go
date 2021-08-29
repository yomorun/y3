package y3

import (
	"bytes"
	"io"

	"github.com/yomorun/y3/encoding"
)

type StreamEncoder struct {
	tag    byte
	buf    *bytes.Buffer
	pbuf   *bytes.Buffer
	len    int
	slen   int
	reader *yR
}

func NewStreamEncoder(tag byte) *StreamEncoder {
	var se = &StreamEncoder{
		tag:  tag,
		buf:  new(bytes.Buffer),
		pbuf: new(bytes.Buffer),
	}

	return se
}

func (se *StreamEncoder) AddPacket(packet *PrimitivePacketEncoder) {
	node := packet.Encode()
	se.AddPacketBuffer(node)
}

func (se *StreamEncoder) AddPacketBuffer(buf []byte) {
	se.pbuf.Write(buf)
	se.growLen(len(se.pbuf.Bytes()))
}

func (se *StreamEncoder) AddStreamPacket(tag byte, length int, reader io.Reader) {
	se.slen = length
	// s-Tag
	se.pbuf.WriteByte(tag)
	se.growLen(1)
	// calculate s-Len
	size := encoding.SizeOfPVarInt32(int32(length))
	codec := encoding.VarCodec{Size: size}
	tmp := make([]byte, size)
	err := codec.EncodePVarInt32(tmp, int32(length))
	if err != nil {
		panic(err)
	}
	se.pbuf.Write(tmp)
	se.growLen(size)

	// total buf
	se.buf.WriteByte(se.tag)
	se.growLen(length)
	// calculate total Len buf
	size = encoding.SizeOfPVarInt32(int32(se.len))
	codec = encoding.VarCodec{Size: size}
	tmp = make([]byte, size)
	err = codec.EncodePVarInt32(tmp, int32(se.len))
	if err != nil {
		panic(err)
	}
	se.buf.Write(tmp) //lenbuf
	se.buf.Write(se.pbuf.Bytes())
	se.growLen(size) // total length
	se.growLen(1)    // parent tag

	se.reader = &yR{
		buf:    se.buf,
		src:    reader,
		length: se.len,
		slen:   se.slen,
	}
}

func (se *StreamEncoder) GetReader() io.Reader {
	if se.reader != nil {
		return se.reader
	}
	return new(bytes.Buffer)
}

// Pipe can pipe data to os.StdOut
func (se *StreamEncoder) Pipe(writer io.Writer) {

}

func (se *StreamEncoder) GetLen() int {
	if se.reader != nil {
		return se.len
	}
	return 0
}

func (se *StreamEncoder) growLen(step int) {
	se.len += step
}

type yR struct {
	src    io.Reader
	buf    *bytes.Buffer
	length int
	off    int
	slen   int
}

func (r *yR) Read(p []byte) (n int, err error) {
	if r.src == nil {
		return 0, nil
	}

	if r.off >= r.length {
		return 0, io.EOF
	}

	if r.off < r.length-r.slen {
		n, err := r.buf.Read(p)
		r.off += n
		if err != nil {
			if err == io.EOF {
				return n, nil
			} else {
				return 0, err
			}
		}
		return n, nil
	} else {
		n, err := r.src.Read(p)
		r.off += n
		if err != nil {
			return n, err
		}
		return n, nil
	}
}
