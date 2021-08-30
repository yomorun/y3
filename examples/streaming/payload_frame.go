package main

import (
	"io"

	"github.com/yomorun/y3"
)

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
func (m *PayloadFrame) SetCarriage(buf []byte) *PayloadFrame {
	m.Carriage = buf
	return m
}

// Encode to Y3 encoded bytes
func (m *PayloadFrame) Encode() []byte {
	carriage := y3.NewPrimitivePacketEncoder(m.Sid)
	carriage.SetBytesValue(m.Carriage)

	payload := y3.NewNodePacketEncoder(byte(TagOfPayloadFrame))
	payload.AddPrimitivePacket(carriage)

	return payload.Encode()
}

func (m *PayloadFrame) SetLength(length int) {
	m.length = length
}

func (m *PayloadFrame) SetCarriageReader(reader io.Reader) {
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
