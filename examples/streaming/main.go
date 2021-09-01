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

type ChunkPayloadFrame struct {
	sid            byte
	carriageReader io.Reader
	carriageSize   int
}

var _ Frame = &ChunkPayloadFrame{}

func NewChunkPayloadFrame(seqID byte) *ChunkPayloadFrame {
	return &ChunkPayloadFrame{
		sid: seqID,
	}
}

func (cp *ChunkPayloadFrame) Encode() []byte {
	return nil
}

func (cp *ChunkPayloadFrame) SetCarriageReader(r io.Reader, size int) {
	cp.carriageReader = r
	cp.carriageSize = size
}

func (cp *ChunkPayloadFrame) Build() (y3.Packet, error) {
	var cary y3.Builder
	cary.SetSeqID(int(cp.sid), false)
	cary.SetValReader(cp.carriageReader, cp.carriageSize)

	pktCarriage, err := cary.Packet()
	if err != nil {
		return nil, err
	}

	var pl y3.Builder
	pl.SetSeqID(int(TagOfPayloadFrame), true)
	pl.AddChunkPacket(pktCarriage)
	return pl.Packet()
}

func (cp *ChunkPayloadFrame) IsChunked() bool {
	return true
}

// PayloadFrame is a Y3 encoded bytes, Tag is a fixed value TYPE_ID_PAYLOAD_FRAME
// the Len is the length of Val. Val is also a Y3 encoded PrimitivePacket, storing
// raw bytes as user's data
type PayloadFrame struct {
	Sid      byte
	Carriage []byte
	reader   io.Reader
	length   int
}

var _ Frame = &PayloadFrame{}

// NewPayloadFrame creates a new PayloadFrame with a given TagID of user's data
func NewPayloadFrame(tag byte) *PayloadFrame {
	return &PayloadFrame{
		Sid: tag,
	}
}

// SetCarriage sets the user's raw data
func (m *PayloadFrame) SetCarriage(buf []byte) {
	m.Carriage = buf
}

// Encode to Y3 encoded bytes
func (m *PayloadFrame) Encode() []byte {
	return nil
}

func (m *PayloadFrame) Build(r io.Reader, size int) (y3.Packet, error) {
	var cary y3.Builder
	cary.SetSeqID(int(m.Sid), false)
	cary.SetValReader(r, size)

	pktCarriage, err := cary.Packet()
	if err != nil {
		return nil, err
	}

	var pl y3.Builder
	pl.SetSeqID(int(TagOfPayloadFrame), true)
	pl.AddChunkPacket(pktCarriage)
	return pl.Packet()
}

// func (m *PayloadFrame) SetLength(length int) {
// 	m.length = length
// }

func (m *PayloadFrame) SetCarriageReader(reader io.Reader, size int) {
	m.reader = reader
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
	metaFrame    *MetaFrame
	payloadFrame *PayloadFrame
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
}

// GetCarriage return user's raw data in `DataFrame`
func (d *DataFrame) GetCarriage() []byte {
	return d.payloadFrame.Carriage
}

// TransactionID return transactionID string
func (d *DataFrame) TransactionID() string {
	return d.metaFrame.TransactionID()
}

// GetDataTagID return the Tag of user's data
func (d *DataFrame) GetDataTagID() byte {
	return d.payloadFrame.Sid
}

// Encode return Y3 encoded bytes of `DataFrame`
func (d *DataFrame) Encode() []byte {
	data := y3.NewNodePacketEncoder(byte(d.Type()))
	// MetaFrame
	data.AddBytes(d.metaFrame.Encode())
	// PayloadFrame
	data.AddBytes(d.payloadFrame.Encode())

	return data.Encode()
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

// NewMetaFrame creates a new MetaFrame with a given transactionID
func NewMetaFrame(tid string) *MetaFrame {
	return &MetaFrame{
		transactionID: tid,
	}
}

var _ Frame = &MetaFrame{}

// MetaFrame defines the data structure of meta data in a `DataFrame`
type MetaFrame struct {
	transactionID string
}

// TransactionID returns the transactionID of the MetaFrame
func (m *MetaFrame) TransactionID() string {
	return m.transactionID
}

// Encode returns Y3 encoded bytes of the MetaFrame
//func (m *MetaFrame) Encode() []byte {
//	metaNode := y3.NewNodePacketEncoder(byte(TagOfMetaFrame))
//	// TransactionID string
//	tidPacket := y3.NewPrimitivePacketEncoder(byte(TagOfTransactionID))
//	tidPacket.SetStringValue(m.transactionID)
//	// add TransactionID to MetaFrame
//	metaNode.AddPrimitivePacket(tidPacket)
//
//	return metaNode.Encode()
//}

// Encode returns Y3 encoded bytes of the MetaFrame
func (m *MetaFrame) Encode() []byte {
	p, err := m.Build()
	if err != nil {
		return nil
	}
	return p.Raw()
}

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
	payloadData := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	payloadReader := &p{buf: payloadData}

	// Prepare a DataFrame
	// DataFrame is combined with a MetaFrame and a PayloadFrame
	// 1. Prepare MetaFrame
	transactionID := "yomo"
	//var tag byte = 0x01
	metaF := NewMetaFrame(transactionID)
	meta, err := metaF.Build()
	if err != nil {
		panic(err)
	}

	log.Printf("meta=[%# x]", meta.Raw())

	plF := NewChunkPayloadFrame(0x22)
	plF.SetCarriageReader(payloadReader, len(payloadData))
	payload, err := plF.Build()
	if err != nil {
		panic(err)
	}

	//log.Printf("buf=%# x, reader=%v", payload.Raw(), payload.Reader())

	/* 	// 2. Prepare PayloadFrame
	   	payload := NewPayloadFrame(tag)
	   	payload.SetLength(len(payloadData))
	   	payload.SetCarriageReader(payloadReader)
	   	// 3. combine to DataFrame
	   	enc := y3.NewStreamEncoder(TagOfDataFrame)
	   	enc.AddPacketBuffer(meta.Raw())
	   	// enc.AddStreamPacket(tag, len(payloadData), payloadReader)
	   	enc.AddStreamPacket(payload.Sid, payload.length, payload.reader)

	   	// try read
	   	fmt.Printf("length=%d\n", enc.GetLen())
	   	r := enc.GetReader()
	*/

	// // method 1: try read all
	// buf, err := ioutil.ReadAll(r)
	// fmt.Printf("err=%v\n", err)
	// fmt.Printf("buf=%# x\n", buf)

	buf := &bytes.Buffer{}
	io.Copy(buf, payload.Reader())
	log.Printf("buf=[%# x]", buf)

	fmt.Println("OVER")
}
