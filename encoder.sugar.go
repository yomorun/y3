package y3

import (
	"github.com/yomorun/y3/encoding"
)

// SetUTF8StringV set utf-8 string type value as V
func (b *Encoder) SetUTF8StringV(v string) {
	buf := []byte(v)
	b.SetBytesV(buf)
}

// SetInt32V set an int32 type value as V
func (b *Encoder) SetInt32V(v int32) error {
	size := encoding.SizeOfNVarInt32(v)
	codec := encoding.VarCodec{Size: size}
	buf := make([]byte, size)
	err := codec.EncodeNVarInt32(buf, v)
	if err != nil {
		return err
	}
	b.SetBytesV(buf)
	return nil
}

// SetUInt32V set an uint32 type value as V
func (b *Encoder) SetUInt32V(v uint32) error {
	size := encoding.SizeOfNVarUInt32(v)
	codec := encoding.VarCodec{Size: size}
	buf := make([]byte, size)
	err := codec.EncodeNVarUInt32(buf, v)
	if err != nil {
		return err
	}
	b.SetBytesV(buf)
	return nil
}

// SetInt64V set an int64 type value as V
func (b *Encoder) SetInt64V(v int64) error {
	size := encoding.SizeOfNVarInt64(v)
	codec := encoding.VarCodec{Size: size}
	buf := make([]byte, size)
	err := codec.EncodeNVarInt64(buf, v)
	if err != nil {
		return err
	}
	b.SetBytesV(buf)
	return nil
}

// SetUInt64V set an uint64 type value as V
func (b *Encoder) SetUInt64V(v uint64) error {
	size := encoding.SizeOfNVarUInt64(v)
	codec := encoding.VarCodec{Size: size}
	buf := make([]byte, size)
	err := codec.EncodeNVarUInt64(buf, v)
	if err != nil {
		return err
	}
	b.SetBytesV(buf)
	return nil
}

// SetFloat32V set an float32 type value as V
func (b *Encoder) SetFloat32V(v float32) error {
	size := encoding.SizeOfVarFloat32(v)
	codec := encoding.VarCodec{Size: size}
	buf := make([]byte, size)
	err := codec.EncodeVarFloat32(buf, v)
	if err != nil {
		return err
	}
	b.SetBytesV(buf)
	return nil
}

// SetFloat64V set an float64 type value as V
func (b *Encoder) SetFloat64V(v float64) error {
	size := encoding.SizeOfVarFloat64(v)
	codec := encoding.VarCodec{Size: size}
	buf := make([]byte, size)
	err := codec.EncodeVarFloat64(buf, v)
	if err != nil {
		return err
	}
	b.SetBytesV(buf)
	return nil
}

// SetBoolV set bool type value as V
func (b *Encoder) SetBoolV(v bool) {
	var size = encoding.SizeOfPVarUInt32(uint32(1))
	codec := encoding.VarCodec{Size: size}
	buf := make([]byte, size)
	codec.EncodePVarBool(buf, v)
	b.SetBytesV(buf)
}
