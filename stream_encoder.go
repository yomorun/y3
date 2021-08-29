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
	return &StreamEncoder{
		tag:  tag,
		buf:  new(bytes.Buffer),
		pbuf: new(bytes.Buffer),
	}
}

func (se *StreamEncoder) AddPacket(packet *PrimitivePacketEncoder) {
	node := packet.Encode()
	se.pbuf.Write(node)
}

func (se *StreamEncoder) AddStreamPacket(tag byte, length int, reader io.Reader) {
	se.slen = length
	se.pbuf.WriteByte(tag)
	// calculate Len
	size := encoding.SizeOfPVarInt32(int32(length))
	codec := encoding.VarCodec{Size: size}
	tmp := make([]byte, size)
	err := codec.EncodePVarInt32(tmp, int32(length))
	if err != nil {
		panic(err)
	}
	se.pbuf.Write(tmp)

	// total buf
	se.buf.WriteByte(se.tag)
	// se.len += 1
	se.len += len(se.pbuf.Bytes()) + length
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
	se.len += size
	se.len += 1

	se.reader = &yR{
		buf:    se.buf,
		src:    reader,
		length: se.len,
		slen:   se.slen,
	}
}

func (se *StreamEncoder) GetReader() io.Reader {
	return se.reader
}

// Pipe can pipe data to os.StdOut
func (se *StreamEncoder) Pipe(writer io.Writer) {

}

func (se *StreamEncoder) GetLen() int {
	return se.len
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
