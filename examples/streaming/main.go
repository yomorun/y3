package main

import (
	"fmt"
	"io"

	// "io/ioutil"
	"sync"
	"time"

	"github.com/yomorun/y3"
)

var (
	TagOfDataFrame     byte = 0x3F
	TagOfMetaFrame     byte = 0x2F
	TagOfPayloadFrame  byte = 0x2E
	TagOfTransactionID byte = 0x01
)

type p struct {
	buf      []byte
	lastRead int
	wg       *sync.WaitGroup
}

func (o *p) Read(buf []byte) (int, error) {
	o.wg.Add(1)
	defer o.wg.Done()
	if o.lastRead >= len(o.buf) {
		return 0, io.EOF
	}
	time.Sleep(1 * time.Second)
	fmt.Printf("(source stream)==>flush:[%# x]\n", o.buf[o.lastRead])
	copy(buf, []byte{o.buf[o.lastRead]})
	o.lastRead++
	return 1, nil
}

func main() {
	var wg sync.WaitGroup
	payloadData := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	payloadReader := &p{buf: payloadData, wg: &wg}
	// payloadReader.Write(payloadData)
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
	enc := y3.NewStreamEncoder(TagOfDataFrame)
	enc.AddPacketBuffer(meta.Encode())
	// enc.AddStreamPacket(tag, len(payloadData), payloadReader)
	enc.AddStreamPacket(payload.Sid, payload.length, payload.reader)

	// try read
	fmt.Printf("length=%d\n", enc.GetLen())
	r := enc.GetReader()

	// // method 1: try read all
	// buf, err := ioutil.ReadAll(r)
	// fmt.Printf("err=%v\n", err)
	// fmt.Printf("buf=%# x\n", buf)

	// method 2: try read from reader
	for {
		sp, err := y3.StreamReadPacket(r)
		if err != nil {
			fmt.Printf("err=%v\n", err)
			break
		}
		fmt.Printf(">> tag=%# x\n", sp.Tag)
		fmt.Printf("length=%d\n", sp.Len)
		// if sp.Tag == tag {
		tmp := make([]byte, 1)
		for {
			n, err := sp.Val.Read(tmp)
			if err != nil {
				if err == io.EOF {
					fmt.Printf("\t-> %# x\n", tmp[:n])
				}
				break
			}
			fmt.Printf("\t-> %# x\n", tmp[:n])
		}
		// }
	}

	wg.Wait()
	fmt.Println("OVER")
}
