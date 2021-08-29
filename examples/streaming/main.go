package main

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/yomorun/y3"
)

var (
	TagOfDataFrame     byte = 0x3F
	TagOfMetaFrame     byte = 0x2F
	TagOfPayloadFrame  byte = 0x2E
	TagOfTransactionID byte = 0x01
)

func main() {
	payloadData := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	payloadReader := new(bytes.Buffer)
	payloadReader.Write(payloadData)
	// Prepare a DataFrame
	// DataFrame is combined with a MetaFrame and a PayloadFrame
	// 1. Prepare MetaFrame
	transactionID := "yomo"
	var tag byte = 0x01
	meta := NewMetaFrame(transactionID)
	// 2. Prepare PayloadFrame
	payload := NewPayloadFrame(tag)
	payload.SetLength(len(payloadData))
	payload.SetCarriageReader(payloadReader)
	// 3. combine to DataFrame
	// data := new(DataFrame)
	// data.metaFrame = meta
	// data.payloadFrame = payload

	// TODO:
	enc := y3.NewStreamEncoder(TagOfDataFrame)
	enc.AddPacketBuffer(meta.Encode())
	// enc.AddStreamPacket(tag, len(payloadData), payloadReader)
	enc.AddStreamPacket(payload.Sid, payload.length, payload.reader)
	fmt.Printf("length=%d\n", enc.GetLen())

	r := enc.GetReader()
	buf, err := ioutil.ReadAll(r)
	fmt.Printf("err=%v\n", err)
	fmt.Printf("buf=%# x\n", buf)
}
