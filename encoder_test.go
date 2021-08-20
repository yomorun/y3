package y3

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncoderWriteTagErrorSeqID(t *testing.T) {
	enc := &encoder{
		seqID: 0x40,
	}
	assert.PanicsWithError(t, "sid should be in [0..0x3F]", enc.writeTag)
}

func TestEncoderWriteTagIsNode(t *testing.T) {
	enc := &encoder{
		seqID:  0x00,
		isNode: true,
		buf:    new(bytes.Buffer),
	}
	enc.writeTag()
	assert.EqualValues(t, 0x80, enc.seqID)
}

func TestEncoderWriteTagIsPrimitive(t *testing.T) {
	enc := &encoder{
		seqID: 0x00,
		buf:   new(bytes.Buffer),
	}
	enc.writeTag()
	assert.EqualValues(t, 0x00, enc.seqID)
}

func TestEncoderWriteTagIsSlice(t *testing.T) {
	enc := &encoder{
		seqID:   0x00,
		isArray: true,
		buf:     new(bytes.Buffer),
	}
	enc.writeTag()
	assert.EqualValues(t, 0x40, enc.seqID)
}
