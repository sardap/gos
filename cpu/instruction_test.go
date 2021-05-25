package cpu_test

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"

	"github.com/sardap/gos/cpu"
	"github.com/sardap/gos/memory"
	"github.com/sardap/gos/ppu"
	"github.com/stretchr/testify/assert"
)

type testCart struct {
	data [0x10000]byte
}

func (c *testCart) WriteBytesPrg(value []byte) error {
	return nil
}

func (c *testCart) WriteBytesChr(value []byte) error {
	return nil
}

func (c *testCart) WriteByteAt(address uint16, value byte) {
	c.data[address] = value
}

func (c *testCart) ReadByteAt(address uint16) byte {
	return c.data[address]
}

func createCpu() *cpu.Cpu {
	result := cpu.CreateCpu(memory.Create(), ppu.Create())
	result.Memory.SetCart(&testCart{})
	return result
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
				c.Memory.WriteByteAt(1, val)

			case cpu.AddressModeZeroPage:
				c.Registers.PC = 0
				c.Memory.WriteByteAt(1, 30)
				c.Memory.WriteByteAt(30, val)

			case cpu.AddressModeZeroPageX:
				c.Registers.PC = 0
				c.Registers.X = 5
				c.Memory.WriteByteAt(1, 30)
				c.Memory.WriteByteAt(35, val)

			case cpu.AddressModeZeroPageY:
				c.Registers.PC = 0
				c.Registers.Y = 15
				c.Memory.WriteByteAt(1, 30)
				c.Memory.WriteByteAt(45, val)

			case cpu.AddressModeAbsolute:
				c.Registers.PC = 0
				c.Memory.WriteUint16At(1, 300)
				c.Memory.WriteByteAt(300, val)

			case cpu.AddressModeAbsoluteX:
				c.Registers.PC = 0
				c.Registers.X = 5
				c.Memory.WriteUint16At(1, 300)
				c.Memory.WriteByteAt(305, val)

			case cpu.AddressModeAbsoluteY:
				c.Registers.PC = 0
				c.Registers.Y = 10
				c.Memory.WriteUint16At(1, 300)
				c.Memory.WriteByteAt(310, val)

			case cpu.AddressModeIndirect:
				c.Registers.PC = 0
				c.Memory.WriteUint16At(1, 2048)
				c.Memory.WriteUint16At(2048, 2050)
				c.Memory.WriteByteAt(2050, val)

			case cpu.AddressModeIndirectX:
				c.Registers.PC = 0
				c.Registers.X = 10
				c.Memory.WriteByteAt(1, 20)
				c.Memory.WriteUint16At(30, 2048)
				c.Memory.WriteByteAt(2048, val)

			case cpu.AddressModeIndirectY:
				c.Registers.PC = 0
				c.Registers.Y = 20
				c.Memory.WriteByteAt(1, 20)
				c.Memory.WriteUint16At(20, 1028)
				c.Memory.WriteByteAt(1048, val)
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
				c.Memory.WriteUint16At(1, val)

			case cpu.AddressModeZeroPage:
				c.Registers.PC = 0
				c.Memory.WriteByteAt(1, 30)
				c.Memory.WriteUint16At(30, val)

			case cpu.AddressModeZeroPageX:
				c.Registers.PC = 0
				c.Registers.X = 5
				c.Memory.WriteByteAt(1, 30)
				c.Memory.WriteUint16At(35, val)

			case cpu.AddressModeAbsolute:
				c.Registers.PC = 0
				c.Memory.WriteUint16At(1, 300)
				c.Memory.WriteUint16At(300, val)

			case cpu.AddressModeAbsoluteX:
				c.Registers.PC = 0
				c.Registers.X = 5
				c.Memory.WriteUint16At(1, 300)
				c.Memory.WriteUint16At(305, val)

			case cpu.AddressModeAbsoluteY:
				c.Registers.PC = 0
				c.Registers.Y = 10
				c.Memory.WriteUint16At(1, 300)
				c.Memory.WriteUint16At(310, val)

			case cpu.AddressModeIndirect:
				c.Registers.PC = 0
				c.Memory.WriteUint16At(c.Registers.PC+1, 1000)
				c.Memory.WriteUint16At(1000, val)

			case cpu.AddressModeIndirectX:
				c.Registers.PC = 0
				c.Registers.X = 10
				c.Memory.WriteByteAt(1, 20)
				c.Memory.WriteUint16At(30, 2048)
				c.Memory.WriteUint16At(2048, val)

			case cpu.AddressModeIndirectY:
				c.Registers.PC = 0
				c.Registers.Y = 20
				c.Memory.WriteByteAt(1, 20)
				c.Memory.WriteUint16At(20, 1028)
				c.Memory.WriteUint16At(1048, val)
			}
		}
	}
}

func readByteFromAddress(c *cpu.Cpu, mode cpu.AddressMode) byte {
	switch mode {
	case cpu.AddressModeAccumulator:
		return c.Registers.A
	default:
		return c.Memory.ReadByteAt(c.GetOprandAddress(mode))
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
		writeByteToAddress(c, test.mode, 0b00001111)
		c.Registers.P.Write(0)
		c.Registers.A = 0b10000101

		cpu.And(c, test.mode)

		assert.Equalf(t, uint8(0b00000101), c.Registers.A, "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())

		writeByteToAddress(c, test.mode, 0xEF)
		c.Registers.P.Write(0)
		c.Registers.A = 0x6F

		cpu.And(c, test.mode)

		assert.Equalf(t, uint8(0x6F), c.Registers.A, "Address Mode %s", test.mode.String())
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

		assert.Equalf(t, uint8(0b00000010), readByteFromAddress(c, test.mode), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagCarry), "Address Mode %s", test.mode.String())

		writeByteToAddress(c, test.mode, 0x80)

		cpu.Asl(c, test.mode)

		assert.Equalf(t, uint8(0x00), readByteFromAddress(c, test.mode), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagCarry), "Address Mode %s", test.mode.String())
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
		c.ExtraCycles = 0
		c.Memory.WriteByteAt(c.Registers.PC+1, 0b00000011)
		c.Registers.P.SetFlag(test.flag, test.valid)

		test.inscut(c, cpu.AddressModeRelative)

		assert.Equal(t, uint16(53), c.Registers.PC)
		assert.Equal(t, uint8(1), c.ExtraCycles)

		//Branch Negtaive on same page
		c.Registers.PC = 50
		c.ExtraCycles = 0
		c.Memory.WriteByteAt(c.Registers.PC+1, 0b11111101)
		c.Registers.P.SetFlag(test.flag, test.valid)

		test.inscut(c, cpu.AddressModeRelative)

		assert.Equal(t, uint16(47), c.Registers.PC)
		assert.Equal(t, uint8(1), c.ExtraCycles)

		//Branch to a new page
		c.Registers.PC = 129
		c.ExtraCycles = 0
		c.Memory.WriteByteAt(c.Registers.PC+1, 127)
		c.Registers.P.SetFlag(test.flag, test.valid)

		test.inscut(c, cpu.AddressModeRelative)

		assert.Equal(t, uint16(256), c.Registers.PC)
		assert.Equal(t, uint8(2), c.ExtraCycles)

		//Don't branch to a new page
		c.Registers.PC = 5
		c.ExtraCycles = 0
		c.Memory.WriteByteAt(c.Registers.PC+1, 5)
		c.Registers.P.SetFlag(test.flag, !test.valid)

		test.inscut(c, cpu.AddressModeRelative)

		assert.Equal(t, uint16(5), c.Registers.PC)
		assert.Equal(t, uint8(0), c.ExtraCycles)
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
		{mode: cpu.AddressModeZeroPageX},
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
	c.Registers.PC = 0

	c.Memory.WriteUint16At(1, 0x1312)

	cpu.Jmp(c, cpu.AddressModeAbsolute)

	assert.Equal(t, uint16(0x1312), c.Registers.PC, "Address Mode ", cpu.AddressModeAbsolute.String())

	// Indirect
	writeUint16ToAddress(c, cpu.AddressModeIndirect, 0x1312)

	cpu.Jmp(c, cpu.AddressModeIndirect)

	assert.Equal(t, uint16(0x1312), c.Registers.PC, "Address Mode ", cpu.AddressModeIndirect.String())
}

func TestJsr(t *testing.T) {
	t.Parallel()

	c := createCpu()
	c.Registers.PC = 0

	c.Memory.WriteUint16At(1, 0x1312)

	cpu.Jsr(c, cpu.AddressModeAbsolute)

	assert.Equal(t, uint16(0x1312), c.Registers.PC)
	assert.Equal(t, uint16(0x0002), c.PopUint16())
}

func TestLd(t *testing.T) {
	t.Parallel()

	c := createCpu()

	testCases := []struct {
		mode       cpu.AddressMode
		extraCycle bool
		inst       cpu.Instruction
		reg        *byte
	}{
		// LDA
		{inst: cpu.Lda, reg: &c.Registers.A, mode: cpu.AddressModeImmediate, extraCycle: false},
		{inst: cpu.Lda, reg: &c.Registers.A, mode: cpu.AddressModeZeroPage, extraCycle: false},
		{inst: cpu.Lda, reg: &c.Registers.A, mode: cpu.AddressModeZeroPageX, extraCycle: false},
		{inst: cpu.Lda, reg: &c.Registers.A, mode: cpu.AddressModeAbsolute, extraCycle: false},
		{inst: cpu.Lda, reg: &c.Registers.A, mode: cpu.AddressModeAbsoluteX, extraCycle: true},
		{inst: cpu.Lda, reg: &c.Registers.A, mode: cpu.AddressModeAbsoluteY, extraCycle: true},
		{inst: cpu.Lda, reg: &c.Registers.A, mode: cpu.AddressModeIndirectX, extraCycle: false},
		{inst: cpu.Lda, reg: &c.Registers.A, mode: cpu.AddressModeIndirectY, extraCycle: true},
		// LDX
		{inst: cpu.Ldx, reg: &c.Registers.X, mode: cpu.AddressModeImmediate, extraCycle: false},
		{inst: cpu.Ldx, reg: &c.Registers.X, mode: cpu.AddressModeZeroPage, extraCycle: false},
		{inst: cpu.Ldx, reg: &c.Registers.X, mode: cpu.AddressModeZeroPageY, extraCycle: false},
		{inst: cpu.Ldx, reg: &c.Registers.X, mode: cpu.AddressModeAbsolute, extraCycle: false},
		{inst: cpu.Ldx, reg: &c.Registers.X, mode: cpu.AddressModeAbsoluteY, extraCycle: true},
		// LDY
		{inst: cpu.Ldy, reg: &c.Registers.Y, mode: cpu.AddressModeImmediate, extraCycle: false},
		{inst: cpu.Ldy, reg: &c.Registers.Y, mode: cpu.AddressModeZeroPage, extraCycle: false},
		{inst: cpu.Ldy, reg: &c.Registers.Y, mode: cpu.AddressModeZeroPageX, extraCycle: false},
		{inst: cpu.Ldy, reg: &c.Registers.Y, mode: cpu.AddressModeAbsolute, extraCycle: false},
		{inst: cpu.Ldy, reg: &c.Registers.Y, mode: cpu.AddressModeAbsoluteX, extraCycle: true},
	}

	for _, test := range testCases {
		c.Registers.P.Write(0)

		// Clean load
		c.ExtraCycles = 0
		writeByteToAddress(c, test.mode, 0b00100000)
		*test.reg = 10

		test.inst(c, test.mode)

		assert.Equalf(t, byte(0b00100000), *test.reg, "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())
		if test.extraCycle {
			assert.Equalf(t, byte(1), c.ExtraCycles, "Address Mode %s", test.mode.String())
		}

		// Load zero
		c.ExtraCycles = 0
		writeByteToAddress(c, test.mode, 0b00000000)
		*test.reg = 10

		test.inst(c, test.mode)

		assert.Equalf(t, byte(0b00000000), *test.reg, "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())
		if test.extraCycle {
			assert.Equalf(t, byte(1), c.ExtraCycles, "Address Mode %s", test.mode.String())
		}

		// Neg load
		c.ExtraCycles = 0
		writeByteToAddress(c, test.mode, 0b10000000)
		*test.reg = 10

		test.inst(c, test.mode)

		assert.Equalf(t, byte(0b10000000), *test.reg, "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())
		if test.extraCycle {
			assert.Equalf(t, byte(1), c.ExtraCycles, "Address Mode %s", test.mode.String())
		}
	}
}

func TestLsr(t *testing.T) {
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
		writeByteToAddress(c, test.mode, 0b00000010)

		cpu.Lsr(c, test.mode)

		assert.Equalf(t, uint8(0b00000001), readByteFromAddress(c, test.mode), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagCarry), "Address Mode %s", test.mode.String())

		// Carry
		cpu.Lsr(c, test.mode)

		assert.Equalf(t, uint8(0b00000000), readByteFromAddress(c, test.mode), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagCarry), "Address Mode %s", test.mode.String())
	}
}

func TestNop(t *testing.T) {
	t.Parallel()

	c := createCpu()

	cpu.Nop(c, cpu.AddressModeImplied)

	assert.Equal(t, byte(cpu.StartingA), c.Registers.A)
	assert.Equal(t, byte(cpu.StartingX), c.Registers.X)
	assert.Equal(t, byte(cpu.StartingY), c.Registers.Y)
	assert.Equal(t, byte(cpu.StartingP), c.Registers.P.Read())
	assert.Equal(t, byte(cpu.StartingSP), c.Registers.SP)
	assert.Equal(t, uint16(cpu.StartingPC), c.Registers.PC)
}

func TestOra(t *testing.T) {
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
		writeByteToAddress(c, test.mode, 0b01000010)
		c.Registers.A = 0b01100000

		cpu.Ora(c, test.mode)

		assert.Equalf(t, uint8(0x62), c.Registers.A, "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())

		// Neg
		writeByteToAddress(c, test.mode, 0b11000010)
		c.Registers.A = 0b01100000

		cpu.Ora(c, test.mode)

		assert.Equalf(t, uint8(0xE2), c.Registers.A, "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())

		// Zero
		writeByteToAddress(c, test.mode, 0)
		c.Registers.A = 0

		cpu.Ora(c, test.mode)

		assert.Equalf(t, byte(0), c.Registers.A, "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())

	}
}

func TestPushies(t *testing.T) {
	t.Parallel()

	c := createCpu()

	// Pha
	c.Registers.A = 0x13

	cpu.Pha(c, cpu.AddressModeImplied)

	assert.Equal(t, byte(0x13), c.PopByte())

	// Php
	c.Registers.P.Write(0x6f)

	cpu.Php(c, cpu.AddressModeImplied)

	assert.Equal(t, byte(0x7f), c.PopByte())
}

func TestPulls(t *testing.T) {
	t.Parallel()

	// Pla
	c := createCpu()

	c.PushByte(0x13)

	cpu.Pla(c, cpu.AddressModeImplied)

	assert.Equal(t, byte(0x13), c.Registers.A)

	// Plp
	c.PushByte(0xcf)

	cpu.Plp(c, cpu.AddressModeImplied)

	assert.Equal(t, byte(0xef), c.Registers.P.Read())

	// Edge cases

	c.PushUint16(0xCE39)

	cpu.Pla(c, cpu.AddressModeImplied)
	assert.Equal(t, byte(0x39), c.Registers.A)
}

func TestRor(t *testing.T) {
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

		writeByteToAddress(c, test.mode, 0b10000000)

		c.Registers.P.Write(0)
		cpu.Ror(c, test.mode)

		assert.Equalf(t, uint8(0b01000000), readByteFromAddress(c, test.mode), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagCarry), "Address Mode %s", test.mode.String())

		// Zero
		c.Registers.P.Write(0)
		writeByteToAddress(c, test.mode, 0b00000000)

		cpu.Ror(c, test.mode)

		assert.Equalf(t, uint8(0b00000000), readByteFromAddress(c, test.mode), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagCarry), "Address Mode %s", test.mode.String())

		// FF
		c.Registers.P.Write(0)
		writeByteToAddress(c, test.mode, 0xFF)

		cpu.Ror(c, test.mode)

		assert.Equalf(t, uint8(0x7F), readByteFromAddress(c, test.mode), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagCarry), "Address Mode %s", test.mode.String())

		// 1
		c.Registers.P.Write(0)
		writeByteToAddress(c, test.mode, 0x01)

		cpu.Ror(c, test.mode)

		assert.Equalf(t, uint8(0x00), readByteFromAddress(c, test.mode), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagCarry), "Address Mode %s", test.mode.String())

	}
}

func TestRol(t *testing.T) {
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

		c.Registers.P.Write(0)
		writeByteToAddress(c, test.mode, 0b10000000)

		cpu.Rol(c, test.mode)

		assert.Equalf(t, uint8(0b00000000), readByteFromAddress(c, test.mode), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagCarry), "Address Mode %s", test.mode.String())

		// Zero
		c.Registers.P.Write(0)
		writeByteToAddress(c, test.mode, 0b00000000)

		cpu.Rol(c, test.mode)

		assert.Equalf(t, uint8(0b00000000), readByteFromAddress(c, test.mode), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagCarry), "Address Mode %s", test.mode.String())

		// FF
		c.Registers.P.Write(0)
		writeByteToAddress(c, test.mode, 0xFF)

		cpu.Rol(c, test.mode)

		assert.Equalf(t, uint8(0xFE), readByteFromAddress(c, test.mode), "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagCarry), "Address Mode %s", test.mode.String())
	}
}

func TestRti(t *testing.T) {
	t.Parallel()

	c := createCpu()

	c.PushUint16(0x1312)
	c.PushByte(0xFF)

	cpu.Rti(c, cpu.AddressModeImplied)

	assert.Equal(t, byte(0xEF), c.Registers.P.Read())
	assert.Equal(t, uint16(0x1312), c.Registers.PC)
}

func TestRts(t *testing.T) {
	t.Parallel()

	c := createCpu()

	c.PushUint16(0x1311)

	cpu.Rts(c, cpu.AddressModeImplied)

	assert.Equal(t, uint16(0x1312), c.Registers.PC)
}

func TestSbc(t *testing.T) {
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

		cpu.Sbc(c, test.mode)

		assert.Equalf(t, uint8(0), c.Registers.A, "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagCarry), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagOverflow), "Address Mode %s", test.mode.String())

		c.Registers.P.Write(0)
		writeByteToAddress(c, test.mode, 0xF1)
		c.Registers.A = 0x0F

		cpu.Sbc(c, test.mode)

		assert.Equalf(t, uint8(0x1D), c.Registers.A, "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s %d", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagCarry), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagOverflow), "Address Mode %s", test.mode.String())
	}
}

func TestFlagSets(t *testing.T) {
	t.Parallel()

	c := createCpu()

	testCases := []struct {
		flag cpu.Flag
		inst cpu.Instruction
	}{
		{inst: cpu.Sec, flag: cpu.FlagCarry},
		{inst: cpu.Sed, flag: cpu.FlagDecimal},
		{inst: cpu.Sei, flag: cpu.FlagInteruprtDisable},
	}

	flags := []cpu.Flag{
		cpu.FlagCarry,
		cpu.FlagDecimal,
		cpu.FlagInteruprtDisable,
	}

	for _, test := range testCases {
		// Clear all flags
		c.Registers.P.Write(0x00)

		test.inst(c, cpu.AddressModeImplied)

		for _, flag := range flags {
			var expected bool
			if flag == test.flag {
				expected = true
			} else {
				expected = false
			}

			assert.Equal(t, expected, c.Registers.P.ReadFlag(flag))
		}
	}
}

func TestSta(t *testing.T) {
	t.Parallel()

	c := createCpu()
	c.Registers.PC = 0

	testCases := []struct {
		mode cpu.AddressMode
	}{
		{mode: cpu.AddressModeZeroPage},
		{mode: cpu.AddressModeZeroPageX},
		{mode: cpu.AddressModeAbsolute},
		{mode: cpu.AddressModeAbsoluteX},
		{mode: cpu.AddressModeAbsoluteY},
		{mode: cpu.AddressModeIndirectX},
		{mode: cpu.AddressModeIndirectY},
	}

	for _, test := range testCases {
		c.Registers.A = 0x13

		cpu.Sta(c, test.mode)

		assert.Equalf(t, byte(0x13), readByteFromAddress(c, test.mode), "Address Mode %s", test.mode.String())
	}
}

func TestStx(t *testing.T) {
	t.Parallel()

	c := createCpu()
	c.Registers.PC = 0

	testCases := []struct {
		mode cpu.AddressMode
	}{
		{mode: cpu.AddressModeZeroPage},
		{mode: cpu.AddressModeZeroPageY},
		{mode: cpu.AddressModeAbsolute},
	}

	for _, test := range testCases {
		c.Registers.X = 0x13

		cpu.Stx(c, test.mode)

		assert.Equalf(t, byte(0x13), readByteFromAddress(c, test.mode), "Address Mode %s", test.mode.String())
	}
}

func TestSty(t *testing.T) {
	t.Parallel()

	c := createCpu()
	c.Registers.PC = 0

	testCases := []struct {
		mode cpu.AddressMode
	}{
		{mode: cpu.AddressModeZeroPage},
		{mode: cpu.AddressModeZeroPageX},
		{mode: cpu.AddressModeAbsolute},
	}

	for _, test := range testCases {
		c.Registers.Y = 0x13

		cpu.Sty(c, test.mode)

		assert.Equalf(t, byte(0x13), readByteFromAddress(c, test.mode), "Address Mode %s", test.mode.String())
	}
}

func TestTransfers(t *testing.T) {
	t.Parallel()

	c := createCpu()

	testCases := []struct {
		inst       cpu.Instruction
		source     *byte
		target     *byte
		checkFlags bool
	}{
		{inst: cpu.Tax, source: &c.Registers.A, target: &c.Registers.X, checkFlags: true},
		{inst: cpu.Tay, source: &c.Registers.A, target: &c.Registers.Y, checkFlags: true},
		{inst: cpu.Tsx, source: &c.Registers.SP, target: &c.Registers.X, checkFlags: true},
		{inst: cpu.Txa, source: &c.Registers.X, target: &c.Registers.A, checkFlags: true},
		{inst: cpu.Txs, source: &c.Registers.X, target: &c.Registers.SP, checkFlags: false},
		{inst: cpu.Tya, source: &c.Registers.Y, target: &c.Registers.A, checkFlags: true},
	}

	for _, test := range testCases {
		*test.source = 0b10000000
		*test.target = 0b00000000

		test.inst(c, cpu.AddressModeImplied)

		assert.Equal(t, *test.source, *test.target, runtime.FuncForPC(reflect.ValueOf(test.inst).Pointer()).Name())

		if !test.checkFlags {
			continue
		}

		assert.Equal(t, true, c.Registers.P.ReadFlag(cpu.FlagNegative), runtime.FuncForPC(reflect.ValueOf(test.inst).Pointer()).Name())
		assert.Equal(t, false, c.Registers.P.ReadFlag(cpu.FlagZero), runtime.FuncForPC(reflect.ValueOf(test.inst).Pointer()).Name())

		*test.source = 0b00000000
		*test.target = 0b10000000

		test.inst(c, cpu.AddressModeImplied)

		assert.Equal(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), runtime.FuncForPC(reflect.ValueOf(test.inst).Pointer()).Name())
		assert.Equal(t, true, c.Registers.P.ReadFlag(cpu.FlagZero), runtime.FuncForPC(reflect.ValueOf(test.inst).Pointer()).Name())
	}
}

func TestLax(t *testing.T) {
	t.Parallel()

	c := createCpu()

	testCases := []struct {
		mode cpu.AddressMode
	}{
		{mode: cpu.AddressModeZeroPage},
		{mode: cpu.AddressModeZeroPageY},
		{mode: cpu.AddressModeAbsolute},
		{mode: cpu.AddressModeAbsoluteY},
		{mode: cpu.AddressModeIndirectX},
		{mode: cpu.AddressModeIndirectY},
	}

	for _, test := range testCases {
		c.Registers.P.Write(0)

		writeByteToAddress(c, test.mode, 1)
		c.Registers.A = 2
		c.Registers.X = 2

		cpu.Lax(c, test.mode)

		assert.Equalf(t, uint8(1), c.Registers.A, "Address Mode %s", test.mode.String())
		assert.Equalf(t, uint8(1), c.Registers.X, "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())

		// Zero
		writeByteToAddress(c, test.mode, 0)
		c.Registers.A = 2
		c.Registers.X = 2

		cpu.Lax(c, test.mode)

		assert.Equalf(t, uint8(0), c.Registers.A, "Address Mode %s", test.mode.String())
		assert.Equalf(t, uint8(0), c.Registers.X, "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())

		// Negative
		writeByteToAddress(c, test.mode, 0b10000000)
		c.Registers.A = 2
		c.Registers.X = 2

		cpu.Lax(c, test.mode)

		assert.Equalf(t, uint8(0b10000000), c.Registers.A, "Address Mode %s", test.mode.String())
		assert.Equalf(t, uint8(0b10000000), c.Registers.X, "Address Mode %s", test.mode.String())
		assert.Equalf(t, true, c.Registers.P.ReadFlag(cpu.FlagNegative), "Address Mode %s", test.mode.String())
		assert.Equalf(t, false, c.Registers.P.ReadFlag(cpu.FlagZero), "Address Mode %s", test.mode.String())
	}
}

func TestAax(t *testing.T) {
	t.Parallel()

	c := createCpu()

	testCases := []struct {
		mode cpu.AddressMode
	}{
		{mode: cpu.AddressModeZeroPage},
		{mode: cpu.AddressModeZeroPageY},
		{mode: cpu.AddressModeAbsolute},
		{mode: cpu.AddressModeIndirectX},
	}

	for _, test := range testCases {
		c.Registers.P.Write(0)

		writeByteToAddress(c, test.mode, 0xFF)
		c.Registers.A = 0b00001010
		c.Registers.X = 0b00001001

		cpu.Aax(c, test.mode)

		assert.Equalf(t, uint8(0b00001000), readByteFromAddress(c, test.mode), "Address Mode %s", test.mode.String())
	}

	c.Registers.A = 0x3E
	c.Registers.X = 0x17
	c.Registers.PC = 0
	c.Memory.WriteByteAt(0x01, 0x49)
	c.Memory.WriteByteAt(0x60, 0x89)
	c.Memory.WriteByteAt(0x61, 0x04)

	cpu.Aax(c, cpu.AddressModeIndirectX)

	assert.Equal(t, uint8(0x3E&0x17), c.Memory.ReadByteAt(0x0489))
}
