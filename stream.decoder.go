package y3

import (
	"bytes"
	"io"

	"github.com/yomorun/y3/spec"
)

// Decoder is the tool for decoding y3 packet from stream
type Decoder struct {
	tag spec.T
	len *spec.L
	rd  io.Reader
}

// NewDecoder returns a Decoder from an io.Reader
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
