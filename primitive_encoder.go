package y3

import (
	"bytes"

	"github.com/yomorun/y3/encoding"
)

// PrimitivePacketEncoder used for encode a primitive packet
type PrimitivePacketEncoder struct {
	encoder
}

// NewPrimitivePacketEncoder return an Encoder for primitive packet
func NewPrimitivePacketEncoder(sid byte) *PrimitivePacketEncoder {
	primEnc := &PrimitivePacketEncoder{
		encoder: encoder{
			isNode: false,
			buf:    new(bytes.Buffer),
		},
	}

	primEnc.seqID = sid
	return primEnc
}

// SetInt32Value encode int32 value
func (enc *PrimitivePacketEncoder) SetInt32Value(v int32) {
	size := encoding.SizeOfPVarInt32(v)
	codec := encoding.VarCodec{Size: size}
	enc.valbuf = make([]byte, size)
	err := codec.EncodePVarInt32(enc.valbuf, v)
	if err != nil {
		panic(err)
	}
	// enc.valBuf.Write(buf)
}

// SetUInt32Value encode uint32 value
func (enc *PrimitivePacketEncoder) SetUInt32Value(v uint32) {
	size := encoding.SizeOfPVarUInt32(v)
	codec := encoding.VarCodec{Size: size}
	enc.valbuf = make([]byte, size)
	err := codec.EncodePVarUInt32(enc.valbuf, v)
	if err != nil {
		panic(err)
	}
}

// SetInt64Value encode int64 value
func (enc *PrimitivePacketEncoder) SetInt64Value(v int64) {
	size := encoding.SizeOfPVarInt64(v)
	codec := encoding.VarCodec{Size: size}
	enc.valbuf = make([]byte, size)
	err := codec.EncodePVarInt64(enc.valbuf, v)
	if err != nil {
		panic(err)
	}
}

// SetUInt64Value encode uint64 value
func (enc *PrimitivePacketEncoder) SetUInt64Value(v uint64) {
	size := encoding.SizeOfPVarUInt64(v)
	codec := encoding.VarCodec{Size: size}
	enc.valbuf = make([]byte, size)
	err := codec.EncodePVarUInt64(enc.valbuf, v)
	if err != nil {
		panic(err)
	}
}

// SetFloat32Value encode float32 value
func (enc *PrimitivePacketEncoder) SetFloat32Value(v float32) {
	var size = encoding.SizeOfVarFloat32(v)
	codec := encoding.VarCodec{Size: size}
	enc.valbuf = make([]byte, size)
	err := codec.EncodeVarFloat32(enc.valbuf, v)
	if err != nil {
		panic(err)
	}
}

// SetFloat64Value encode float64 value
func (enc *PrimitivePacketEncoder) SetFloat64Value(v float64) {
	var size = encoding.SizeOfVarFloat64(v)
	codec := encoding.VarCodec{Size: size}
	enc.valbuf = make([]byte, size)
	err := codec.EncodeVarFloat64(enc.valbuf, v)
	if err != nil {
		panic(err)
	}
}

// SetBoolValue encode bool value
func (enc *PrimitivePacketEncoder) SetBoolValue(v bool) {
	var size = encoding.SizeOfPVarUInt32(uint32(1))
	codec := encoding.VarCodec{Size: size}
	enc.valbuf = make([]byte, size)
	err := codec.EncodePVarBool(enc.valbuf, v)
	if err != nil {
		panic(err)
	}
}

// SetStringValue encode string
func (enc *PrimitivePacketEncoder) SetStringValue(v string) {
	// buf := []byte(v)
	// enc.valBuf.Write(buf)
	enc.valbuf = []byte(v)
}

// SetBytesValue encode []byte
func (enc *PrimitivePacketEncoder) SetBytesValue(v []byte) {
	enc.valbuf = v
}
