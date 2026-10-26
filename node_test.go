package y3

import (
	"fmt"
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

func TestTruncatedNode(t *testing.T) {
	sub := NewPrimitivePacketEncoder(0x02)
	sub.SetStringValue("hello")
	node := NewNodePacketEncoder(0x01)
	node.AddPrimitivePacket(sub)
	buf := node.Encode()
	assert.Equal(t, []byte{0x81, 0x07, 0x02, 0x05, 0x68, 0x65, 0x6C, 0x6C, 0x6F}, buf)

	for i := 1; i < len(buf); i++ {
		// with and without spare capacity behind the truncated buffer
		for _, in := range [][]byte{buf[:i:i], buf[:i]} {
			assert.NotPanics(t, func() {
				_, err := DecodeToNodePacket(in, &NodePacket{})
				assert.Error(t, err, "input % x", in)
			})
		}
	}

	// the outer node is complete, the nested node is truncated
	assert.NotPanics(t, func() {
		_, err := DecodeToNodePacket([]byte{0x81, 0x04, 0x82, 0x05, 0x01, 0x01}, &NodePacket{})
		assert.Error(t, err)
	})
}

func TestNodeRangeRepeatedSeqID(t *testing.T) {
	node := NewNodePacketEncoder(0x01)
	for _, s := range []string{"a", "b", "c"} {
		p := NewPrimitivePacketEncoder(0x02)
		p.SetStringValue(s)
		node.AddPrimitivePacket(p)
	}

	res := &NodePacket{}
	_, err := DecodeToNodePacket(node.Encode(), res)
	assert.NoError(t, err)
	// the map keeps only the last child for a sequence ID
	assert.Equal(t, 1, len(res.PrimitivePackets))

	var got []string
	err = res.Range(func(seqID byte, node *NodePacket, prim *PrimitivePacket) bool {
		assert.EqualValues(t, 0x02, seqID)
		assert.Nil(t, node)
		v, err := prim.ToUTF8String()
		assert.NoError(t, err)
		got = append(got, v)
		return true
	})
	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, got)
}

func TestNodeRangeOrder(t *testing.T) {
	// a list of messages, each a node with the same sequence ID, between two primitives
	list := NewNodePacketEncoder(0x01)
	first := NewPrimitivePacketEncoder(0x02)
	first.SetInt32Value(1)
	list.AddPrimitivePacket(first)
	for _, s := range []string{"hi", "hello"} {
		msg := NewNodePacketEncoder(0x03)
		content := NewPrimitivePacketEncoder(0x01)
		content.SetStringValue(s)
		msg.AddPrimitivePacket(content)
		list.AddNodePacket(msg)
	}
	last := NewPrimitivePacketEncoder(0x02)
	last.SetInt32Value(2)
	list.AddPrimitivePacket(last)

	res := &NodePacket{}
	_, err := DecodeToNodePacket(list.Encode(), res)
	assert.NoError(t, err)

	var got []string
	err = res.Range(func(seqID byte, node *NodePacket, prim *PrimitivePacket) bool {
		if node != nil {
			assert.Nil(t, prim)
			content := node.PrimitivePackets[0x01]
			v, err := content.ToUTF8String()
			assert.NoError(t, err)
			got = append(got, fmt.Sprintf("node %#x: %s", seqID, v))
		} else {
			v, err := prim.ToInt32()
			assert.NoError(t, err)
			got = append(got, fmt.Sprintf("prim %#x: %d", seqID, v))
		}
		return true
	})
	assert.NoError(t, err)
	assert.Equal(t, []string{"prim 0x2: 1", "node 0x3: hi", "node 0x3: hello", "prim 0x2: 2"}, got)
}

func TestNodeRangeStop(t *testing.T) {
	node := NewNodePacketEncoder(0x01)
	for i := 0; i < 3; i++ {
		p := NewPrimitivePacketEncoder(0x02)
		p.SetInt32Value(int32(i))
		node.AddPrimitivePacket(p)
	}

	res := &NodePacket{}
	_, err := DecodeToNodePacket(node.Encode(), res)
	assert.NoError(t, err)

	calls := 0
	err = res.Range(func(byte, *NodePacket, *PrimitivePacket) bool {
		calls++
		return false
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, calls)
}

func TestNodeRangeEmpty(t *testing.T) {
	sub := NewNodePacketEncoder(0x03)
	sub.AddBytes([]byte{})
	node := NewNodePacketEncoder(0x06)
	node.AddNodePacket(sub)

	res := &NodePacket{}
	_, err := DecodeToNodePacket(node.Encode(), res)
	assert.NoError(t, err)

	var seqIDs []byte
	err = res.Range(func(seqID byte, node *NodePacket, _ *PrimitivePacket) bool {
		seqIDs = append(seqIDs, seqID)
		// the child itself is empty
		assert.NoError(t, node.Range(func(byte, *NodePacket, *PrimitivePacket) bool {
			t.Error("empty node has no children")
			return true
		}))
		return true
	})
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x03}, seqIDs)

	// a node that was never decoded has no children either
	assert.NoError(t, NodePacket{}.Range(func(byte, *NodePacket, *PrimitivePacket) bool {
		t.Error("zero NodePacket has no children")
		return true
	}))
}
