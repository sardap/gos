package cpu

import (
	"fmt"
	"log"
	"strings"

	nesmath "github.com/sardap/gos/math"
	"github.com/sardap/gos/memory"
	"github.com/sardap/gos/ppu"
)

type Cpu struct {
	Registers   *Registers
	Memory      *memory.Memory
	Ppu         *ppu.Ppu
	Cycles      int
	ExtraCycles byte
	// http://nesdev.com/the%20%27B%27%20flag%20&%20BRK%20instruction.txt
	Interupt bool
}

func CreateCpu(mem *memory.Memory, ppu *ppu.Ppu) *Cpu {
	return &Cpu{
		Registers: CreateRegisters(),
		Memory:    mem,
		Ppu:       ppu,
		Cycles:    0,
		Interupt:  true,
	}
}

func (c *Cpu) PushByte(value byte) {
	c.Memory.WriteByteAt(memory.StackOffset+uint16(c.Registers.SP), value)
	c.Registers.SP--
}

func (c *Cpu) PushP() {
	value := c.Registers.P.Read()
	value = nesmath.SetBit(value, byte(FlagBreakCommand), c.Interupt)
	value = nesmath.SetBit(value, byte(FlagUnsued), true)

	c.PushByte(value)
}

func (c *Cpu) PopByte() byte {
	c.Registers.SP++
	result := c.Memory.ReadByteAt(memory.StackOffset + uint16(c.Registers.SP))
	return result
}

func (c *Cpu) PopP() {
	value := c.PopByte()
	value = nesmath.SetBit(value, byte(FlagUnsued), true)
	value = nesmath.SetBit(value, byte(FlagBreakCommand), false)

	c.Registers.P.Write(value)
}

func (c *Cpu) PushUint16(value uint16) {
	c.Memory.WriteUint16At(memory.StackOffset+uint16(c.Registers.SP), value)
	c.Registers.SP -= 2
}

func (c *Cpu) PopUint16() uint16 {
	c.Registers.SP += 2
	result := c.Memory.ReadUint16At(memory.StackOffset + uint16(c.Registers.SP))
	return result
}

func (c *Cpu) logStep(operation Operation) {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("%04X  ", c.Registers.PC))
	builder.WriteString(fmt.Sprintf("%02X ", c.Memory.ReadByteAt(c.Registers.PC)))
	if operation.Length >= 2 || operation.AddressMode == AddressModeRelative {
		builder.WriteString(fmt.Sprintf("%02X ", c.Memory.ReadByteAt(c.Registers.PC+1)))
	}
	if operation.Length >= 3 {
		builder.WriteString(fmt.Sprintf("%02X ", c.Memory.ReadByteAt(c.Registers.PC+2)))
	}
	for builder.Len() < 16 {
		builder.WriteString(" ")
	}

	builder.WriteString(strings.Split(operation.Name, " ")[0])
	builder.WriteString(" ")
	if operation.Length >= 2 || operation.Length == 0 {
		if operation.AddressMode == AddressModeImmediate {
			builder.WriteString("#$")
		} else {
			builder.WriteString("$")
		}
		if operation.Length >= 3 || operation.Length == 0 {
			builder.WriteString(fmt.Sprintf("%02X", c.Memory.ReadByteAt(c.Registers.PC+2)))
		}
		builder.WriteString(fmt.Sprintf("%02X", c.Memory.ReadByteAt(c.Registers.PC+1)))
	}

	for builder.Len() < 48 {
		builder.WriteString(" ")
	}

	builder.WriteString(
		fmt.Sprintf(
			"A:%02X X:%02X Y:%02X P:%02X SP:%02X ",
			c.Registers.A, c.Registers.X, c.Registers.Y,
			c.Registers.P.Read(), c.Registers.SP,
		),
	)

	builder.WriteString("\n")

	log.Print(builder.String())
}

func (c *Cpu) Excute() {
	opcode := c.Memory.ReadByteAt(c.Registers.PC)

	operation, ok := opcodes[opcode]
	if !ok {
		panic(fmt.Errorf("unkown opcode %02X", opcode))
	}

	c.logStep(*operation)

	operation.Inst(c, operation.AddressMode)

	c.Registers.PC += operation.Length
	c.Cycles += operation.MinCycles + int(c.ExtraCycles)
	c.ExtraCycles = 0
}

func (c *Cpu) GetOprandAddress(addressMode AddressMode) uint16 {
	byteOperand := c.Memory.ReadByteAt(c.Registers.PC + 1)

	switch addressMode {
	case AddressModeImmediate:
		return c.Registers.PC + 1

	case AddressModeZeroPage:
		return uint16(byteOperand)

	case AddressModeZeroPageX:
		return uint16(c.Memory.ReadByteAt(c.Registers.PC+1)) + uint16(c.Registers.X)&0x00FF

	case AddressModeZeroPageY:
		return uint16(c.Memory.ReadByteAt(c.Registers.PC+1)) + uint16(c.Registers.Y)&0x00FF

	case AddressModeAbsolute:
		return c.Memory.ReadUint16At(c.Registers.PC + 1)

	case AddressModeAbsoluteX:
		address := c.Memory.ReadUint16At(c.Registers.PC+1) + uint16(c.Registers.X)
		if samePage(address, c.Registers.PC) {
			c.ExtraCycles++
		}
		return address

	case AddressModeAbsoluteY:
		address := c.Memory.ReadUint16At(c.Registers.PC+1) + uint16(c.Registers.Y)
		if samePage(address, c.Registers.PC) {
			c.ExtraCycles++
		}
		return address

	case AddressModeIndirect:
		operand := c.Memory.ReadUint16At(c.Registers.PC + 1)
		return c.Memory.ReadUint16At(operand)

	case AddressModeIndirectX:
		return c.Memory.ReadUint16At(uint16(byteOperand + c.Registers.X))

	case AddressModeIndirectY:
		address := uint16(c.Memory.ReadUint16At(uint16(byteOperand))) + uint16(c.Registers.Y)
		if samePage(address, c.Registers.PC) {
			c.ExtraCycles++
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
		return c.Memory.ReadByteAt(address)
	}
}

func (c *Cpu) writeByte(mode AddressMode, value byte) {
	switch mode {
	case AddressModeAccumulator:
		c.Registers.A = value
	default:
		c.Memory.WriteByteAt(c.GetOprandAddress(mode), value)
	}
}
