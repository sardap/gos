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

func TestZeroPageX(t *testing.T) {
	c := createCpu()

	// Wraping
	c.Registers.PC = 0x1000
	c.Registers.X = 0x8A
	c.Memory.WriteUint16At(0x1001, 0xFF)

	assert.Equal(t, uint16(0x89), c.GetOprandAddress(cpu.AddressModeZeroPageX))

	// Normal
	c.Registers.PC = 0x1000
	c.Registers.X = 0x8A
	c.Memory.WriteUint16At(0x1001, 0x01)

	assert.Equal(t, uint16(0x8B), c.GetOprandAddress(cpu.AddressModeZeroPageX))

}

func TestInderect(t *testing.T) {
	c := createCpu()

	// 6502 wraping bug https://atariage.com/forums/topic/72382-6502-indirect-addressing-ff-behavior/
	c.Registers.PC = 0
	c.Memory.WriteUint16At(0x01, 0x02FF)
	c.Memory.WriteByteAt(0x02FF, 0x00)
	c.Memory.WriteByteAt(0x0200, 0x03)

	assert.Equal(t, uint16(0x0300), c.GetOprandAddress(cpu.AddressModeIndirect))

	// Normal
	c.Registers.PC = 0
	c.Memory.WriteUint16At(0x01, 0x02FF)
	c.Memory.WriteByteAt(0x03FE, 0x00)
	c.Memory.WriteByteAt(0x03FF, 0x03)

	assert.Equal(t, uint16(0x0300), c.GetOprandAddress(cpu.AddressModeIndirect))
}

func TestInderectY(t *testing.T) {
	c := createCpu()

	// Inderect
	c.Registers.Y = 0xFF
	c.Registers.PC = 0
	c.Memory.WriteByteAt(0x01, 0xFF)
	c.Memory.WriteByteAt(0xFF, 0x46)
	c.Memory.WriteByteAt(0x00, 0x01)
	c.Memory.WriteUint16At(0x0146, 0x0245)
	c.Memory.WriteByteAt(0x0245, 0x12)

	assert.Equal(t, byte(0x12), c.ReadByteByMode(cpu.AddressModeIndirectY))
}

func TestInderectXWrap(t *testing.T) {
	c := createCpu()

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
