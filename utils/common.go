package utils

// MSB is `1000 0000` describes this is a node packet, otherwise, is a primitive packet
const MSB byte = 0x80

// DropMSB is `0111 1111`, used to remove MSB flag bit
const DropMSB byte = 0x3F

// SliceFlag is `0100 0000`, describes this packet is a Slice type
const SliceFlag byte = 0x40

// DropMSBArrayFlag is `0011 1111`, used to remove MSB and Slice flag bit
const DropMSBArrayFlag byte = 0x3F
