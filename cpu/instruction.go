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
	case AddressModeAccumulator:
		return "Accumulator"
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
	AddressModeAccumulator
	AddressModeLength
)

type instruction func(c *Cpu, mode AddressMode)

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
		//And A AND M -> A
		0x29: {Inst: And, Length: 2, MinCycles: 2, AddressMode: AddressModeImmediate, Name: "AND #oper"},
		0x25: {Inst: And, Length: 2, MinCycles: 3, AddressMode: AddressModeZeroPage, Name: "AND oper"},
		0x35: {Inst: And, Length: 2, MinCycles: 4, AddressMode: AddressModeZeroPageX, Name: "AND oper,X"},
		0x2D: {Inst: And, Length: 3, MinCycles: 4, AddressMode: AddressModeAbsolute, Name: "AND oper"},
		0x3D: {Inst: And, Length: 3, MinCycles: 4, AddressMode: AddressModeAbsoluteX, Name: "AND oper,X"},
		0x39: {Inst: And, Length: 3, MinCycles: 4, AddressMode: AddressModeAbsoluteY, Name: "AND oper,Y"},
		0x21: {Inst: And, Length: 2, MinCycles: 6, AddressMode: AddressModeIndirectX, Name: "AND (oper,X)"},
		0x31: {Inst: And, Length: 2, MinCycles: 5, AddressMode: AddressModeIndirectY, Name: "AND (oper),Y"},
		//Asl C <- [76543210] <- 0
		0x0A: {Inst: Asl, Length: 1, MinCycles: 2, AddressMode: AddressModeAccumulator, Name: "ASL A"},
		0x06: {Inst: Asl, Length: 2, MinCycles: 5, AddressMode: AddressModeZeroPage, Name: "ASL oper"},
		0x16: {Inst: Asl, Length: 2, MinCycles: 6, AddressMode: AddressModeZeroPageX, Name: "ASL oper,X"},
		0x0E: {Inst: Asl, Length: 3, MinCycles: 6, AddressMode: AddressModeAbsolute, Name: "ASL oper"},
		0x1E: {Inst: Asl, Length: 3, MinCycles: 7, AddressMode: AddressModeAbsoluteX, Name: "ASL oper,X"},
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

	operation.Inst(c, operation.AddressMode)

	fmt.Printf("%02X %02X %02X\n", opcode, oprand1, oprand2)
}

func (c *Cpu) GetOprandAddress(addressMode AddressMode) uint16 {
	operand := c.Memory.ReadByte(c.Registers.PC + 1)

	switch addressMode {
	case AddressModeImmediate:
		return c.Registers.PC + 1

	case AddressModeZeroPage:
		return uint16(operand)

	case AddressModeZeroPageX:
		return uint16(c.Memory.ReadByte(c.Registers.PC+1)) + uint16(c.Registers.X)&0x00FF

	case AddressModeAbsolute:
		return c.Memory.ReadUint16(c.Registers.PC + 1)

	case AddressModeAbsoluteX:
		return c.Memory.ReadUint16(c.Registers.PC+1) + uint16(c.Registers.X)

	case AddressModeAbsoluteY:
		return c.Memory.ReadUint16(c.Registers.PC+1) + uint16(c.Registers.Y)

	case AddressModeIndirectX:
		return c.Memory.ReadUint16(uint16(operand + c.Registers.X))

	case AddressModeIndirectY:
		return uint16(c.Memory.ReadUint16(uint16(operand))) + uint16(c.Registers.Y)

	default:
		panic(fmt.Errorf("address mode not implmented"))
	}
}

func (c *Cpu) readByte(mode AddressMode) byte {
	switch mode {
	case AddressModeAccumulator:
		return c.Registers.A
	default:
		address := c.GetOprandAddress(mode)
		return c.Memory.ReadByte(address)
	}
}

func (c *Cpu) writeByte(mode AddressMode, value byte) {
	switch mode {
	case AddressModeAccumulator:
		c.Registers.A = value
	default:
		c.Memory.WriteByte(c.GetOprandAddress(mode), value)
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

func Adc(c *Cpu, mode AddressMode) {
	oprand := c.readByte(mode)

	a := c.Registers.A
	carry := uint16(c.Registers.P.ReadFlagByte(FlagCarry))
	result := uint16(a) + uint16(oprand) + carry

	c.Registers.P.SetFlag(FlagNegative, negativeHappend(result))
	c.Registers.P.SetFlag(FlagZero, zeroHappend(result))
	c.Registers.P.SetFlag(FlagCarry, carryHappend(result))
	c.Registers.P.SetFlag(FlagOverflow, overflowHappend(a, oprand, byte(result)))

	c.Registers.A = byte(result)
}

func And(c *Cpu, mode AddressMode) {
	oprand := c.readByte(mode)

	a := c.Registers.A
	result := uint16(a & oprand)

	c.Registers.P.SetFlag(FlagNegative, negativeHappend(result))
	c.Registers.P.SetFlag(FlagZero, zeroHappend(result))

	c.Registers.A = byte(result)
}

func Asl(c *Cpu, mode AddressMode) {
	oprand := c.readByte(mode)

	result := uint16(oprand << 1)

	c.Registers.P.SetFlag(FlagNegative, negativeHappend(result))
	c.Registers.P.SetFlag(FlagZero, zeroHappend(result))
	c.Registers.P.SetFlag(FlagCarry, carryHappend(result))

	c.writeByte(mode, byte(result))
}
