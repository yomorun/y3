package y3

import (
	"github.com/yomorun/y3/encoding"
)

// the minimal length of a packet is 2 bytes
const primitivePacketBufferMinimalLength = 2

// PrimitivePacket describes primitive value type,
type PrimitivePacket struct {
	*basePacket
}

// ToInt32 parse raw as int32 value
func (p *PrimitivePacket) ToInt32() (int32, error) {
	var val int32
	codec := encoding.VarCodec{Size: len(p.valbuf)}
	err := codec.DecodeNVarInt32(p.basePacket.valbuf, &val)
	if err != nil {
		return 0, err
	}
	return val, nil
}

// ToUInt32 parse raw as uint32 value
func (p *PrimitivePacket) ToUInt32() (uint32, error) {
	var val uint32
	codec := encoding.VarCodec{Size: len(p.valbuf)}
	err := codec.DecodeNVarUInt32(p.valbuf, &val)
	if err != nil {
		return 0, err
	}
	return val, nil
}

// ToInt64 parse raw as int64 value
func (p *PrimitivePacket) ToInt64() (int64, error) {
	var val int64
	codec := encoding.VarCodec{Size: len(p.valbuf)}
	err := codec.DecodeNVarInt64(p.valbuf, &val)
	if err != nil {
		return 0, err
	}
	return val, nil
}

// ToUInt64 parse raw as uint64 value
func (p *PrimitivePacket) ToUInt64() (uint64, error) {
	var val uint64
	codec := encoding.VarCodec{Size: len(p.valbuf)}
	err := codec.DecodeNVarUInt64(p.valbuf, &val)
	if err != nil {
		return 0, err
	}
	return val, nil
}

// ToFloat32 parse raw as float32 value
func (p *PrimitivePacket) ToFloat32() (float32, error) {
	var val float32
	codec := encoding.VarCodec{Size: len(p.valbuf)}
	err := codec.DecodeVarFloat32(p.valbuf, &val)
	if err != nil {
		return 0, err
	}
	return val, nil
}

// ToFloat64 parse raw as float64 value
func (p *PrimitivePacket) ToFloat64() (float64, error) {
	var val float64
	codec := encoding.VarCodec{Size: len(p.valbuf)}
	err := codec.DecodeVarFloat64(p.valbuf, &val)
	if err != nil {
		return 0, err
	}
	return val, nil
}

// ToBool parse raw as bool value
func (p *PrimitivePacket) ToBool() (bool, error) {
	var val bool
	codec := encoding.VarCodec{Size: len(p.valbuf)}
	err := codec.DecodePVarBool(p.valbuf, &val)
	if err != nil {
		return false, err
	}
	return val, nil
}

// ToUTF8String parse raw data as string value
func (p *PrimitivePacket) ToUTF8String() (string, error) {
	return string(p.valbuf), nil
}

// ToBytes returns raw buffer data
func (p *PrimitivePacket) ToBytes() []byte {
	return p.valbuf
}
