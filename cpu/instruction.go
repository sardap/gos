package cpu

import (
	"fmt"
	"math"

	nesmath "github.com/sardap/gos/math"
)

type AddressMode int

func (a AddressMode) String() string {
	switch a {
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
	case AddressModeIndirect:
		return "Indirect"
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
	AddressModeIndirect
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
	Length      uint16
	MinCycles   int
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
		0x90: {Inst: Bcc, Length: 2, MinCycles: 2, AddressMode: AddressModeRelative, Name: "BCC oper"},
		// Branch on Carry Set
		0xB0: {Inst: Bcs, Length: 2, MinCycles: 2, AddressMode: AddressModeRelative, Name: "BCS oper"},
		// Branch on Result Zero
		0xF0: {Inst: Beq, Length: 2, MinCycles: 2, AddressMode: AddressModeRelative, Name: "BEQ oper"},
		// Branch on Result Minus
		0x30: {Inst: Bmi, Length: 2, MinCycles: 2, AddressMode: AddressModeRelative, Name: "BMI oper"},
		// Branch on Result not Zero
		0xD0: {Inst: Bne, Length: 2, MinCycles: 2, AddressMode: AddressModeRelative, Name: "BNE oper"},
		// Branch on Result Plus
		0x10: {Inst: Bpl, Length: 2, MinCycles: 2, AddressMode: AddressModeRelative, Name: "BPL oper"},
		// interrupt; push PC+2; push SR
		0x00: {Inst: Brk, Length: 1, MinCycles: 7, AddressMode: AddressModeImplied, Name: "BRK"},
		// Branch on Overflow Clear
		0x50: {Inst: Bvc, Length: 2, MinCycles: 2, AddressMode: AddressModeRelative, Name: "BVC oper"},
		// Branch on Overflow Set
		0x70: {Inst: Bvs, Length: 2, MinCycles: 2, AddressMode: AddressModeRelative, Name: "BVS oper"},
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
		0xCE: {Inst: Dec, Length: 3, MinCycles: 6, AddressMode: AddressModeAbsolute, Name: "DEC oper"},
		0xDE: {Inst: Dec, Length: 3, MinCycles: 7, AddressMode: AddressModeAbsoluteX, Name: "DEC oper,X"},
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
		// M + 1 -> M
		0xE6: {Inst: Inc, Length: 2, MinCycles: 5, AddressMode: AddressModeZeroPage, Name: "INC oper"},
		0xF6: {Inst: Inc, Length: 2, MinCycles: 6, AddressMode: AddressModeZeroPageX, Name: "INC oper,X"},
		0xEE: {Inst: Inc, Length: 3, MinCycles: 6, AddressMode: AddressModeAbsolute, Name: "INC oper"},
		0xFE: {Inst: Inc, Length: 3, MinCycles: 7, AddressMode: AddressModeAbsoluteX, Name: "INC oper,X"},
		// X + 1 -> X
		0xE8: {Inst: Inx, Length: 1, MinCycles: 2, AddressMode: AddressModeImplied, Name: "INX"},
		// Y + 1 -> Y
		0xC8: {Inst: Iny, Length: 1, MinCycles: 2, AddressMode: AddressModeImplied, Name: "INY"},
		// (PC+1) -> PCL; (PC+2) -> PCH
		0x4C: {Inst: Jmp, Length: 0, MinCycles: 3, AddressMode: AddressModeAbsolute, Name: "JMP oper"},
		0x6C: {Inst: Jmp, Length: 0, MinCycles: 5, AddressMode: AddressModeIndirect, Name: "JMP (oper)"},
		// push (PC+2); (PC+1) -> PCL; (PC+2) -> PCH
		0x20: {Inst: Jsr, Length: 0, MinCycles: 6, AddressMode: AddressModeAbsolute, Name: "JSR oper"},
		// M -> A
		0xA9: {Inst: Lda, Length: 2, MinCycles: 2, AddressMode: AddressModeImmediate, Name: "LDA #oper"},
		0xA5: {Inst: Lda, Length: 2, MinCycles: 3, AddressMode: AddressModeZeroPage, Name: "LDA oper"},
		0xB5: {Inst: Lda, Length: 2, MinCycles: 4, AddressMode: AddressModeZeroPageX, Name: "LDA oper,X"},
		0xAD: {Inst: Lda, Length: 3, MinCycles: 4, AddressMode: AddressModeAbsolute, Name: "LDA oper"},
		0xBD: {Inst: Lda, Length: 3, MinCycles: 4, AddressMode: AddressModeAbsoluteX, Name: "LDA oper,X"}, // Extra Cycles
		0xB9: {Inst: Lda, Length: 3, MinCycles: 4, AddressMode: AddressModeAbsoluteY, Name: "LDA oper,Y"}, // Extra Cycles
		0xA1: {Inst: Lda, Length: 2, MinCycles: 6, AddressMode: AddressModeIndirectX, Name: "LDA (oper,X)"},
		0xB1: {Inst: Lda, Length: 2, MinCycles: 5, AddressMode: AddressModeIndirectY, Name: "LDA (oper),Y"}, // Extra Cycles
		// M -> X
		0xA2: {Inst: Ldx, Length: 2, MinCycles: 2, AddressMode: AddressModeImmediate, Name: "LDX #oper"},
		0xA6: {Inst: Ldx, Length: 2, MinCycles: 3, AddressMode: AddressModeZeroPage, Name: "LDX oper"},
		0xB6: {Inst: Ldx, Length: 2, MinCycles: 4, AddressMode: AddressModeZeroPageY, Name: "LDX oper,Y"},
		0xAE: {Inst: Ldx, Length: 3, MinCycles: 4, AddressMode: AddressModeAbsolute, Name: "LDX oper"},
		0xBE: {Inst: Ldx, Length: 3, MinCycles: 4, AddressMode: AddressModeAbsoluteY, Name: "LDX oper,Y"}, // Extra Cycles
		// M -> Y
		0xA0: {Inst: Ldy, Length: 2, MinCycles: 2, AddressMode: AddressModeImmediate, Name: "LDY #oper"},
		0xA4: {Inst: Ldy, Length: 2, MinCycles: 3, AddressMode: AddressModeZeroPage, Name: "LDY oper"},
		0xB4: {Inst: Ldy, Length: 2, MinCycles: 4, AddressMode: AddressModeZeroPageX, Name: "LDY oper,X"},
		0xAC: {Inst: Ldy, Length: 3, MinCycles: 4, AddressMode: AddressModeAbsolute, Name: "LDY oper"},
		0xBC: {Inst: Ldy, Length: 3, MinCycles: 4, AddressMode: AddressModeAbsoluteX, Name: "LDY oper,X"}, // Extra Cycles
		// 0 -> [76543210] -> C
		0x4A: {Inst: Lsr, Length: 1, MinCycles: 2, AddressMode: AddressModeAccumulator, Name: "LSR A"},
		0x46: {Inst: Lsr, Length: 2, MinCycles: 5, AddressMode: AddressModeZeroPage, Name: "LSR oper"},
		0x56: {Inst: Lsr, Length: 2, MinCycles: 6, AddressMode: AddressModeZeroPageX, Name: "LSR oper,X"},
		0x4E: {Inst: Lsr, Length: 3, MinCycles: 6, AddressMode: AddressModeAbsolute, Name: "LSR oper"},
		0x5E: {Inst: Lsr, Length: 3, MinCycles: 7, AddressMode: AddressModeAbsoluteX, Name: "LSR oper,X"},
		// No Operation
		0xEA: {Inst: Nop, Length: 1, MinCycles: 2, AddressMode: AddressModeImplied, Name: "NOP"}, // Extra Cycles
		// A OR M -> A
		0x09: {Inst: Ora, Length: 2, MinCycles: 2, AddressMode: AddressModeImmediate, Name: "ORA #oper"},
		0x05: {Inst: Ora, Length: 2, MinCycles: 3, AddressMode: AddressModeZeroPage, Name: "ORA oper"},
		0x15: {Inst: Ora, Length: 2, MinCycles: 4, AddressMode: AddressModeZeroPageX, Name: "ORA oper,X"},
		0x0D: {Inst: Ora, Length: 3, MinCycles: 4, AddressMode: AddressModeAbsolute, Name: "ORA oper"},
		0x1D: {Inst: Ora, Length: 3, MinCycles: 4, AddressMode: AddressModeAbsoluteX, Name: "ORA oper,X"}, //Extra cycles
		0x19: {Inst: Ora, Length: 3, MinCycles: 4, AddressMode: AddressModeAbsoluteY, Name: "ORA oper,Y"}, //Extra cycles
		0x01: {Inst: Ora, Length: 2, MinCycles: 6, AddressMode: AddressModeIndirectX, Name: "ORA (oper,X)"},
		0x11: {Inst: Ora, Length: 2, MinCycles: 5, AddressMode: AddressModeIndirectY, Name: "ORA (oper),Y"}, //Extra cycles
		// push A
		0x48: {Inst: Pha, Length: 1, MinCycles: 3, AddressMode: AddressModeImplied, Name: "PHA"},
		// push SR
		0x08: {Inst: Php, Length: 1, MinCycles: 3, AddressMode: AddressModeImplied, Name: "PHP"},
		// pull A
		0x68: {Inst: Pla, Length: 1, MinCycles: 4, AddressMode: AddressModeImplied, Name: "PLA"},
		// pull SR
		0x28: {Inst: Plp, Length: 1, MinCycles: 4, AddressMode: AddressModeImplied, Name: "PLP"},
		// c <- [76543210] <- c
		0x2A: {Inst: Rol, Length: 1, MinCycles: 2, AddressMode: AddressModeAccumulator, Name: "ROL A"},
		0x26: {Inst: Rol, Length: 2, MinCycles: 5, AddressMode: AddressModeZeroPage, Name: "ROL oper"},
		0x36: {Inst: Rol, Length: 2, MinCycles: 6, AddressMode: AddressModeZeroPageX, Name: "ROL oper,X"},
		0x2E: {Inst: Rol, Length: 3, MinCycles: 6, AddressMode: AddressModeAbsolute, Name: "ROL oper"},
		0x3E: {Inst: Rol, Length: 3, MinCycles: 7, AddressMode: AddressModeAbsoluteX, Name: "ROL oper,X"},
		// c -> [76543210] -> c
		0x6A: {Inst: Ror, Length: 1, MinCycles: 2, AddressMode: AddressModeAccumulator, Name: "ROR A"},
		0x66: {Inst: Ror, Length: 2, MinCycles: 5, AddressMode: AddressModeZeroPage, Name: "ROR oper"},
		0x76: {Inst: Ror, Length: 2, MinCycles: 6, AddressMode: AddressModeZeroPageX, Name: "ROR oper,X"},
		0x6E: {Inst: Ror, Length: 3, MinCycles: 6, AddressMode: AddressModeAbsolute, Name: "ROR oper"},
		0x7E: {Inst: Ror, Length: 3, MinCycles: 7, AddressMode: AddressModeAbsoluteX, Name: "ROR oper,X"},
		// pull SR; pull PC
		0x40: {Inst: Rti, Length: 0, MinCycles: 6, AddressMode: AddressModeImplied, Name: "RTI"},
		// pull PC, PC+1 -> PC
		0x60: {Inst: Rts, Length: 0, MinCycles: 6, AddressMode: AddressModeImplied, Name: "RTS"},
		// A - M - C -> A
		0xE9: {Inst: Sbc, Length: 2, MinCycles: 2, AddressMode: AddressModeImmediate, Name: "SBC #oper"},
		0xE5: {Inst: Sbc, Length: 2, MinCycles: 3, AddressMode: AddressModeZeroPage, Name: "SBC oper"},
		0xF5: {Inst: Sbc, Length: 2, MinCycles: 4, AddressMode: AddressModeZeroPageX, Name: "SBC oper,X"},
		0xED: {Inst: Sbc, Length: 3, MinCycles: 4, AddressMode: AddressModeAbsolute, Name: "SBC oper"},
		0xFD: {Inst: Sbc, Length: 3, MinCycles: 4, AddressMode: AddressModeAbsoluteX, Name: "SBC oper,X"}, //Extra cycles
		0xF9: {Inst: Sbc, Length: 3, MinCycles: 4, AddressMode: AddressModeAbsoluteY, Name: "SBC oper,Y"}, //Extra cycles
		0xE1: {Inst: Sbc, Length: 2, MinCycles: 6, AddressMode: AddressModeIndirectX, Name: "SBC (oper,X)"},
		0xF1: {Inst: Sbc, Length: 2, MinCycles: 5, AddressMode: AddressModeIndirectY, Name: "SBC (oper),Y"}, //Extra cycles
		// 1 -> C
		0x38: {Inst: Sec, Length: 1, MinCycles: 2, AddressMode: AddressModeImplied, Name: "SEC"},
		// 1 -> D
		0xF8: {Inst: Sed, Length: 1, MinCycles: 2, AddressMode: AddressModeImplied, Name: "SED"},
		// 1 -> I
		0x78: {Inst: Sei, Length: 1, MinCycles: 2, AddressMode: AddressModeImplied, Name: "SEI"},
		// A -> M
		0x85: {Inst: Sta, Length: 2, MinCycles: 3, AddressMode: AddressModeZeroPage, Name: "STA oper"},
		0x95: {Inst: Sta, Length: 2, MinCycles: 4, AddressMode: AddressModeZeroPageX, Name: "STA oper,X"},
		0x8D: {Inst: Sta, Length: 3, MinCycles: 4, AddressMode: AddressModeAbsolute, Name: "STA oper"},
		0x9D: {Inst: Sta, Length: 3, MinCycles: 5, AddressMode: AddressModeAbsoluteX, Name: "STA oper,X"},
		0x99: {Inst: Sta, Length: 3, MinCycles: 5, AddressMode: AddressModeAbsoluteY, Name: "STA oper,Y"},
		0x81: {Inst: Sta, Length: 2, MinCycles: 6, AddressMode: AddressModeIndirectX, Name: "STA (oper,X)"},
		0x91: {Inst: Sta, Length: 2, MinCycles: 6, AddressMode: AddressModeIndirectY, Name: "STA (oper),Y"},
		// X -> M
		0x86: {Inst: Stx, Length: 2, MinCycles: 3, AddressMode: AddressModeZeroPage, Name: "STX oper"},
		0x96: {Inst: Stx, Length: 2, MinCycles: 4, AddressMode: AddressModeZeroPageY, Name: "STA oper,Y"},
		0x8E: {Inst: Stx, Length: 3, MinCycles: 4, AddressMode: AddressModeAbsolute, Name: "STA oper"},
		// Y -> M
		0x84: {Inst: Sty, Length: 2, MinCycles: 3, AddressMode: AddressModeZeroPage, Name: "STX oper"},
		0x94: {Inst: Sty, Length: 2, MinCycles: 4, AddressMode: AddressModeZeroPageX, Name: "STA oper,X"},
		0x8C: {Inst: Sty, Length: 3, MinCycles: 4, AddressMode: AddressModeAbsolute, Name: "STA oper"},
		// A -> X
		0xAA: {Inst: Tax, Length: 1, MinCycles: 2, AddressMode: AddressModeImplied, Name: "TAX"},
		// A -> Y
		0xA8: {Inst: Tay, Length: 1, MinCycles: 2, AddressMode: AddressModeImplied, Name: "TAY"},
		// SP -> X
		0xBA: {Inst: Tsx, Length: 1, MinCycles: 2, AddressMode: AddressModeImplied, Name: "TSX"},
		// X -> A
		0x8A: {Inst: Txa, Length: 1, MinCycles: 2, AddressMode: AddressModeImplied, Name: "TXA"},
		// X -> SP
		0x9A: {Inst: Txs, Length: 1, MinCycles: 2, AddressMode: AddressModeImplied, Name: "TXS"},
		// Y -> A
		0x98: {Inst: Tya, Length: 1, MinCycles: 2, AddressMode: AddressModeImplied, Name: "TYA"},
	}
}

func GetOpcodes() map[byte]*Operation {
	return opcodes
}

func samePage(a, b uint16) bool {
	return 256/math.Max(1, float64(a)) != 256/math.Max(1, float64(b))
}

func overflowHappend(left, right, result byte) bool {
	if int8(left) > 0 && int8(right) > 0 {
		return nesmath.BitSet(result, 7)
	} else if int8(left) < 0 && int8(right) < 0 {
		return !nesmath.BitSet(result, 7)
	}

	return false
}

func postiveCarryHappend(result uint16) bool {
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
	c.Registers.P.SetFlag(FlagCarry, postiveCarryHappend(result))
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
	operand := c.readByte(mode)

	result := uint16(operand) << 1

	c.Registers.P.SetFlag(FlagNegative, negativeHappend(result))
	c.Registers.P.SetFlag(FlagZero, zeroHappend(result))
	c.Registers.P.SetFlag(FlagCarry, postiveCarryHappend(result))

	c.writeByte(mode, byte(result))
}

func branchOnFlag(c *Cpu, flag bool) {
	orginalAddress := c.Registers.PC
	if flag {
		c.Registers.PC += uint16(int8(c.Memory.ReadByteAt(c.Registers.PC + 1)))

		oldPage := orginalAddress / 256
		newPage := c.Registers.PC / 256
		if oldPage == newPage {
			c.ExtraCycles++
		} else {
			c.ExtraCycles += 2
		}
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

func Brk(c *Cpu, mode AddressMode) {
	c.PushUint16(c.Registers.PC + 2)
	c.PushP()
	c.Interupt = true
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

func incerment(c *Cpu, value uint8) uint8 {
	result := value + 1

	c.Registers.P.SetFlag(FlagNegative, negativeHappend(uint16(result)))
	c.Registers.P.SetFlag(FlagZero, zeroHappend(uint16(result)))

	return result
}

func Inc(c *Cpu, mode AddressMode) {
	operand := c.readByte(mode)
	result := incerment(c, operand)
	c.writeByte(mode, result)
}

func Inx(c *Cpu, mode AddressMode) {
	c.Registers.X = incerment(c, c.Registers.X)
}

func Iny(c *Cpu, mode AddressMode) {
	c.Registers.Y = incerment(c, c.Registers.Y)
}

func Jmp(c *Cpu, mode AddressMode) {
	c.Registers.PC = c.GetOprandAddress(mode)

}

// Jsr trick https://wiki.nesdev.com/w/index.php/RTS_Trick
func Jsr(c *Cpu, mode AddressMode) {
	c.PushUint16(c.Registers.PC + 2)
	c.Registers.PC = c.GetOprandAddress(mode)
}

func Lda(c *Cpu, mode AddressMode) {
	operand := c.readByte(mode)
	c.Registers.A = operand

	c.Registers.P.SetFlag(FlagNegative, negativeHappend(uint16(operand)))
	c.Registers.P.SetFlag(FlagZero, zeroHappend(uint16(operand)))
}

func Ldx(c *Cpu, mode AddressMode) {
	operand := c.readByte(mode)
	c.Registers.X = operand

	c.Registers.P.SetFlag(FlagNegative, negativeHappend(uint16(operand)))
	c.Registers.P.SetFlag(FlagZero, zeroHappend(uint16(operand)))
}

func Ldy(c *Cpu, mode AddressMode) {
	operand := c.readByte(mode)
	c.Registers.Y = operand

	c.Registers.P.SetFlag(FlagNegative, negativeHappend(uint16(operand)))
	c.Registers.P.SetFlag(FlagZero, zeroHappend(uint16(operand)))
}

func Lsr(c *Cpu, mode AddressMode) {
	oprand := c.readByte(mode)

	result := uint16(oprand >> 1)

	c.Registers.P.SetFlag(FlagNegative, false)
	c.Registers.P.SetFlag(FlagZero, zeroHappend(result))
	// this isn't a mistake
	c.Registers.P.SetFlag(FlagCarry, result == 0)

	c.writeByte(mode, byte(result))
}

func Nop(c *Cpu, mode AddressMode) {
}

func Ora(c *Cpu, mode AddressMode) {
	operand := c.readByte(mode)

	result := c.Registers.A | operand

	c.Registers.P.SetFlag(FlagNegative, negativeHappend(uint16(result)))
	c.Registers.P.SetFlag(FlagZero, zeroHappend(uint16(result)))

	c.Registers.A = result
}

func Pha(c *Cpu, mode AddressMode) {
	c.PushByte(c.Registers.A)
}

func Php(c *Cpu, mode AddressMode) {
	c.PushP()
}

func Pla(c *Cpu, mode AddressMode) {
	a := c.PopByte()

	c.Registers.P.SetFlag(FlagNegative, negativeHappend(uint16(a)))
	c.Registers.P.SetFlag(FlagZero, zeroHappend(uint16(a)))

	c.Registers.A = a
}

func Plp(c *Cpu, mode AddressMode) {
	c.PopP()
}

func Rol(c *Cpu, mode AddressMode) {
	operand := c.readByte(mode)

	result := (uint16(operand) << 1) | uint16(c.Registers.P.ReadFlagByte(FlagCarry))

	c.Registers.P.SetFlag(FlagNegative, negativeHappend(result))
	c.Registers.P.SetFlag(FlagZero, zeroHappend(result))
	c.Registers.P.SetFlag(FlagCarry, postiveCarryHappend(result))

	c.writeByte(mode, byte(result))
}

func Ror(c *Cpu, mode AddressMode) {
	operand := c.readByte(mode)

	result := (uint16(operand) >> 1) | uint16(c.Registers.P.ReadFlagByte(FlagCarry)<<7)

	c.Registers.P.SetFlag(FlagNegative, negativeHappend(result))
	c.Registers.P.SetFlag(FlagZero, zeroHappend(result))
	c.Registers.P.SetFlag(FlagCarry, result&0b00000001 > 0)

	c.writeByte(mode, byte(result))
}

func Rti(c *Cpu, mode AddressMode) {
	c.PopP()
	c.Registers.PC = c.PopUint16()
}

func Rts(c *Cpu, mode AddressMode) {
	c.Registers.PC = c.PopUint16() + 1
}

func Sbc(c *Cpu, mode AddressMode) {
	// Fucking stole this what the fuck is this shit
	// Adding to minus too much big brains me think
	oprand := c.readByte(mode) ^ 0xFF

	a := c.Registers.A
	carry := uint16(c.Registers.P.ReadFlagByte(FlagCarry))
	result := uint16(a) + uint16(oprand) + carry

	c.Registers.P.SetFlag(FlagNegative, negativeHappend(result))
	c.Registers.P.SetFlag(FlagZero, zeroHappend(result))
	c.Registers.P.SetFlag(FlagCarry, postiveCarryHappend(result))
	c.Registers.P.SetFlag(FlagOverflow, overflowHappend(a, oprand, byte(result)))

	c.Registers.A = byte(result)
}

func Sec(c *Cpu, mode AddressMode) {
	c.Registers.P.SetFlag(FlagCarry, true)
}

func Sed(c *Cpu, mode AddressMode) {
	c.Registers.P.SetFlag(FlagDecimal, true)
}

func Sei(c *Cpu, mode AddressMode) {
	c.Registers.P.SetFlag(FlagInteruprtDisable, true)
}

func Sta(c *Cpu, mode AddressMode) {
	c.writeByte(mode, c.Registers.A)
}

func Stx(c *Cpu, mode AddressMode) {
	c.writeByte(mode, c.Registers.X)
}

func Sty(c *Cpu, mode AddressMode) {
	c.writeByte(mode, c.Registers.Y)
}

func Tax(c *Cpu, mode AddressMode) {
	c.Registers.X = c.Registers.A

	c.Registers.P.SetFlag(FlagNegative, negativeHappend(uint16(c.Registers.X)))
	c.Registers.P.SetFlag(FlagZero, zeroHappend(uint16(c.Registers.X)))
}

func Tay(c *Cpu, mode AddressMode) {
	c.Registers.Y = c.Registers.A

	c.Registers.P.SetFlag(FlagNegative, negativeHappend(uint16(c.Registers.Y)))
	c.Registers.P.SetFlag(FlagZero, zeroHappend(uint16(c.Registers.Y)))
}

func Tsx(c *Cpu, mode AddressMode) {
	c.Registers.X = c.Registers.SP

	c.Registers.P.SetFlag(FlagNegative, negativeHappend(uint16(c.Registers.X)))
	c.Registers.P.SetFlag(FlagZero, zeroHappend(uint16(c.Registers.X)))
}

func Txa(c *Cpu, mode AddressMode) {
	c.Registers.A = c.Registers.X

	c.Registers.P.SetFlag(FlagNegative, negativeHappend(uint16(c.Registers.A)))
	c.Registers.P.SetFlag(FlagZero, zeroHappend(uint16(c.Registers.A)))
}

func Txs(c *Cpu, mode AddressMode) {
	c.Registers.SP = c.Registers.X
}

func Tya(c *Cpu, mode AddressMode) {
	c.Registers.A = c.Registers.Y

	c.Registers.P.SetFlag(FlagNegative, negativeHappend(uint16(c.Registers.Y)))
	c.Registers.P.SetFlag(FlagZero, zeroHappend(uint16(c.Registers.Y)))
}
