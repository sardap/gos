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
		address := c.GetOprandAddress(test.mode)
		c.Registers.P.Write(0)

		c.Memory.WriteByte(address, 1)
		c.Registers.A = 2

		cpu.Adc(c, address)

		assert.Equalf(t, uint8(3), c.Registers.A, "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagCarry), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagOverflow), "Address Mode %s", test.mode.String())

		//Overflow
		c.Memory.WriteByte(address, 0b11110001)
		c.Registers.A = 0b00001111

		cpu.Adc(c, address)

		assert.Equalf(t, uint8(0x00), c.Registers.A, "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s %d", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagCarry), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagOverflow), "Address Mode %s", test.mode.String())

		//Add with carry
		c.Memory.WriteByte(address, 1)
		c.Registers.A = 2

		cpu.Adc(c, address)

		assert.Equalf(t, uint8(4), c.Registers.A, "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagCarry), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagOverflow), "Address Mode %s", test.mode.String())
	}
}
