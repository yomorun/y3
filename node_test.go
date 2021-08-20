package y3

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyNode(t *testing.T) {
	np := NewNodePacketEncoder(0x06)
	np.AddBytes([]byte{})
	assert.Equal(t, []byte{0x86, 0x00}, np.Encode())
	assert.Equal(t, true, np.IsEmpty())
	res := &NodePacket{}
	endPos, err := DecodeToNodePacket(np.Encode(), res)
	assert.NoError(t, err)
	assert.Equal(t, np.Encode(), res.GetRawBytes())
	assert.Equal(t, 0, len(res.NodePackets))
	assert.Equal(t, 0, len(res.PrimitivePackets))
	assert.Equal(t, 2, endPos)
}

func TestSubEmptyNode(t *testing.T) {
	sub := NewNodePacketEncoder(0x03)
	sub.AddBytes([]byte{})
	assert.Equal(t, []byte{0x83, 0x00}, sub.Encode())

	node := NewNodePacketEncoder(0x06)
	node.AddNodePacket(sub)
	assert.Equal(t, []byte{0x86, 0x02, 0x83, 0x00}, node.Encode())

	res := &NodePacket{}
	endPos, err := DecodeToNodePacket(node.Encode(), res)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x83, 0x00}, res.GetValBuf())
	assert.Equal(t, 4, endPos)
	assert.Equal(t, node.Encode(), res.GetRawBytes())
	assert.Equal(t, 1, len(res.NodePackets))
	assert.Equal(t, 0, len(res.PrimitivePackets))
	val, ok := res.NodePackets[0x03]
	assert.EqualValues(t, true, ok)
	assert.Equal(t, []byte{0x83, 0x00}, val.GetRawBytes())
	if ok {
		assert.NoError(t, err)
		assert.Equal(t, 0, val.Length())
	}
}

// Assume a JSON object like this：
// '0x04': {
//   '0x01': -1,
// },
// YoMo Codec should ->
// 0x84 (is a node, sequence id=4)
//   0x03 (node value length is 4 bytes)
//     0x01, 0x01, 0x7F (pvarint: -1)
func TestSimple1Node(t *testing.T) {
	sub := NewPrimitivePacketEncoder(0x01)
	sub.SetInt32Value(-1)
	assert.Equal(t, []byte{0x01, 0x01, 0xFF}, sub.Encode())

	node := NewNodePacketEncoder(0x04)
	node.AddPrimitivePacket(sub)
	assert.Equal(t, []byte{0x84, 0x03, 0x01, 0x01, 0xFF}, node.Encode())

	res := &NodePacket{}
	consumedBytes, err := DecodeToNodePacket(node.Encode(), res)
	assert.NoError(t, err)
	assert.Equal(t, node.Encode(), res.GetRawBytes())
	assert.Equal(t, 0, len(res.NodePackets))
	assert.Equal(t, 1, len(res.PrimitivePackets))
	assert.EqualValues(t, 0x04, res.SeqID())

	val, ok := res.PrimitivePackets[1]
	assert.EqualValues(t, true, ok)
	if ok {
		assert.Equal(t, []byte{0x01, 0x01, 0xFF}, val.GetRawBytes())
		v, err := val.ToInt32()
		assert.NoError(t, err)
		assert.Equal(t, []byte{0x01, 0x01, 0xFF}, val.GetRawBytes())
		assert.Equal(t, int32(-1), v)
	}
	assert.Equal(t, 5, consumedBytes)
}

// Assume a JSON object like this：
// '0x04': {
//   '0x01': -1,
// },
// YoMo Codec should ->
// 0x84 (is a node, sequence id=4)
//   0x03 (node value length is 4 bytes)
//     0x01, 0x01, 0x7F (pvarint: -1)
func TestSimpleNodes(t *testing.T) {
	buf := []byte{0x85, 0x05, 0x84, 0x03, 0x01, 0x01, 0xFF}
	res := &NodePacket{}
	consumedBytes, err := DecodeToNodePacket(buf, res)
	assert.NoError(t, err)
	assert.Equal(t, buf, res.GetRawBytes())
	assert.Equal(t, 1, len(res.NodePackets))
	assert.Equal(t, 0, len(res.PrimitivePackets))
	assert.EqualValues(t, 0x05, res.SeqID())

	n, ok := res.NodePackets[0x04]
	assert.Equal(t, true, ok)
	assert.Equal(t, []byte{0x84, 0x03, 0x01, 0x01, 0xFF}, n.GetRawBytes())
	assert.Equal(t, 0, len(n.NodePackets))
	assert.Equal(t, 1, len(n.PrimitivePackets))
	assert.EqualValues(t, 0x04, n.SeqID())

	val, ok := n.PrimitivePackets[0x01]
	assert.EqualValues(t, true, ok)
	assert.Equal(t, []byte{0x01, 0x01, 0xFF}, val.GetRawBytes())
	if ok {
		v, err := val.ToInt32()
		assert.NoError(t, err)
		assert.Equal(t, int32(-1), v)
	}
	assert.Equal(t, 7, consumedBytes)
}

// Assume a JSON object like this：
// '0x03': {
//   '0x01': -1,
//   '0x02':  1,
// },
// YoMo Codec should ->
// 0x83 (is a node, sequence id=3)
//   0x06 (node value length is 8 bytes)
//     0x01, 0x01, 0x7F (pvarint: -1)
//     0x02, 0x01, 0x01 (pvarint: 1)
func TestSimple2Nodes(t *testing.T) {
	buf := []byte{0x83, 0x06, 0x01, 0x01, 0xFF, 0x02, 0x01, 0x01}
	res := &NodePacket{}
	consumedBytes, err := DecodeToNodePacket(buf, res)
	assert.NoError(t, err)
	assert.Equal(t, buf, res.GetRawBytes())
	assert.Equal(t, len(buf), consumedBytes)
	assert.Equal(t, 0, len(res.NodePackets))
	assert.Equal(t, 2, len(res.PrimitivePackets))

	v1, ok := res.PrimitivePackets[0x01]
	assert.EqualValues(t, true, ok)
	v, err := v1.ToInt32()
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x01, 0x01, 0xFF}, v1.GetRawBytes())
	assert.EqualValues(t, -1, v)
	assert.NoError(t, err)

	v2, ok := res.PrimitivePackets[0x02]
	assert.EqualValues(t, true, ok)
	assert.Equal(t, []byte{0x02, 0x01, 0x01}, v2.GetRawBytes())
	v, err = v2.ToInt32()
	assert.NoError(t, err)
	assert.EqualValues(t, 1, v)
}

// Assume a JSON object like this：
// '0x05': {
//	'0x04': {
//     '0x01': -1,
//     '0x02':  1,
//  },
//	'0x03': {
//     '0x01': -2,
//  },
// }
// YoMo Codec should ->
// 0x85
//   0x0D(node value length is 15 bytes)
//     0x84 (is a node, sequence id=3)
//       0x06 (node value length is 8 bytes)
//         0x01, 0x01, 0x7F (varint: -1)
//         0x02, 0x01, 0x43 (string: "C")
//     0x83 (is a node, sequence id=4)
//       0x03 (node value length is 4 bytes)
//         0x01, 0x01, 0x7E (varint: -2)
func TestComplexNodes(t *testing.T) {
	buf := []byte{0x85, 0x0D, 0x84, 0x06, 0x01, 0x01, 0xFF, 0x02, 0x01, 0x43, 0x83, 0x03, 0x01, 0x01, 0xFE}
	res := &NodePacket{}
	consumedBytes, err := DecodeToNodePacket(buf, res)
	assert.NoError(t, err)
	assert.Equal(t, buf, res.GetRawBytes())
	assert.Equal(t, len(buf), consumedBytes)
	assert.Equal(t, 2, len(res.NodePackets))
	assert.Equal(t, 0, len(res.PrimitivePackets))

	n1, ok := res.NodePackets[0x04]
	assert.EqualValues(t, true, ok)
	assert.Equal(t, []byte{0x84, 0x06, 0x01, 0x01, 0xFF, 0x02, 0x01, 0x43}, n1.GetRawBytes())
	assert.Equal(t, 2, len(n1.PrimitivePackets))

	n1p1, ok := n1.PrimitivePackets[0x01]
	assert.EqualValues(t, true, ok)
	assert.Equal(t, []byte{0x01, 0x01, 0xFF}, n1p1.GetRawBytes())
	vn1p1, err := n1p1.ToInt32()
	assert.NoError(t, err)
	assert.EqualValues(t, -1, vn1p1)

	n1p2, ok := n1.PrimitivePackets[0x02]
	assert.EqualValues(t, true, ok)
	assert.Equal(t, []byte{0x02, 0x01, 0x43}, n1p2.GetRawBytes())
	vn1p2, err := n1p2.ToUTF8String()
	assert.NoError(t, err)
	assert.Equal(t, "C", vn1p2)

	n2, ok := res.NodePackets[0x03]
	assert.EqualValues(t, true, ok)
	assert.Equal(t, []byte{0x83, 0x03, 0x01, 0x01, 0xFE}, n2.GetRawBytes())
	assert.Equal(t, 1, len(n2.PrimitivePackets))

	n2p1, ok := n2.PrimitivePackets[0x01]
	assert.EqualValues(t, true, ok)
	assert.Equal(t, []byte{0x01, 0x01, 0xFE}, n2p1.GetRawBytes())
	vn2p1, err := n2p1.ToInt32()
	assert.NoError(t, err)
	assert.EqualValues(t, -2, vn2p1)
}
