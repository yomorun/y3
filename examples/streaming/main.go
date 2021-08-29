package main

import (
	"github.com/yomorun/y3"
)

var (
	TagOfDataFrame     byte = 0x3F
	TagOfMetaFrame     byte = 0x2F
	TagOfPayloadFrame  byte = 0x2E
	TagOfTransactionID byte = 0x01
)

func main() {
	// TODO:
	enc := y3.NewStreamEncoder(TagOfDataFrame)
	enc.GetLen()
}
