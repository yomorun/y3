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
	var tid y3.Builder
	tid.SetSeqID(int(TagOfTransactionID), false)
	tid.AddValBytes([]byte(m.transactionID))

	pktTransaction, err := tid.Packet()
	if err != nil {
		return nil, err
	}

	var meta y3.Builder
	meta.SetSeqID(int(TagOfMetaFrame), true)
	meta.AddPacket(pktTransaction)

	return meta.Packet()
}

// DecodeToMetaFrame decodes Y3 encoded bytes to a MetaFrame
func DecodeToMetaFrame(buf []byte) (*MetaFrame, error) {
	packet := &y3.NodePacket{}
	_, err := y3.DecodeToNodePacket(buf, packet)

	if err != nil {
		return nil, err
	}

	var tid string
	if s, ok := packet.PrimitivePackets[0x01]; ok {
		tid, err = s.ToUTF8String()
		if err != nil {
			return nil, err
		}
	}

	meta := &MetaFrame{
		transactionID: tid,
	}
	return meta, nil
}

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

// Build returns a y3 Packet
func (cp *ChunkedPayloadFrame) Build() (y3.Packet, error) {
	var cary y3.Builder
	cary.SetSeqID(int(cp.sid), false)
	cary.SetValReader(cp.carriageReader, cp.carriageSize)

	pktCarriage, err := cary.Packet()
	if err != nil {
		return nil, err
	}

	var pl y3.Builder
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
	var cary y3.Builder
	cary.SetSeqID(int(m.Sid), false)
	cary.AddValBytes(m.Carriage)

	pktCarriage, err := cary.Packet()
	if err != nil {
		return nil, err
	}

	var pl y3.Builder
	pl.SetSeqID(int(TagOfPayloadFrame), true)
	pl.AddPacket(pktCarriage)
	return pl.Packet()
}

// DecodeToPayloadFrame decodes Y3 encoded bytes to PayloadFrame
func DecodeToPayloadFrame(buf []byte) (*PayloadFrame, error) {
	nodeBlock := y3.NodePacket{}
	_, err := y3.DecodeToNodePacket(buf, &nodeBlock)
	if err != nil {
		return nil, err
	}

	payload := &PayloadFrame{}
	for _, v := range nodeBlock.PrimitivePackets {
		payload.Sid = v.SeqID()
		payload.Carriage = v.GetValBuf()
		break
	}

	return payload, nil
}

// DataFrame defines the data structure carried with user's data
// when transfering within YoMo
type DataFrame struct {
	metaFrame           *MetaFrame
	payloadFrame        *PayloadFrame
	chunkedPayloadFrame *ChunkedPayloadFrame
	isChunked           bool
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

	var b y3.Builder
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
	return d.payloadFrame.Carriage
}

// TransactionID return transactionID string
func (d *DataFrame) TransactionID() string {
	return d.metaFrame.TransactionID()
}

// GetDataTagID return the Tag of user's data
func (d *DataFrame) CarriageSeqID() byte {
	return d.payloadFrame.Sid
}

// Encode return Y3 encoded bytes of `DataFrame`
func (d *DataFrame) Encode() []byte {
	panic("not implemented")
}

// DecodeToDataFrame decode Y3 encoded bytes to `DataFrame`
func DecodeToDataFrame(buf []byte) (*DataFrame, error) {
	packet := y3.NodePacket{}
	_, err := y3.DecodeToNodePacket(buf, &packet)
	if err != nil {
		return nil, err
	}

	data := &DataFrame{}

	if metaBlock, ok := packet.NodePackets[byte(TagOfMetaFrame)]; ok {
		meta, err := DecodeToMetaFrame(metaBlock.GetRawBytes())
		if err != nil {
			return nil, err
		}
		data.metaFrame = meta
	}

	if payloadBlock, ok := packet.NodePackets[byte(TagOfPayloadFrame)]; ok {
		payload, err := DecodeToPayloadFrame(payloadBlock.GetRawBytes())
		if err != nil {
			return nil, err
		}
		data.payloadFrame = payload
	}

	return data, nil
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
	fmt.Printf("(source stream)==>flush:[%# x]\n", o.buf[o.lastRead])
	copy(buf, []byte{o.buf[o.lastRead]})
	o.lastRead++
	return 1, nil
}

func main() {
	log.Println(">>> Start: Emit data ---")
	emit()

	log.Println(">>> Start: Emit data in Chunked Mode ---")
	emitInChunkedMode()

	fmt.Println("OVER")
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
	log.Printf("DONE, total buf=[%# x]\n\n", data.Raw())
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
