package y3

// NodePacket describes complex values
type NodePacket struct {
	*basePacket
	// NodePackets store all the node packets
	NodePackets map[byte]NodePacket
	// PrimitivePackets store all the primitive packets
	PrimitivePackets map[byte]PrimitivePacket
}
