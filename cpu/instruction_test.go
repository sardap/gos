package cpu_test

import (
	"fmt"
	"testing"

	"github.com/sardap/gos/cpu"
	"github.com/sardap/gos/memory"
	"github.com/stretchr/testify/assert"
)

func createCpu() *cpu.Cpu {
	return cpu.CreateCpu(memory.Create())
}

type writeFunc func(c *cpu.Cpu, mode cpu.AddressMode, val byte)
type readFunc func(c *cpu.Cpu, mode cpu.AddressMode) byte

func writeX(c *cpu.Cpu, _ cpu.AddressMode, val byte) {
	c.Registers.X = val
}

func readX(c *cpu.Cpu, _ cpu.AddressMode) byte {
	return c.Registers.X
}

func writeY(c *cpu.Cpu, _ cpu.AddressMode, val byte) {
	c.Registers.Y = val
}

func readY(c *cpu.Cpu, _ cpu.AddressMode) byte {
	return c.Registers.Y
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

			case cpu.AddressModeIndirect:
				c.Registers.PC = 0
				c.Memory.WriteShort(1, 2048)
				c.Memory.WriteByte(2048, val)

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

func writeUint16ToAddress(c *cpu.Cpu, mode cpu.AddressMode, val uint16) {
	switch mode {
	case cpu.AddressModeAccumulator:
		panic(fmt.Errorf("cannot set accumulator to a uint16"))
	default:
		{
			switch mode {
			case cpu.AddressModeImmediate:
				c.Registers.PC = 0
				c.Memory.WriteShort(1, val)

			case cpu.AddressModeZeroPage:
				c.Registers.PC = 0
				c.Memory.WriteByte(1, 30)
				c.Memory.WriteShort(30, val)

			case cpu.AddressModeZeroPageX:
				c.Registers.PC = 0
				c.Registers.X = 5
				c.Memory.WriteByte(1, 30)
				c.Memory.WriteShort(35, val)

			case cpu.AddressModeAbsolute:
				c.Registers.PC = 0
				c.Memory.WriteShort(1, 300)
				c.Memory.WriteShort(300, val)

			case cpu.AddressModeAbsoluteX:
				c.Registers.PC = 0
				c.Registers.X = 5
				c.Memory.WriteShort(1, 300)
				c.Memory.WriteShort(305, val)

			case cpu.AddressModeAbsoluteY:
				c.Registers.PC = 0
				c.Registers.Y = 10
				c.Memory.WriteShort(1, 300)
				c.Memory.WriteShort(310, val)

			case cpu.AddressModeIndirect:
				c.Registers.PC = 0
				c.Memory.WriteShort(1, 2048)
				c.Memory.WriteShort(2048, val)

			case cpu.AddressModeIndirectX:
				c.Registers.PC = 0
				c.Registers.X = 10
				c.Memory.WriteByte(1, 20)
				c.Memory.WriteShort(30, 2048)
				c.Memory.WriteShort(2048, val)

			case cpu.AddressModeIndirectY:
				c.Registers.PC = 0
				c.Registers.Y = 20
				c.Memory.WriteByte(1, 20)
				c.Memory.WriteShort(20, 1028)
				c.Memory.WriteShort(1048, val)
			}
		}
	}
}

func readByteFromAddress(c *cpu.Cpu, mode cpu.AddressMode) byte {
	switch mode {
	case cpu.AddressModeAccumulator:
		return c.Registers.A
	default:
		return c.Memory.ReadByte(c.GetOprandAddress(mode))
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

func TestBranchFlags(t *testing.T) {
	t.Parallel()

	c := createCpu()

	testCases := []struct {
		inscut cpu.Instruction
		flag   cpu.Flag
		valid  bool
	}{
		{inscut: cpu.Bcc, flag: cpu.FlagCarry, valid: false},
		{inscut: cpu.Bcs, flag: cpu.FlagCarry, valid: true},
		{inscut: cpu.Bne, flag: cpu.FlagZero, valid: false},
		{inscut: cpu.Beq, flag: cpu.FlagZero, valid: true},
		{inscut: cpu.Bpl, flag: cpu.FlagNegative, valid: false},
		{inscut: cpu.Bmi, flag: cpu.FlagNegative, valid: true},
		{inscut: cpu.Bvc, flag: cpu.FlagOverflow, valid: false},
		{inscut: cpu.Bvs, flag: cpu.FlagOverflow, valid: true},
	}

	for _, test := range testCases {
		//Branch postive on same page
		c.Registers.PC = 50
		c.ExtraTicks = 0
		c.Memory.WriteByte(c.Registers.PC+1, 0b00000011)
		c.Registers.P.SetFlag(test.flag, test.valid)

		test.inscut(c, cpu.AddressModeRelative)

		assert.Equal(t, uint16(53), c.Registers.PC)
		assert.Equal(t, uint8(1), c.ExtraTicks)

		//Branch Negtaive on same page
		c.Registers.PC = 50
		c.ExtraTicks = 0
		c.Memory.WriteByte(c.Registers.PC+1, 0b11111101)
		c.Registers.P.SetFlag(test.flag, test.valid)

		test.inscut(c, cpu.AddressModeRelative)

		assert.Equal(t, uint16(47), c.Registers.PC)
		assert.Equal(t, uint8(1), c.ExtraTicks)

		//Branch to a new page
		c.Registers.PC = 129
		c.ExtraTicks = 0
		c.Memory.WriteByte(c.Registers.PC+1, 127)
		c.Registers.P.SetFlag(test.flag, test.valid)

		test.inscut(c, cpu.AddressModeRelative)

		assert.Equal(t, uint16(256), c.Registers.PC)
		assert.Equal(t, uint8(2), c.ExtraTicks)

		//Don't branch to a new page
		c.Registers.PC = 5
		c.ExtraTicks = 0
		c.Memory.WriteByte(c.Registers.PC+1, 5)
		c.Registers.P.SetFlag(test.flag, !test.valid)

		test.inscut(c, cpu.AddressModeRelative)

		assert.Equal(t, uint16(7), c.Registers.PC)
		assert.Equal(t, uint8(0), c.ExtraTicks)
	}
}

func TestBit(t *testing.T) {
	t.Parallel()

	c := createCpu()

	testCases := []struct {
		mode cpu.AddressMode
	}{
		{mode: cpu.AddressModeZeroPage},
		{mode: cpu.AddressModeAbsolute},
	}

	for _, test := range testCases {
		// Not zero
		writeByteToAddress(c, test.mode, 0b00100000)
		c.Registers.A = 0b00100000

		cpu.Bit(c, test.mode)

		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagOverflow), "Address Mode %s", test.mode.String())

		// All zero
		writeByteToAddress(c, test.mode, 0b00000000)
		c.Registers.A = 0b00000000

		cpu.Bit(c, test.mode)

		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagOverflow), "Address Mode %s", test.mode.String())

		// Neg set
		writeByteToAddress(c, test.mode, 0b10000000)
		c.Registers.A = 0b00000000

		cpu.Bit(c, test.mode)

		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagOverflow), "Address Mode %s", test.mode.String())

		// Overflow set
		writeByteToAddress(c, test.mode, 0b01000000)
		c.Registers.A = 0b00000000

		cpu.Bit(c, test.mode)

		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagOverflow), "Address Mode %s", test.mode.String())

		// Neg, overflow set
		writeByteToAddress(c, test.mode, 0b11000000)
		c.Registers.A = 0b11000000

		cpu.Bit(c, test.mode)

		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagOverflow), "Address Mode %s", test.mode.String())

		// All set
		writeByteToAddress(c, test.mode, 0b11000000)
		c.Registers.A = 0b00000000

		cpu.Bit(c, test.mode)

		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagOverflow), "Address Mode %s", test.mode.String())
	}
}

func TestFlagClears(t *testing.T) {
	t.Parallel()

	c := createCpu()

	testCases := []struct {
		flag cpu.Flag
		inst cpu.Instruction
	}{
		{inst: cpu.Clc, flag: cpu.FlagCarry},
		{inst: cpu.Cld, flag: cpu.FlagDecimal},
		{inst: cpu.Cli, flag: cpu.FlagInteruprtDisable},
		{inst: cpu.Clv, flag: cpu.FlagOverflow},
	}

	flags := []cpu.Flag{
		cpu.FlagCarry,
		cpu.FlagDecimal,
		cpu.FlagInteruprtDisable,
		cpu.FlagOverflow,
	}

	for _, test := range testCases {
		// Clear all flags
		c.Registers.P.Write(0xFF)

		test.inst(c, cpu.AddressModeImplied)

		for _, flag := range flags {
			var expected bool
			if flag == test.flag {
				expected = false
			} else {
				expected = true
			}

			assert.Equal(t, expected, c.Registers.P.ReadFlag(flag))
		}
	}
}

func TestCompares(t *testing.T) {
	t.Parallel()

	c := createCpu()

	testCases := []struct {
		mode cpu.AddressMode
		reg  *uint8
		inst cpu.Instruction
	}{
		// Cmp
		{reg: &c.Registers.A, inst: cpu.Cmp, mode: cpu.AddressModeImmediate},
		{reg: &c.Registers.A, inst: cpu.Cmp, mode: cpu.AddressModeZeroPage},
		{reg: &c.Registers.A, inst: cpu.Cmp, mode: cpu.AddressModeZeroPageX},
		{reg: &c.Registers.A, inst: cpu.Cmp, mode: cpu.AddressModeAbsolute},
		{reg: &c.Registers.A, inst: cpu.Cmp, mode: cpu.AddressModeAbsoluteX},
		{reg: &c.Registers.A, inst: cpu.Cmp, mode: cpu.AddressModeAbsoluteY},
		{reg: &c.Registers.A, inst: cpu.Cmp, mode: cpu.AddressModeIndirectX},
		{reg: &c.Registers.A, inst: cpu.Cmp, mode: cpu.AddressModeIndirectY},
		// Cpx
		{reg: &c.Registers.X, inst: cpu.Cpx, mode: cpu.AddressModeImmediate},
		{reg: &c.Registers.X, inst: cpu.Cpx, mode: cpu.AddressModeZeroPage},
		{reg: &c.Registers.X, inst: cpu.Cpx, mode: cpu.AddressModeAbsolute},
		// Cpy
		{reg: &c.Registers.Y, inst: cpu.Cpy, mode: cpu.AddressModeImmediate},
		{reg: &c.Registers.Y, inst: cpu.Cpy, mode: cpu.AddressModeZeroPage},
		{reg: &c.Registers.Y, inst: cpu.Cpy, mode: cpu.AddressModeAbsolute},
	}

	for _, test := range testCases {
		// Check doesn't change anything
		writeByteToAddress(c, test.mode, 3)
		*test.reg = 5

		test.inst(c, test.mode)

		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagCarry), "Address Mode %s", test.mode.String())

		// Check Zero flag
		writeByteToAddress(c, test.mode, 5)
		*test.reg = 5

		test.inst(c, test.mode)

		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagCarry), "Address Mode %s", test.mode.String())

		// Check Carry flag
		writeByteToAddress(c, test.mode, 6)
		*test.reg = 5

		test.inst(c, test.mode)

		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagCarry), "Address Mode %s", test.mode.String())

		// Check Negative numbers
		writeByteToAddress(c, test.mode, 0x81)
		*test.reg = 0x85

		test.inst(c, test.mode)

		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagCarry), "Address Mode %s", test.mode.String())
	}
}

func TestDec(t *testing.T) {
	t.Parallel()

	c := createCpu()

	testCases := []struct {
		mode  cpu.AddressMode
		write writeFunc
		read  readFunc
		inst  cpu.Instruction
	}{
		{inst: cpu.Dec, write: writeByteToAddress, read: readByteFromAddress, mode: cpu.AddressModeZeroPage},
		{inst: cpu.Dec, write: writeByteToAddress, read: readByteFromAddress, mode: cpu.AddressModeZeroPageX},
		{inst: cpu.Dec, write: writeByteToAddress, read: readByteFromAddress, mode: cpu.AddressModeAbsolute},
		{inst: cpu.Dec, write: writeByteToAddress, read: readByteFromAddress, mode: cpu.AddressModeAbsoluteX},
		{inst: cpu.Dex, write: writeX, read: readX, mode: cpu.AddressModeImplied},
		{inst: cpu.Dey, write: writeY, read: readY, mode: cpu.AddressModeImplied},
	}

	for _, test := range testCases {
		// basic dec
		test.write(c, test.mode, 3)

		test.inst(c, test.mode)

		assert.Equalf(t, byte(2), test.read(c, test.mode), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())

		// Wrap
		test.write(c, test.mode, 0)

		test.inst(c, test.mode)

		assert.Equalf(t, byte(0xFF), test.read(c, test.mode), "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())

		// Zero
		test.write(c, test.mode, 1)

		test.inst(c, test.mode)

		assert.Equalf(t, byte(0), test.read(c, test.mode), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())
	}
}

func TestEor(t *testing.T) {
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

		// Zero
		writeByteToAddress(c, test.mode, 0b00000000)
		c.Registers.A = 0b00000000

		cpu.Eor(c, test.mode)

		assert.Equalf(t, uint8(0), c.Registers.A, "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())

		// Zero
		writeByteToAddress(c, test.mode, 0b11111111)
		c.Registers.A = 0b11111111

		cpu.Eor(c, test.mode)

		assert.Equalf(t, uint8(0), c.Registers.A, "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())

		// Negative
		writeByteToAddress(c, test.mode, 0b01111111)
		c.Registers.A = 0b11111111

		cpu.Eor(c, test.mode)

		assert.Equalf(t, uint8(0b10000000), c.Registers.A, "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())

		// Good old Xor
		writeByteToAddress(c, test.mode, 0b00011010)
		c.Registers.A = 0b01010010

		cpu.Eor(c, test.mode)

		assert.Equalf(t, uint8(0b01001000), c.Registers.A, "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())
	}
}

func TestInc(t *testing.T) {
	t.Parallel()

	c := createCpu()

	testCases := []struct {
		mode  cpu.AddressMode
		write writeFunc
		read  readFunc
		inst  cpu.Instruction
	}{
		{inst: cpu.Inc, write: writeByteToAddress, read: readByteFromAddress, mode: cpu.AddressModeZeroPage},
		{inst: cpu.Inc, write: writeByteToAddress, read: readByteFromAddress, mode: cpu.AddressModeZeroPageX},
		{inst: cpu.Inc, write: writeByteToAddress, read: readByteFromAddress, mode: cpu.AddressModeAbsolute},
		{inst: cpu.Inc, write: writeByteToAddress, read: readByteFromAddress, mode: cpu.AddressModeAbsoluteX},
		{inst: cpu.Inx, write: writeX, read: readX, mode: cpu.AddressModeImplied},
		{inst: cpu.Iny, write: writeY, read: readY, mode: cpu.AddressModeImplied},
	}

	for _, test := range testCases {
		// basic dec
		test.write(c, test.mode, 3)

		test.inst(c, test.mode)

		assert.Equalf(t, byte(4), test.read(c, test.mode), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())

		// Wrap
		test.write(c, test.mode, 0xFF)

		test.inst(c, test.mode)

		assert.Equalf(t, byte(0x00), test.read(c, test.mode), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())
	}
}

func TestJmp(t *testing.T) {
	t.Parallel()

	c := createCpu()

	testCases := []struct {
		mode cpu.AddressMode
	}{
		{mode: cpu.AddressModeAbsolute},
		{mode: cpu.AddressModeIndirect},
	}

	for _, test := range testCases {
		// Zero
		writeUint16ToAddress(c, test.mode, 0x1312)

		cpu.Jmp(c, test.mode)

		assert.Equalf(t, uint16(0x1312), c.Registers.PC, "Address Mode %s", test.mode.String())
	}
}
