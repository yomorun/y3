package y3

import (
	"bytes"
)

// NodePacketEncoder used for encode a node packet
type NodePacketEncoder struct {
	*encoder
}

// NewNodePacketEncoder returns an Encoder for node packet
func NewNodePacketEncoder(sid byte) *NodePacketEncoder {
	nodeEnc := &NodePacketEncoder{
		encoder: &encoder{
			isNode: true,
			buf:    new(bytes.Buffer),
		},
	}

	nodeEnc.seqID = sid
	return nodeEnc
}

// // NewNodeSlicePacketEncoder returns an Encoder for node packet that is a slice
// func NewNodeSlicePacketEncoder(sid byte) *NodePacketEncoder {
// 	nodeEnc := &NodePacketEncoder{
// 		encoder: encoder{
// 			isNode:  true,
// 			isArray: true,
// 			buf:     new(bytes.Buffer),
// 		},
// 	}

// 	nodeEnc.seqID = sid
// 	return nodeEnc
// }

// AddNodePacket add new node to this node
func (enc *NodePacketEncoder) AddNodePacket(np *NodePacketEncoder) {
	enc.addRawPacket(np)
}

// AddPrimitivePacket add new primitive to this node
func (enc *NodePacketEncoder) AddPrimitivePacket(np *PrimitivePacketEncoder) {
	enc.addRawPacket(np)
}
