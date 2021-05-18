package cpu

import (
	"fmt"
	"math"

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
	case AddressModeImplied:
		return "Implied"
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
	AddressModeRelative
	AddressModeImplied
	AddressModeLength
)

type Instruction func(c *Cpu, mode AddressMode)

type Operation struct {
	Name        string
	Inst        Instruction
	Length      byte
	MinCycles   byte
	AddressMode AddressMode
}

var (
	opcodes map[byte]*Operation
)

func init() {
	opcodes = map[byte]*Operation{
		// Adc A + M + C -> A, C
		0x69: {Inst: Adc, Length: 2, MinCycles: 2, AddressMode: AddressModeImmediate, Name: "ADC #oper"},
		0x65: {Inst: Adc, Length: 2, MinCycles: 3, AddressMode: AddressModeZeroPage, Name: "ADC oper"},
		0x75: {Inst: Adc, Length: 2, MinCycles: 4, AddressMode: AddressModeZeroPageX, Name: "ADC oper,X"},
		0x6D: {Inst: Adc, Length: 3, MinCycles: 4, AddressMode: AddressModeAbsolute, Name: "ADC oper"},
		0x7D: {Inst: Adc, Length: 3, MinCycles: 4, AddressMode: AddressModeAbsoluteX, Name: "ADC oper,X"},
		0x79: {Inst: Adc, Length: 3, MinCycles: 4, AddressMode: AddressModeAbsoluteY, Name: "ADC oper,Y"},
		0x61: {Inst: Adc, Length: 2, MinCycles: 6, AddressMode: AddressModeIndirectX, Name: "ADC (oper,X)"},
		0x71: {Inst: Adc, Length: 2, MinCycles: 5, AddressMode: AddressModeIndirectY, Name: "ADC (oper),Y"},
		// And A AND M -> A
		0x29: {Inst: And, Length: 2, MinCycles: 2, AddressMode: AddressModeImmediate, Name: "AND #oper"},
		0x25: {Inst: And, Length: 2, MinCycles: 3, AddressMode: AddressModeZeroPage, Name: "AND oper"},
		0x35: {Inst: And, Length: 2, MinCycles: 4, AddressMode: AddressModeZeroPageX, Name: "AND oper,X"},
		0x2D: {Inst: And, Length: 3, MinCycles: 4, AddressMode: AddressModeAbsolute, Name: "AND oper"},
		0x3D: {Inst: And, Length: 3, MinCycles: 4, AddressMode: AddressModeAbsoluteX, Name: "AND oper,X"},
		0x39: {Inst: And, Length: 3, MinCycles: 4, AddressMode: AddressModeAbsoluteY, Name: "AND oper,Y"},
		0x21: {Inst: And, Length: 2, MinCycles: 6, AddressMode: AddressModeIndirectX, Name: "AND (oper,X)"},
		0x31: {Inst: And, Length: 2, MinCycles: 5, AddressMode: AddressModeIndirectY, Name: "AND (oper),Y"},
		// Asl C <- [76543210] <- 0
		0x0A: {Inst: Asl, Length: 1, MinCycles: 2, AddressMode: AddressModeAccumulator, Name: "ASL A"},
		0x06: {Inst: Asl, Length: 2, MinCycles: 5, AddressMode: AddressModeZeroPage, Name: "ASL oper"},
		0x16: {Inst: Asl, Length: 2, MinCycles: 6, AddressMode: AddressModeZeroPageX, Name: "ASL oper,X"},
		0x0E: {Inst: Asl, Length: 3, MinCycles: 6, AddressMode: AddressModeAbsolute, Name: "ASL oper"},
		0x1E: {Inst: Asl, Length: 3, MinCycles: 7, AddressMode: AddressModeAbsoluteX, Name: "ASL oper,X"},
		/*
			Function handles PC
			Extra cycles
		*/
		// Bcc branch on C = 0
		0x90: {Inst: Bcc, Length: 0, MinCycles: 2, AddressMode: AddressModeRelative, Name: "BCC oper"},
		// Branch on Carry Set
		0xB0: {Inst: Bcs, Length: 0, MinCycles: 2, AddressMode: AddressModeRelative, Name: "BCS oper"},
		// Branch on Result Zero
		0xF0: {Inst: Beq, Length: 0, MinCycles: 2, AddressMode: AddressModeRelative, Name: "BEQ oper"},
		// Branch on Result Minus
		0x30: {Inst: Bmi, Length: 0, MinCycles: 2, AddressMode: AddressModeRelative, Name: "BMI oper"},
		// Branch on Result not Zero
		0xD0: {Inst: Bne, Length: 0, MinCycles: 2, AddressMode: AddressModeRelative, Name: "BNE oper"},
		// Branch on Result Plus
		0x10: {Inst: Bpl, Length: 0, MinCycles: 2, AddressMode: AddressModeRelative, Name: "BPL oper"},
		// Branch on Overflow Clear
		0x50: {Inst: Bvc, Length: 0, MinCycles: 2, AddressMode: AddressModeRelative, Name: "BVC oper"},
		// Branch on Overflow Set
		0x70: {Inst: Bvs, Length: 0, MinCycles: 2, AddressMode: AddressModeRelative, Name: "BVS oper"},
		// A AND M, M7 -> N, M6 -> V
		0x24: {Inst: Bit, Length: 2, MinCycles: 3, AddressMode: AddressModeZeroPage, Name: "BIT oper"},
		0x2C: {Inst: Bit, Length: 3, MinCycles: 4, AddressMode: AddressModeAbsolute, Name: "BIT oper"},
		// Clears
		// 0 -> C
		0x18: {Inst: Clc, Length: 1, MinCycles: 2, AddressMode: AddressModeImplied, Name: "CLC"},
		// 0 -> D
		0xD8: {Inst: Cld, Length: 1, MinCycles: 2, AddressMode: AddressModeImplied, Name: "CLD"},
		// 0 -> I
		0x58: {Inst: Cli, Length: 1, MinCycles: 2, AddressMode: AddressModeImplied, Name: "CLI"},
		// 0 -> V
		0xB8: {Inst: Clv, Length: 1, MinCycles: 2, AddressMode: AddressModeImplied, Name: "CLV"},
		// A - M
		0xC9: {Inst: Cmp, Length: 2, MinCycles: 2, AddressMode: AddressModeImmediate, Name: "CMP #oper"},
		0xC5: {Inst: Cmp, Length: 2, MinCycles: 3, AddressMode: AddressModeZeroPage, Name: "CMP oper"},
		0xD5: {Inst: Cmp, Length: 2, MinCycles: 4, AddressMode: AddressModeZeroPageX, Name: "CMP oper,X"},
		0xCD: {Inst: Cmp, Length: 3, MinCycles: 4, AddressMode: AddressModeAbsolute, Name: "CMP oper"},
		0xDD: {Inst: Cmp, Length: 3, MinCycles: 4, AddressMode: AddressModeAbsoluteX, Name: "CMP oper,X"},
		0xD9: {Inst: Cmp, Length: 3, MinCycles: 4, AddressMode: AddressModeAbsoluteY, Name: "CMP oper,Y"},
		0xC1: {Inst: Cmp, Length: 2, MinCycles: 6, AddressMode: AddressModeIndirectX, Name: "CMP (oper,X)"},
		0xD1: {Inst: Cmp, Length: 2, MinCycles: 5, AddressMode: AddressModeIndirectY, Name: "CMP (oper),Y"},
		// X - M
		0xE0: {Inst: Cpx, Length: 2, MinCycles: 2, AddressMode: AddressModeImmediate, Name: "CPX #oper"},
		0xE4: {Inst: Cpx, Length: 2, MinCycles: 3, AddressMode: AddressModeZeroPage, Name: "CPX oper"},
		0xEC: {Inst: Cpx, Length: 3, MinCycles: 4, AddressMode: AddressModeAbsolute, Name: "CPX oper"},
		// Y - M
		0xC0: {Inst: Cpy, Length: 2, MinCycles: 2, AddressMode: AddressModeImmediate, Name: "CPY #oper"},
		0xC4: {Inst: Cpy, Length: 2, MinCycles: 3, AddressMode: AddressModeZeroPage, Name: "CPY oper"},
		0xCC: {Inst: Cpy, Length: 3, MinCycles: 4, AddressMode: AddressModeAbsolute, Name: "CPY oper"},
		// M - 1 -> M
		0xC6: {Inst: Dec, Length: 2, MinCycles: 5, AddressMode: AddressModeZeroPage, Name: "DEC oper"},
		0xD6: {Inst: Dec, Length: 2, MinCycles: 6, AddressMode: AddressModeZeroPageX, Name: "DEC oper,X"},
		0xCE: {Inst: Dec, Length: 2, MinCycles: 6, AddressMode: AddressModeAbsolute, Name: "DEC oper"},
		0xDE: {Inst: Dec, Length: 2, MinCycles: 7, AddressMode: AddressModeAbsoluteX, Name: "DEC oper,X"},
		// X - 1 -> X
		0xCA: {Inst: Dex, Length: 1, MinCycles: 2, AddressMode: AddressModeImplied, Name: "DEX"},
		// Y - 1 -> Y
		0x88: {Inst: Dey, Length: 1, MinCycles: 2, AddressMode: AddressModeImplied, Name: "DEY"},
		// A EOR M -> A
		0x49: {Inst: Eor, Length: 2, MinCycles: 2, AddressMode: AddressModeImmediate, Name: "EOR #oper"},
		0x45: {Inst: Eor, Length: 2, MinCycles: 3, AddressMode: AddressModeZeroPage, Name: "EOR oper"},
		0x55: {Inst: Eor, Length: 2, MinCycles: 4, AddressMode: AddressModeZeroPageX, Name: "EOR oper,X"},
		0x4D: {Inst: Eor, Length: 3, MinCycles: 4, AddressMode: AddressModeAbsolute, Name: "EOR oper"},
		0x5D: {Inst: Eor, Length: 3, MinCycles: 4, AddressMode: AddressModeAbsoluteX, Name: "EOR oper,X"}, //Extra cycles
		0x59: {Inst: Eor, Length: 3, MinCycles: 4, AddressMode: AddressModeAbsoluteY, Name: "EOR oper,Y"}, //Extra cycles
		0x41: {Inst: Eor, Length: 2, MinCycles: 6, AddressMode: AddressModeIndirectX, Name: "EOR (oper,X)"},
		0x51: {Inst: Eor, Length: 2, MinCycles: 5, AddressMode: AddressModeIndirectY, Name: "EOR (oper),Y"}, //Extra cycles
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

func samePage(a, b uint16) bool {
	return 256/math.Max(1, float64(a)) != 256/math.Max(1, float64(b))
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
		address := c.Memory.ReadUint16(c.Registers.PC+1) + uint16(c.Registers.X)
		if samePage(address, c.Registers.PC) {
			c.ExtraTicks++
		}
		return address

	case AddressModeAbsoluteY:
		address := c.Memory.ReadUint16(c.Registers.PC+1) + uint16(c.Registers.Y)
		if samePage(address, c.Registers.PC) {
			c.ExtraTicks++
		}
		return address

	case AddressModeIndirectX:
		return c.Memory.ReadUint16(uint16(operand + c.Registers.X))

	case AddressModeIndirectY:
		address := uint16(c.Memory.ReadUint16(uint16(operand))) + uint16(c.Registers.Y)
		if samePage(address, c.Registers.PC) {
			c.ExtraTicks++
		}
		return address

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

// CMP and CPX are speical
func cmpCarryHappend(a, right uint8) bool {
	return a >= right
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

func branchOnFlag(c *Cpu, flag bool) {
	orginalAddress := c.Registers.PC
	if flag {
		c.Registers.PC += uint16(int8(c.Memory.ReadByte(c.Registers.PC + 1)))

		oldPage := orginalAddress / 256
		newPage := c.Registers.PC / 256
		if oldPage == newPage {
			c.ExtraTicks++
		} else {
			c.ExtraTicks += 2
		}
	} else {
		c.Registers.PC += 2
	}
}

func Bcc(c *Cpu, mode AddressMode) {
	branchOnFlag(c, !c.Registers.P.ReadFlag(FlagCarry))
}

func Bcs(c *Cpu, mode AddressMode) {
	branchOnFlag(c, c.Registers.P.ReadFlag(FlagCarry))
}

func Beq(c *Cpu, mode AddressMode) {
	branchOnFlag(c, c.Registers.P.ReadFlag(FlagZero))
}

func Bmi(c *Cpu, mode AddressMode) {
	branchOnFlag(c, c.Registers.P.ReadFlag(FlagNegative))
}

func Bne(c *Cpu, mode AddressMode) {
	branchOnFlag(c, !c.Registers.P.ReadFlag(FlagZero))
}

func Bpl(c *Cpu, mode AddressMode) {
	branchOnFlag(c, !c.Registers.P.ReadFlag(FlagNegative))
}

func Bvc(c *Cpu, mode AddressMode) {
	branchOnFlag(c, !c.Registers.P.ReadFlag(FlagOverflow))
}

func Bvs(c *Cpu, mode AddressMode) {
	branchOnFlag(c, c.Registers.P.ReadFlag(FlagOverflow))
}

func Bit(c *Cpu, mode AddressMode) {
	operand := c.readByte(mode)

	c.Registers.P.SetFlag(FlagNegative, nesmath.BitSet(operand, 7))
	c.Registers.P.SetFlag(FlagZero, zeroHappend(uint16(c.Registers.A&operand)))
	c.Registers.P.SetFlag(FlagOverflow, nesmath.BitSet(operand, 6))

	c.writeByte(mode, c.Registers.A&operand)
}

func clearBit(c *Cpu, flag Flag) {
	c.Registers.P.SetFlag(flag, false)
}

func Clc(c *Cpu, mode AddressMode) {
	clearBit(c, FlagCarry)
}

func Cld(c *Cpu, mode AddressMode) {
	clearBit(c, FlagDecimal)
}

func Cli(c *Cpu, mode AddressMode) {
	clearBit(c, FlagInteruprtDisable)
}

func Clv(c *Cpu, mode AddressMode) {
	clearBit(c, FlagOverflow)
}

func compare(c *Cpu, mode AddressMode, reg uint8) {
	operand := c.readByte(mode)

	result := uint16(reg) - uint16(operand)

	c.Registers.P.SetFlag(FlagNegative, negativeHappend(result))
	c.Registers.P.SetFlag(FlagZero, zeroHappend(result))
	c.Registers.P.SetFlag(FlagCarry, cmpCarryHappend(reg, operand))
}

func Cmp(c *Cpu, mode AddressMode) {
	compare(c, mode, c.Registers.A)
}

func Cpx(c *Cpu, mode AddressMode) {
	compare(c, mode, c.Registers.X)
}

func Cpy(c *Cpu, mode AddressMode) {
	compare(c, mode, c.Registers.Y)
}

func decerment(c *Cpu, value uint8) uint8 {
	result := value - 1

	c.Registers.P.SetFlag(FlagNegative, negativeHappend(uint16(result)))
	c.Registers.P.SetFlag(FlagZero, zeroHappend(uint16(result)))

	return result
}

func Dec(c *Cpu, mode AddressMode) {
	operand := c.readByte(mode)
	result := decerment(c, operand)
	c.writeByte(mode, result)
}

func Dex(c *Cpu, mode AddressMode) {
	c.Registers.X = decerment(c, c.Registers.X)
}

func Dey(c *Cpu, mode AddressMode) {
	c.Registers.Y = decerment(c, c.Registers.Y)
}

func Eor(c *Cpu, mode AddressMode) {
	operand := c.readByte(mode)

	result := c.Registers.A ^ operand

	c.Registers.P.SetFlag(FlagNegative, negativeHappend(uint16(result)))
	c.Registers.P.SetFlag(FlagZero, zeroHappend(uint16(result)))

	c.Registers.A ^= operand
}
