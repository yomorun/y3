package y3

import (
	"bytes"
	"io"

	"github.com/yomorun/y3/spec"
)

// Encoder is the tool for creating a y3 packet easily
type Decoder struct {
	tag spec.T
	len *spec.L
	rd  io.Reader
	vr  io.Reader
}

func NewDecoder(reader io.Reader) *Decoder {
	return &Decoder{
		rd: reader,
	}
}

// SeqID return the SequenceID of the decoding packet
func (d *Decoder) SeqID() int {
	return d.tag.Sid()
}

// UnderlyingReader returns the reader this decoder using
func (d *Decoder) UnderlyingReader() io.Reader {
	return d.rd
}

// // SetChunkedDataReader set chunked io.Reader
// func (d *Decoder) SetChunkedDataReader(r io.Reader) {
// 	d.vr = r
// }

// // ChunkedDataReader return chunked io.Reader
// func (d *Decoder) ChunkedDataReader() io.Reader {
// 	return d.vr
// }

// ReadHeader will block until io.EOF or recieve T and L of a packet.
func (d *Decoder) ReadHeader() error {
	// only read T and L
	return d.readTL()
}

// GetChunkedPacket will block until io.EOF or recieve V of a packet in chunked mode.
func (d *Decoder) GetChunkedPacket() Packet {
	return &StreamPacket{
		t:         d.tag,
		l:         *d.len,
		vr:        d.rd,
		chunkMode: true,
		chunkSize: d.len.VSize(),
	}
}

// GetFullfilledPacket read full Packet from given io.Reader
func (d *Decoder) GetFullfilledPacket() (packet Packet, err error) {
	// read V
	buf := new(bytes.Buffer)
	total := 0
	for {
		valbuf := make([]byte, d.len.VSize())
		n, err := d.rd.Read(valbuf)
		if n > 0 {
			total += n
			buf.Write(valbuf[:n])
		}
		if total >= d.len.VSize() || err != nil {
			break
		}
	}

	packet = &StreamPacket{
		t:         d.tag,
		l:         *d.len,
		vbuf:      buf.Bytes(),
		chunkMode: false,
	}

	return packet, nil
}

func (d *Decoder) readTL() (err error) {
	if d.rd == nil {
		return errNilReader
	}

	// read T
	d.tag, err = spec.ReadT(d.rd)
	if err != nil {
		return err
	}

	// read L
	d.len, err = spec.ReadL(d.rd)

	return err
}
