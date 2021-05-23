package cpu_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPushPopUint16(t *testing.T) {
	c := createCpu()

	// Normal case
	c.PushUint16(0x1312)
	assert.Equal(t, uint16(0x1312), c.PopUint16())

	// Fucked case
	c.PushUint16(0x1312)
	assert.Equal(t, byte(0x12), c.PopByte())
	assert.Equal(t, byte(0x13), c.PopByte())
}
