package cpu_test

import (
	"testing"

	"github.com/sardap/gos/cpu"
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

func TestInderectWrap(t *testing.T) {
	c := createCpu()

	// Inderect
	c.Registers.A = 0
	c.Registers.PC = 0
	c.Memory.WriteByteAt(0x01, 0xFF)
	c.Memory.WriteByteAt(0xFF, 0x5D)
	c.Memory.WriteByteAt(0x5D, 0x04)

	assert.Equal(t, byte(0x04), c.ReadByteByMode(cpu.AddressModeIndirect))

	// X
	c.Registers.X = 0x0A
	c.Registers.A = 0
	c.Registers.PC = 0
	c.Memory.WriteByteAt(0x01, 0xF5)
	c.Memory.WriteByteAt(0xFF, 0x00)
	c.Memory.WriteByteAt(0x00, 0x04)
	c.Memory.WriteByteAt(0x0400, 0x5D)

	assert.Equal(t, byte(0x5D), c.ReadByteByMode(cpu.AddressModeIndirectX))

	// X
	c.Registers.X = 0x81
	c.Registers.A = 0
	c.Registers.PC = 0
	c.Memory.WriteByteAt(0x01, 0xFF)
	c.Memory.WriteByteAt(0x80, 0x00)
	c.Memory.WriteByteAt(0x81, 0x02)
	c.Memory.WriteByteAt(0x0200, 0x5A)

	assert.Equal(t, byte(0x5A), c.ReadByteByMode(cpu.AddressModeIndirectX))
}
