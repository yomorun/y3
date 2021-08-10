package y3

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNodeTag(t *testing.T) {
	var expected byte = 0x81
	tag := NewTag(expected)
	assert.True(t, tag.IsNode())
	assert.False(t, tag.IsSlice())
	assert.EqualValues(t, expected, tag.Raw())
	assert.Equal(t, byte(0x01), tag.SeqID())
}

func TestSliceTag(t *testing.T) {
	var expected byte = 0x42
	tag := NewTag(expected)
	assert.False(t, tag.IsNode())
	assert.True(t, tag.IsSlice())
	assert.EqualValues(t, expected, tag.Raw())
	assert.Equal(t, byte(0x02), tag.SeqID())
}
