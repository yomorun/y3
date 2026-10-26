> 📚 VERSION: draft-01
> ⛳️ STATE: v1.0.0

# Y3

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fyomorun%2Fy3.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fyomorun%2Fy3?ref=badge_shield)

Y3 is the golang implementation of [Y3 Codec](https://github.com/yomorun/y3-codec), which describe a fast and low CPU binding data encoder/decoder focus on edge computing and streaming processing.

# Advantage

- Super fast encode/decode for streaming data
- Low CPU consumption when decoding large data
- Random access
- Tuned for QUIC protocol
- Designed for global communication at high frequency

## Y3 Codec Specification

See [Y3 Codec SPEC](https://github.com/yomorun/y3-codec)

## Test

`make test`

## Use

`go get -u github.com/yomorun/y3`

## Examples

### Encode examples

```go
package main

import (
	"fmt"

	y3 "github.com/yomorun/y3"
)

func main() {
	// if we want to represent `var obj = &foo{ID: -1, bar: &bar{Name: "C"}}`
	// in Y3-Codec:

	// 0x81 -> node
	var foo = y3.NewNodePacketEncoder(0x01)

	// 0x02 -> foo.ID=-1
	var yp1 = y3.NewPrimitivePacketEncoder(0x02)
	yp1.SetInt32Value(-1)
	foo.AddPrimitivePacket(yp1)

	// 0x83 -> &bar{}
	var bar = y3.NewNodePacketEncoder(0x03)

	// 0x04 -> bar.Name="C"
	var yp2 = y3.NewPrimitivePacketEncoder(0x04)
	yp2.SetStringValue("C")
	bar.AddPrimitivePacket(yp2)

	// -> foo.bar=&bar
	foo.AddNodePacket(bar)

	fmt.Printf("res=%#v", foo.Encode()) // res=[]byte{0x81, 0x8, 0x2, 0x1, 0xff, 0x83, 0x3, 0x4, 0x1, 0x43}
}
```

### Decode examples 1: decode a primitive packet

```go
package main

import (
	"fmt"

	y3 "github.com/yomorun/y3"
)

func main() {
	fmt.Println(">> Parsing [0x0A, 0x01, 0x7F], which like Key-Value format = 0x0A: 127")
	buf := []byte{0x0A, 0x01, 0x7F}
	var res y3.PrimitivePacket
	if _, err := y3.DecodeToPrimitivePacket(buf, &res); err != nil {
		panic(err)
	}
	v1, err := res.ToUInt32()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Tag Key=[%#X], Value=%v\n", res.SeqID(), v1) // Tag Key=[0XA], Value=127
}
```

### Decode examples 2: decode a node packet

Children of a node are kept in maps keyed by their sequence ID.

```go
package main

import (
	"fmt"

	y3 "github.com/yomorun/y3"
)

func main() {
	fmt.Println(">> Parsing [0x84, 0x06, 0x0A, 0x01, 0xFF, 0x0B, 0x01, 0x43] EQUALS JSON= 0x84: { 0x0A: -1, 0x0B: 'C' }")
	buf := []byte{0x84, 0x06, 0x0A, 0x01, 0xFF, 0x0B, 0x01, 0x43}
	var res y3.NodePacket
	if _, err := y3.DecodeToNodePacket(buf, &res); err != nil {
		panic(err)
	}

	v1 := res.PrimitivePackets[0x0A]
	p1, err := v1.ToInt32()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Tag Key=[%#X.%#X], Value=%v\n", res.SeqID(), v1.SeqID(), p1) // Tag Key=[0X4.0XA], Value=-1

	v2 := res.PrimitivePackets[0x0B]
	p2, err := v2.ToUTF8String()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Tag Key=[%#X.%#X], Value=%v\n", res.SeqID(), v2.SeqID(), p2) // Tag Key=[0X4.0XB], Value=C
}
```

## Streaming

Y3 packets carry their length up front, so they can be read from and written to an `io.Reader`, such as a QUIC stream, one at a time. See [`examples/streaming`](examples/streaming) for a complete program.

### Read packets from a stream

`ReadPacket` reads one whole packet (Tag, Length and Value) from the stream:

```go
for {
	buf, err := y3.ReadPacket(stream)
	if err == io.EOF {
		break // the stream was closed between packets
	}
	if err != nil {
		return err // y3.ErrMalformed: the stream ended inside a packet
	}
	// buf holds one complete packet, ready for DecodeToNodePacket or DecodeToPrimitivePacket
}
```

### Read the value lazily

`StreamReadPacket` reads only Tag and Length. It returns the value as an `io.Reader`, so a large value can be copied onward without holding it in memory:

```go
sr, err := y3.StreamReadPacket(stream)
if err != nil {
	return err
}
// sr.Val yields exactly sr.Len bytes and then io.EOF,
// or io.ErrUnexpectedEOF if the stream ends first
_, err = io.Copy(dst, sr.Val)
```

`sr.Val` reads from the same stream, so read it to the end before reading the next packet. To walk the children of a node, call `StreamReadPacket(sr.Val)` until it returns `io.EOF`.

### Encode with a streamed value

`StreamEncoder` builds a node whose last child is read lazily from an `io.Reader`:

```go
meta := y3.NewPrimitivePacketEncoder(0x01)
meta.SetStringValue("yomo")

enc := y3.NewStreamEncoder(0x81)         // the raw tag byte; 0x80 marks a node
enc.AddPacket(meta)                      // packets known up front
enc.AddStreamPacket(0x02, size, src)     // the value: size bytes, read lazily from the io.Reader src
_, err := io.Copy(conn, enc.GetReader()) // writes enc.GetLen() bytes
```

Call `AddStreamPacket` exactly once, after all other packets. `GetReader` returns an empty reader until it has been called. The encoder reads exactly `size` bytes from `src`. If `src` ends sooner, `GetReader` returns `io.ErrUnexpectedEOF`.

## Types

Y3 implements the [Y3 Codec](https://github.com/yomorun/y3-codec) protocol and supports the following Golang data types:

| Go type | Encode | Decode | Value | Encoded with sequence ID `0x01` |
|---|---|---|---|---|
| `int32` | `SetInt32Value` | `ToInt32` | `-1` | `01 01 FF` |
| `uint32` | `SetUInt32Value` | `ToUInt32` | `255` | `01 02 00 FF` |
| `int64` | `SetInt64Value` | `ToInt64` | `-300` | `01 02 FE D4` |
| `uint64` | `SetUInt64Value` | `ToUInt64` | `1 << 40` | `01 06 01 00 00 00 00 00` |
| `float32` | `SetFloat32Value` | `ToFloat32` | `1.5` | `01 02 3F C0` |
| `float64` | `SetFloat64Value` | `ToFloat64` | `2` | `01 01 40` |
| `bool` | `SetBoolValue` | `ToBool` | `true` | `01 01 01` |
| `string` | `SetStringValue` | `ToUTF8String` | `"yomo"` | `01 04 79 6F 6D 6F` |
| `[]byte` | `SetBytesValue` | `ToBytes` | `[]byte("yomo")` | `01 04 79 6F 6D 6F` |

Integers use a variable-length encoding, so small values take fewer bytes. Floats drop trailing zero bytes.

```go
p := y3.NewPrimitivePacketEncoder(0x01)
p.SetInt32Value(-1)
buf := p.Encode() // []byte{0x01, 0x01, 0xFF}

var res y3.PrimitivePacket
if _, err := y3.DecodeToPrimitivePacket(buf, &res); err != nil {
	panic(err)
}
v, err := res.ToInt32() // v == -1
```

## Limitations

- Sequence IDs range from `0x00` to `0x3F`. The encoders panic on larger values.
- `DecodeToNodePacket` keeps one child per sequence ID, so a node whose children share a sequence ID (a list) decodes to its last child only. See [#20](https://github.com/yomorun/y3/issues/20).

## Contributors

[//]: contributor-faces

<a href="https://github.com/figroc"><img src="https://avatars1.githubusercontent.com/u/2026460?v=3" title="figroc" width="60" height="60"></a>

[//]: contributor-faces

## License

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fyomorun%2Fy3.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fyomorun%2Fy3?ref=badge_large)
