package cpu_test

import (
	"testing"

	"github.com/sardap/gos/cpu"
	"github.com/sardap/gos/memory"
	"github.com/stretchr/testify/assert"
)

func createCpu() *cpu.Cpu {
	return cpu.CreateCpu(memory.Create())
}

func writeByteToAddress(c *cpu.Cpu, mode cpu.AddressMode, val byte) {
	switch mode {
	case cpu.AddressModeAccumulator:
		c.Registers.A = val
	default:
		{
			switch mode {
			case cpu.AddressModeImmediate:
				c.Registers.PC = 0
				c.Memory.WriteByte(1, val)

			case cpu.AddressModeZeroPage:
				c.Registers.PC = 0
				c.Memory.WriteByte(1, 30)
				c.Memory.WriteByte(30, val)

			case cpu.AddressModeZeroPageX:
				c.Registers.PC = 0
				c.Registers.X = 5
				c.Memory.WriteByte(1, 30)
				c.Memory.WriteByte(35, val)

			case cpu.AddressModeAbsolute:
				c.Registers.PC = 0
				c.Memory.WriteShort(1, 300)
				c.Memory.WriteByte(300, val)

			case cpu.AddressModeAbsoluteX:
				c.Registers.PC = 0
				c.Registers.X = 5
				c.Memory.WriteShort(1, 300)
				c.Memory.WriteByte(305, val)

			case cpu.AddressModeAbsoluteY:
				c.Registers.PC = 0
				c.Registers.Y = 10
				c.Memory.WriteShort(1, 300)
				c.Memory.WriteByte(310, val)

			case cpu.AddressModeIndirectX:
				c.Registers.PC = 0
				c.Registers.X = 10
				c.Memory.WriteByte(1, 20)
				c.Memory.WriteShort(30, 2048)
				c.Memory.WriteByte(2048, val)

			case cpu.AddressModeIndirectY:
				c.Registers.PC = 0
				c.Registers.Y = 20
				c.Memory.WriteByte(1, 20)
				c.Memory.WriteShort(20, 1028)
				c.Memory.WriteByte(1048, val)
			}
		}
	}
}

func TestAdc(t *testing.T) {
	t.Parallel()

	c := createCpu()

	testCases := []struct {
		mode cpu.AddressMode
	}{
		{mode: cpu.AddressModeImmediate},
		{mode: cpu.AddressModeZeroPage},
		{mode: cpu.AddressModeZeroPageX},
		{mode: cpu.AddressModeAbsolute},
		{mode: cpu.AddressModeAbsoluteX},
		{mode: cpu.AddressModeAbsoluteY},
		{mode: cpu.AddressModeIndirectX},
		{mode: cpu.AddressModeIndirectY},
	}

	for _, test := range testCases {
		c.Registers.P.Write(0)

		writeByteToAddress(c, test.mode, 1)
		c.Registers.A = 2

		cpu.Adc(c, test.mode)

		assert.Equalf(t, uint8(3), c.Registers.A, "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagCarry), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagOverflow), "Address Mode %s", test.mode.String())

		//Overflow
		writeByteToAddress(c, test.mode, 0b11110001)
		c.Registers.A = 0b00001111

		cpu.Adc(c, test.mode)

		assert.Equalf(t, uint8(0x00), c.Registers.A, "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s %d", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagCarry), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagOverflow), "Address Mode %s", test.mode.String())

		//Add with carry
		writeByteToAddress(c, test.mode, 1)
		c.Registers.A = 2

		cpu.Adc(c, test.mode)

		assert.Equalf(t, uint8(4), c.Registers.A, "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagCarry), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagOverflow), "Address Mode %s", test.mode.String())
	}
}

func TestAnd(t *testing.T) {
	t.Parallel()

	c := createCpu()

	testCases := []struct {
		mode cpu.AddressMode
	}{
		{mode: cpu.AddressModeImmediate},
		{mode: cpu.AddressModeZeroPage},
		{mode: cpu.AddressModeZeroPageX},
		{mode: cpu.AddressModeAbsolute},
		{mode: cpu.AddressModeAbsoluteX},
		{mode: cpu.AddressModeAbsoluteY},
		{mode: cpu.AddressModeIndirectX},
		{mode: cpu.AddressModeIndirectY},
	}

	for _, test := range testCases {
		address := c.GetOprandAddress(test.mode)
		c.Registers.P.Write(0)

		c.Memory.WriteByte(address, 0b00001111)
		c.Registers.A = 0b10000101

		cpu.And(c, test.mode)

		assert.Equalf(t, uint8(0b00000101), c.Registers.A, "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())
	}
}

func TestAsl(t *testing.T) {
	t.Parallel()

	c := createCpu()

	testCases := []struct {
		mode cpu.AddressMode
	}{
		{mode: cpu.AddressModeAccumulator},
		{mode: cpu.AddressModeZeroPage},
		{mode: cpu.AddressModeZeroPageX},
		{mode: cpu.AddressModeAbsolute},
		{mode: cpu.AddressModeAbsoluteX},
	}

	for _, test := range testCases {
		writeByteToAddress(c, test.mode, 0b00000001)

		cpu.Asl(c, test.mode)

		assert.Equalf(t, uint8(0b00000010), c.Registers.A, "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())
	}
}
