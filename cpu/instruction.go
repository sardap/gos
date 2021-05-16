package cpu

import (
	"fmt"

	nesmath "github.com/sardap/gos/math"
)

type AddressMode int

func (a *AddressMode) String() string {
	switch *a {
	case AddressModeImmediate:
		return "Immediate"
	case AddressModeZeroPage:
		return "ZeroPage"
	case AddressModeZeroPageX:
		return "ZeroPageX"
	case AddressModeZeroPageY:
		return "ZeroPageY"
	case AddressModeAbsolute:
		return "Absolute"
	case AddressModeAbsoluteX:
		return "AbsoluteX"
	case AddressModeAbsoluteY:
		return "AbsoluteY"
	case AddressModeIndirectX:
		return "IndirectX"
	case AddressModeIndirectY:
		return "IndirectY"
	}

	panic(fmt.Errorf("unkown addressMode string"))
}

const (
	AddressModeImmediate AddressMode = iota
	AddressModeZeroPage
	AddressModeZeroPageX
	AddressModeZeroPageY
	AddressModeAbsolute
	AddressModeAbsoluteX
	AddressModeAbsoluteY
	AddressModeIndirectX
	AddressModeIndirectY
	AddressModeLength
)

type instruction func(c *Cpu, OprandAddress uint16)

type Operation struct {
	Name        string
	Inst        instruction
	Length      byte
	MinCycles   byte
	AddressMode AddressMode
}

var (
	opcodes map[byte]*Operation
)

func init() {
	opcodes = map[byte]*Operation{
		//Adc A + M + C -> A, C
		0x69: {Inst: Adc, Length: 2, MinCycles: 2, AddressMode: AddressModeImmediate, Name: "ADC #oper"},
		0x65: {Inst: Adc, Length: 2, MinCycles: 3, AddressMode: AddressModeZeroPage, Name: "ADC oper"},
		0x75: {Inst: Adc, Length: 2, MinCycles: 4, AddressMode: AddressModeZeroPageX, Name: "ADC oper,X"},
		0x6D: {Inst: Adc, Length: 3, MinCycles: 4, AddressMode: AddressModeAbsolute, Name: "ADC oper"},
		0x7D: {Inst: Adc, Length: 3, MinCycles: 4, AddressMode: AddressModeAbsoluteX, Name: "ADC oper,X"},
		0x79: {Inst: Adc, Length: 3, MinCycles: 4, AddressMode: AddressModeAbsoluteY, Name: "ADC oper,Y"},
		0x61: {Inst: Adc, Length: 2, MinCycles: 6, AddressMode: AddressModeIndirectX, Name: "ADC (oper,X)"},
		0x71: {Inst: Adc, Length: 2, MinCycles: 5, AddressMode: AddressModeIndirectY, Name: "ADC (oper),Y"},
	}
}

func GetOpcodes() map[byte]*Operation {
	return opcodes
}

func (c *Cpu) Excute() {
	opcode := c.Memory.ReadByte(c.Registers.PC)
	oprand1 := c.Memory.ReadByte(c.Registers.PC + 1)
	oprand2 := c.Memory.ReadByte(c.Registers.PC + 2)

	operation := opcodes[opcode]

	oprandAddress := c.GetOprandAddress(operation.AddressMode)

	operation.Inst(c, oprandAddress)

	fmt.Printf("%02X %02X %02X\n", opcode, oprand1, oprand2)
}

func (c *Cpu) GetOprandAddress(addressMode AddressMode) uint16 {
	oprand := c.Memory.ReadByte(c.Registers.PC + 1)

	switch addressMode {
	case AddressModeImmediate:
		return c.Registers.PC + 1

	case AddressModeZeroPage:
		return uint16(c.Memory.ReadByte(c.Registers.PC + 1))

	case AddressModeZeroPageX:
		return uint16(c.Memory.ReadByte(c.Registers.PC+1)) + uint16(c.Registers.X)&0x00FF

	case AddressModeAbsolute:
		return c.Memory.ReadUint16(c.Registers.PC + 1)

	case AddressModeAbsoluteX:
		return c.Memory.ReadUint16(c.Registers.PC+1) + uint16(c.Registers.X)

	case AddressModeAbsoluteY:
		return c.Memory.ReadUint16(c.Registers.PC+1) + uint16(c.Registers.Y)

	case AddressModeIndirectX:
		return c.Memory.ReadUint16((uint16(oprand) + uint16(c.Registers.X)) & 0xFF)

	case AddressModeIndirectY:
		return c.Memory.ReadUint16((uint16(c.Memory.ReadByte(uint16(oprand))) + uint16(c.Registers.Y)))

	default:
		panic(fmt.Errorf("address mode not implmented"))
	}
}

func overflowHappend(left, right, result byte) bool {
	if int8(left) > 0 && int8(right) > 0 {
		return nesmath.BitSet(result, 7)
	} else if int8(left) < 0 && int8(right) < 0 {
		return !nesmath.BitSet(result, 7)
	}

	return false
}

func carryHappend(result uint16) bool {
	return result&0xFF00 > 0
}

func negativeHappend(result uint16) bool {
	return nesmath.BitSet(byte(result), 7)
}

func zeroHappend(result uint16) bool {
	return byte(result) == 0
}

func Adc(c *Cpu, OprandAddress uint16) {
	oprand := c.Memory.ReadByte(OprandAddress)

	a := c.Registers.A
	carry := uint16(c.Registers.P.ReadFlagByte(FlagCarry))
	result := uint16(a) + uint16(oprand) + carry

	c.Registers.P.SetFlag(FlagNegative, negativeHappend(result))
	c.Registers.P.SetFlag(FlagZero, zeroHappend(result))
	c.Registers.P.SetFlag(FlagCarry, carryHappend(result))
	c.Registers.P.SetFlag(FlagOverflow, overflowHappend(a, oprand, byte(result)))

	c.Registers.A = byte(result)
}
