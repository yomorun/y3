package y3

// NodePacket describes complex values
type NodePacket struct {
	*basePacket
	// NodePackets store the node packets by sequence ID; when several
	// children share a sequence ID only the last one is kept, use Range
	// to visit all of them
	NodePackets map[byte]NodePacket
	// PrimitivePackets store the primitive packets by sequence ID; when
	// several children share a sequence ID only the last one is kept, use
	// Range to visit all of them
	PrimitivePackets map[byte]PrimitivePacket
}

// Range calls fn for each child of the node in wire order, including
// children that share a sequence ID. For each child exactly one of node
// and prim is non-nil. Range stops when fn returns false.
func (n NodePacket) Range(fn func(seqID byte, node *NodePacket, prim *PrimitivePacket) bool) error {
	if n.basePacket == nil {
		return nil
	}

	buf := n.valbuf
	for pos := 0; pos < len(buf); {
		consumed, isNode, np, pp, err := parsePayload(buf[pos:])
		if err != nil {
			return err
		}
		pos += consumed

		if isNode {
			if !fn(np.SeqID(), np, nil) {
				return nil
			}
		} else if !fn(pp.SeqID(), nil, pp) {
			return nil
		}
	}
	return nil
}
