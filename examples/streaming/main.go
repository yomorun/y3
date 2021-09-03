package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/yomorun/y3"
)

// Frame defines frames
type Frame interface {
	Encode() []byte
}

func ReadChunkedDataFrame(d *y3.Decoder) (*DataFrame, error) {
	// 1. decoding MetaFrame
	metaFrame, err := ReadMetaFrame(d)
	if err != nil {
		return nil, err
	}
	// 2. decoding PayloadFrame in chunked mode
	payloadFrame, err := ReadChunkedPayloadFrame(d)
	if err != nil {
		return nil, err
	}

	// d.SetChunkedDataReader(payloadFrame.CarriageReader())
	p := d.GetChunkedPacket()

	return &DataFrame{
		metaFrame:           metaFrame,
		isChunked:           true,
		chunkedPayloadFrame: payloadFrame,
		rd:                  p.Reader(),
		chunkedSize:         payloadFrame.carriageSize,
	}, nil
}

// decoding PayloadFrame in chunked mode
func ReadChunkedPayloadFrame(dec *y3.Decoder) (*ChunkedPayloadFrame, error) {
	d := y3.NewDecoder(dec.UnderlyingReader())
	err := d.ReadHeader()
	if err != nil {
		return nil, err
	}
	cplPacket := d.GetChunkedPacket()

	// decoding carriage of PayloadFrame
	caryDec := y3.NewDecoder(cplPacket.VReader())
	err = caryDec.ReadHeader()
	if err != nil {
		return nil, err
	}
	caryPacket := caryDec.GetChunkedPacket()
	return &ChunkedPayloadFrame{
		sid:            byte(caryPacket.SeqID()),
		carriageSize:   caryPacket.VSize(),
		carriageReader: caryPacket.VReader(),
	}, nil
}

func ReadDataFrame(d *y3.Decoder) (*DataFrame, error) {
	// 1. decoding MetaFrame
	metaFrame, err := ReadMetaFrame(d)
	if err != nil {
		return nil, err
	}
	// 2. decoding PayloadFrame in fullfilled mode
	payloadFrame, err := ReadPayloadFrame(d)
	if err != nil {
		return nil, err
	}

	return &DataFrame{
		metaFrame:    metaFrame,
		isChunked:    false,
		payloadFrame: payloadFrame,
	}, nil
}

func ReadPayloadFrame(dec *y3.Decoder) (*PayloadFrame, error) {
	d := y3.NewDecoder(dec.UnderlyingReader())
	err := d.ReadHeader()
	if err != nil {
		return nil, err
	}
	plPacket, err := d.GetFullfilledPacket()
	if err != nil {
		return nil, err
	}

	// decode Carriage from this packet
	var caryDec = y3.NewDecoder(plPacket.VReader())
	err = caryDec.ReadHeader()
	if err != nil {
		return nil, err
	}
	caryPacket, err := caryDec.GetFullfilledPacket()
	if err != nil {
		return nil, err
	}

	return &PayloadFrame{
		Sid:      byte(caryPacket.SeqID()),
		Carriage: caryPacket.BytesV(),
	}, nil
}

func ReadMetaFrame(dec *y3.Decoder) (*MetaFrame, error) {
	d := y3.NewDecoder(dec.UnderlyingReader())
	err := d.ReadHeader()
	if err != nil {
		return nil, err
	}
	metaPacket, err := d.GetFullfilledPacket()
	if err != nil {
		return nil, err
	}

	// decode Transaction from this packet
	var tidDec = y3.NewDecoder(metaPacket.VReader())
	err = tidDec.ReadHeader()
	if err != nil {
		return nil, err
	}
	tidPacket, err := tidDec.GetFullfilledPacket()
	if err != nil {
		return nil, err
	}

	meta := &MetaFrame{
		transactionID: tidPacket.UTF8StringV(),
	}
	return meta, nil
}

var (
	TagOfDataFrame     byte = 0x3F
	TagOfMetaFrame     byte = 0x2F
	TagOfPayloadFrame  byte = 0x2E
	TagOfTransactionID byte = 0x01
)

// MetaFrame defines the Meta data structure in a `DataFrame`, transactionID is
type MetaFrame struct {
	transactionID string
}

var _ Frame = &MetaFrame{}

// NewMetaFrame creates a new MetaFrame with a given transactionID
func NewMetaFrame(tid string) *MetaFrame {
	return &MetaFrame{
		transactionID: tid,
	}
}

// TransactionID returns the transactionID of the MetaFrame
func (m *MetaFrame) TransactionID() string {
	return m.transactionID
}

// Encode returns Y3 encoded bytes of the MetaFrame
func (m *MetaFrame) Encode() []byte {
	panic("not implemented")
}

// Build returns a Y3 Packet
func (m *MetaFrame) Build() (y3.Packet, error) {
	var tid y3.Encoder
	tid.SetSeqID(int(TagOfTransactionID), false)
	// tid.SetBytesV([]byte(m.transactionID))
	tid.SetUTF8StringV(m.transactionID)

	pktTransaction, err := tid.Packet()
	if err != nil {
		return nil, err
	}

	var meta y3.Encoder
	meta.SetSeqID(int(TagOfMetaFrame), true)
	meta.AddPacket(pktTransaction)

	return meta.Packet()
}

// DecodeToMetaFrame decodes Y3 encoded bytes to a MetaFrame
func DecodeToMetaFrame(r []byte) (*MetaFrame, error) { return nil, nil }

// ChunkedPayloadFrame represents a Payload with chunked carriage data.
type ChunkedPayloadFrame struct {
	sid            byte
	carriageReader io.Reader
	carriageSize   int
}

var _ Frame = &ChunkedPayloadFrame{}

// NewChunkedPayloadFrame create a ChunkedPayloadFrame
func NewChunkedPayloadFrame(seqID byte) *ChunkedPayloadFrame {
	return &ChunkedPayloadFrame{
		sid: seqID,
	}
}

// Encode returns y3 encoded raw bytes
func (cp *ChunkedPayloadFrame) Encode() []byte {
	panic("not implemented")
}

// SetCarriageReader set the V of a y3 packet as a io.Reader, and provide the
// size of V.
func (cp *ChunkedPayloadFrame) SetCarriageReader(r io.Reader, size int) {
	cp.carriageReader = r
	cp.carriageSize = size
}

// CarriageReader returns the V of y3 packet as io.Reader
func (cp *ChunkedPayloadFrame) CarriageReader() io.Reader {
	return cp.carriageReader
}

// Build returns a y3 Packet
func (cp *ChunkedPayloadFrame) Build() (y3.Packet, error) {
	var cary y3.Encoder
	cary.SetSeqID(int(cp.sid), false)
	cary.SetReaderV(cp.carriageReader, cp.carriageSize)

	pktCarriage, err := cary.Packet()
	if err != nil {
		return nil, err
	}

	var pl y3.Encoder
	pl.SetSeqID(int(TagOfPayloadFrame), true)
	pl.AddStreamPacket(pktCarriage)
	return pl.Packet()
}

// IsChunked returns a bool value indicates if this Frame is chunked.
func (cp *ChunkedPayloadFrame) IsChunked() bool {
	return true
}

// PayloadFrame represents Payload in Y3 encoded bytes, seqID is a fixed value
// with TYPE_ID_PAYLOAD_FRAME, when carriage is small, this Frame is not memory
// efficiency but easy for use.
type PayloadFrame struct {
	Sid      byte
	Carriage []byte
}

var _ Frame = &PayloadFrame{}

// NewPayloadFrame creates a new PayloadFrame with a given TagID of user's data
func NewPayloadFrame(seqID byte) *PayloadFrame {
	return &PayloadFrame{
		Sid: seqID,
	}
}

// SetCarriage sets the user's raw data
func (m *PayloadFrame) SetCarriage(buf []byte) {
	m.Carriage = buf
}

// Encode to Y3 encoded bytes
func (m *PayloadFrame) Encode() []byte {
	panic("not implemented")
}

func (m *PayloadFrame) Build() (y3.Packet, error) {
	var cary y3.Encoder
	cary.SetSeqID(int(m.Sid), false)
	cary.SetBytesV(m.Carriage)

	pktCarriage, err := cary.Packet()
	if err != nil {
		return nil, err
	}

	var pl y3.Encoder
	pl.SetSeqID(int(TagOfPayloadFrame), true)
	pl.AddPacket(pktCarriage)
	return pl.Packet()
}

// DataFrame defines the data structure carried with user's data
// when transfering within YoMo
type DataFrame struct {
	metaFrame           *MetaFrame
	payloadFrame        *PayloadFrame
	chunkedPayloadFrame *ChunkedPayloadFrame
	isChunked           bool
	rd                  io.Reader
	chunkedSize         int
}

var _ Frame = &DataFrame{}

// NewDataFrame create `DataFrame` with a transactionID string,
// consider change transactionID to UUID type later
func NewDataFrame(transactionID string) *DataFrame {
	data := &DataFrame{
		metaFrame: NewMetaFrame(transactionID),
	}
	return data
}

// Type gets the type of Frame.
func (d *DataFrame) Type() byte {
	return TagOfDataFrame
}

// SetCarriage set user's raw data in `DataFrame`
func (d *DataFrame) SetCarriage(sid byte, carriage []byte) {
	d.payloadFrame = NewPayloadFrame(sid)
	d.payloadFrame.SetCarriage(carriage)
	d.isChunked = false
}

func (d *DataFrame) SetCarriageReader(sid byte, r io.Reader, size int) {
	d.chunkedPayloadFrame = NewChunkedPayloadFrame(sid)
	d.chunkedPayloadFrame.SetCarriageReader(r, size)
	d.isChunked = true
}

func (d *DataFrame) Build() (y3.Packet, error) {
	meta, err := d.metaFrame.Build()
	if err != nil {
		return nil, err
	}

	var payload y3.Packet
	if d.isChunked {
		payload, err = d.chunkedPayloadFrame.Build()
	} else {
		payload, err = d.payloadFrame.Build()
	}

	if err != nil {
		return nil, err
	}

	var b y3.Encoder
	b.SetSeqID(int(TagOfDataFrame), true)
	b.AddPacket(meta)
	if d.isChunked {
		b.AddStreamPacket(payload)
	} else {
		b.AddPacket(payload)
	}

	return b.Packet()
}

// GetCarriage return user's raw data in `DataFrame`
func (d *DataFrame) Carriage() []byte {
	if d.isChunked {
		panic("error")
	}
	return d.payloadFrame.Carriage
}

// CarriageReader return an io.Reader as user data
func (d *DataFrame) Reader() io.Reader {
	return d.rd
}

func (d *DataFrame) ChunkedSize() int {
	return d.chunkedSize
}

// TransactionID return transactionID string
func (d *DataFrame) TransactionID() string {
	return d.metaFrame.TransactionID()
}

// GetDataTagID return the Tag of user's data
func (d *DataFrame) CarriageSeqID() byte {
	if d.isChunked {
		return d.chunkedPayloadFrame.sid
	}
	return d.payloadFrame.Sid
}

// Encode return Y3 encoded bytes of `DataFrame`
func (d *DataFrame) Encode() []byte {
	panic("not implemented")
}

// Test process
type p struct {
	buf      []byte
	lastRead int
}

func (o *p) Read(buf []byte) (int, error) {
	if o.lastRead >= len(o.buf) {
		return 0, io.EOF
	}
	time.Sleep(100 * time.Millisecond)
	fmt.Printf("(source stream)==>flush:[%#x]\n", o.buf[o.lastRead])
	copy(buf, []byte{o.buf[o.lastRead]})
	o.lastRead++
	return 1, nil
}

func ReadFrame(stream io.Reader) error {
	dec := y3.NewDecoder(stream)
	// read T at first, then will know the seqID of current packet
	err := dec.ReadHeader()
	if err != nil {
		return err
	}

	switch dec.SeqID() {
	case int(TagOfDataFrame):
		d, err := ReadDataFrame(dec)
		if err != nil {
			return err
		}
		log.Printf("data-m=%v", d.metaFrame)
		log.Printf("data-p=%v", d.payloadFrame)
		log.Printf("R-data=tid:%s, csid=%# x, isStreamMode=%v, cary=[%# x]", d.TransactionID(), d.CarriageSeqID(), d.isChunked, d.Carriage())
	default:
		panic("unknow packet")
	}

	return err
}

func ReadFrameInChunkedMode(stream io.Reader) error {
	dec := y3.NewDecoder(stream)
	// read T at first, then will know the seqID of current packet
	err := dec.ReadHeader()
	if err != nil {
		return err
	}

	switch dec.SeqID() {
	case int(TagOfDataFrame):
		d, err := ReadChunkedDataFrame(dec)
		if err != nil {
			return err
		}
		log.Printf("data-m=%v", d.metaFrame)
		log.Printf("data-p=%v", d.chunkedPayloadFrame)
		log.Printf("R-data=tid:%s, csid=%# x, isStreamMode=%v, chunkedSize=%d", d.TransactionID(), d.CarriageSeqID(), d.isChunked, d.ChunkedSize())
		// operate the reader
		r := d.Reader()
		buf := make([]byte, d.ChunkedSize())
		for {
			n, err := r.Read(buf)
			if n >= 0 || err == io.EOF {
				log.Printf("data=%# x", buf[:n])
				break
			}
			if err != nil {
				panic(err)
			}
		}
	default:
		panic("unknow packet")
	}

	return err
}

func main() {
	log.Println(">>> Start: Receive data in Chunked Mode---")
	recvInChunkedMode()
	// return

	log.Println(">>> Start: Receive data ---")
	recv()

	log.Println(">>> Start: Emit data ---")
	emit()

	log.Println(">>> Start: Emit data in Chunked Mode ---")
	emitInChunkedMode()

	fmt.Println("OVER")
}

func recvInChunkedMode() {
	data := []byte{
		TagOfDataFrame | 0x80, 0x0D,
		TagOfMetaFrame | 0x80, 0x06, TagOfTransactionID, 0x04, 0x79, 0x6f, 0x6d, 0x6f,
		TagOfPayloadFrame | 0x80, 0x04, 0x09, 0x02, 0xFF, 0xFE,
		// TagOfDataFrame | 0x80, 0x0C,
		// TagOfMetaFrame | 0x80, 0x06, TagOfTransactionID, 0x04, 0x6f, 0x6f, 0x6f, 0x6f,
		// TagOfPayloadFrame | 0x80, 0x03, 0x09, 0x01, 0x01,
	}
	stream := &p{buf: data}
	// decode
	for {
		err := ReadFrameInChunkedMode(stream)
		if err != nil {
			if err == io.EOF {
				log.Printf("DONE recv")
				break
			}
			panic(err)
		}
	}
}

func recv() {
	data := []byte{
		TagOfDataFrame | 0x80, 0x0D,
		TagOfMetaFrame | 0x80, 0x06, TagOfTransactionID, 0x04, 0x79, 0x6f, 0x6d, 0x6f,
		TagOfPayloadFrame | 0x80, 0x04, 0x09, 0x02, 0xFF, 0xFE,
		TagOfDataFrame | 0x80, 0x0C,
		TagOfMetaFrame | 0x80, 0x06, TagOfTransactionID, 0x04, 0x6f, 0x6f, 0x6f, 0x6f,
		TagOfPayloadFrame | 0x80, 0x03, 0x09, 0x01, 0x01,
	}
	stream := &p{buf: data}
	// decode
	for {
		err := ReadFrame(stream)
		if err != nil {
			if err == io.EOF {
				log.Printf("DONE recv")
				break
			}
			panic(err)
		}
	}
}

func emit() {
	payloadData := []byte{0x01, 0x02, 0x03, 0x04, 0x05}

	transactionID := "yomo"
	var dataSeqID byte = 0x30

	// Prepare DataFrame
	df := NewDataFrame(transactionID)
	df.SetCarriage(dataSeqID, payloadData)
	data, err := df.Build()
	if err != nil {
		panic(err)
	}
	log.Printf("DONE, total buf=[%# x]\n\n", data.Bytes())
}

func emitInChunkedMode() {
	payloadData := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	payloadReader := &p{buf: payloadData}

	transactionID := "yomo"
	var dataSeqID byte = 0x30

	// Prepare DataFrame
	df := NewDataFrame(transactionID)
	// df.SetCarriage(dataSeqID, payloadData)
	df.SetCarriageReader(dataSeqID, payloadReader, len(payloadData))
	data, err := df.Build()
	if err != nil {
		panic(err)
	}

	buf := &bytes.Buffer{}
	io.Copy(buf, data.Reader())
	log.Printf("DONE, total buf=[%# x]", buf)
}
