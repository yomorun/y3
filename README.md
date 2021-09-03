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
	// if we want to repesent `var obj = &foo{ID: -1, bar: &bar{Name: "C"}}`
	// in Y3-Codec:

	// 0x81 -> node
	var foo y3.Encoder
	foo.SetSeqID(0x01, true)

	// 0x02 -> foo.ID=-11
	var id y3.Encoder
	id.SetSeqID(0x02, false)
	id.SetInt32V(-1)

	foo.AddPacket(id)

	// 0x83 -> &bar{}
	var bar = y3.NewNodePacketEncoder(0x03)

	// 0x04 -> bar.Name="C"
	var name y3.Encoder
	name.SetSeqID(0x04)
	name.SetStringV("C")
	bar.AddPacket(name)

	// -> foo.bar=&bar
	foo.AddNodePacket(bar)

	// Buid to Packet
	packet, _ = foo.Packet()

	// Read to buf
	buf := &bytes.Buffer{}
	io.Copy(buf, packet.Reader())
	fmt.Printf("res=%#v", buf) // res=[]byte{0x81, 0x08, 0x02, 0x01, 0x7F, 0x83, 0x03, 0x04, 0x01, 0x43}
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
	res, _, err := y3.DecodePrimitivePacket(buf)
	v1, err := res.ToUInt32()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Tag Key=[%#X], Value=%v\n", res.SeqID(), v1)
}
```

### Decode examples 2: decode a node packet

```go
package main

import (
	"fmt"
	y3 "github.com/yomorun/y3"
)

func main() {
	fmt.Println(">> Parsing [0x84, 0x06, 0x0A, 0x01, 0x7F, 0x0B, 0x01, 0x43] EQUALS JSON= 0x84: { 0x0A: -1, 0x0B: 'C' }")
	buf := []byte{0x84, 0x06, 0x0A, 0x01, 0x7F, 0x0B, 0x01, 0x43}
	res, _, err := y3.DecodeNodePacket(buf)
	v1 := res.PrimitivePackets[0]

	p1, err := v1.ToInt32()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Tag Key=[%#X.%#X], Value=%v\n", res.SeqID(), v1.SeqID(), p1)

	v2 := res.PrimitivePackets[1]

	p2, err := v2.ToUTF8String()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Tag Key=[%#X.%#X], Value=%v\n", res.SeqID(), v2.SeqID(), p2)
}
```

More examples in `/examples/`

## Types

Y3 implements the [YoMo Codec](https://github.com/yomorun/yomo-codec) protocol and supports the following Golang data types:

<details>
  <summary>int32</summary>
		
```golang

````

</details>

<details>
  <summary>uint32</summary>

```golang
````

</details>

<details>
  <summary>int64</summary>
	
```golang 
```

</details>

<details>
  <summary>uint64</summary>
	
```golang 
```

</details>

<details>
  <summary>float32</summary>
	
```golang
```

</details>

<details>
  <summary>float64</summary>
	
```golang
```

</details>

<details>
  <summary>bool</summary>
	
```golang 
```

</details>

<details>
  <summary>string</summary>
  
```golang
```

</details>

<details>
  <summary>bytes</summary>
	
```golang
buf := []byte("yomo")
p := NewPrimitivePacketEncoder(0x02)
p.SetBytesValue(buf)
res := p.Encode()
// res -> { 0x02, 0x04, 0x79, 0x6F, 0x6D, 0x6F }
```

</details>

## Contributors

[//]: contributor-faces

<a href="https://github.com/figroc"><img src="https://avatars1.githubusercontent.com/u/2026460?v=3" title="figroc" width="60" height="60"></a>

[//]: contributor-faces

## License

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fyomorun%2Fy3.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fyomorun%2Fy3?ref=badge_large)
