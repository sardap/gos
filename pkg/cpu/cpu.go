package cpu

import (
	"encoding/binary"
	"fmt"
	"log"
	"strings"

	"github.com/sardap/gos/pkg/bus"
	"github.com/sardap/gos/pkg/memory"
	"github.com/sardap/gos/pkg/utility"
)

const (
	stackOffset = 0x0100
	irqOffset   = 0xFFFE
	nimOffset   = 0xFFFa
)

type Cpu struct {
	Registers   *Registers
	Memory      *memory.Memory
	Cycles      int
	ExtraCycles byte
	// http://nesdev.com/the%20%27B%27%20flag%20&%20BRK%20instruction.txt
	Interupt bool
	bus      *bus.Bus
}

func CreateCpu(mem *memory.Memory, b *bus.Bus) *Cpu {
	result := &Cpu{
		Registers: CreateRegisters(),
		Memory:    mem,
		bus:       b,
		Cycles:    0,
		Interupt:  true,
	}
	b.Cpu = result
	return result
}

func (c *Cpu) PushByte(value byte) {
	c.Memory.WriteByteAt(stackOffset+uint16(c.Registers.SP), value)
	c.Registers.SP--
}

func (c *Cpu) PushP() {
	value := c.Registers.P.Read()
	value = utility.SetBit(value, byte(FlagBreakCommand), c.Interupt)
	value = utility.SetBit(value, byte(FlagUnsued), true)

	c.PushByte(value)
}

func (c *Cpu) PopByte() byte {
	c.Registers.SP++
	result := c.Memory.ReadByteAt(stackOffset + uint16(c.Registers.SP))
	return result
}

func (c *Cpu) PopP() {
	value := c.PopByte()
	value = utility.SetBit(value, byte(FlagUnsued), true)
	value = utility.SetBit(value, byte(FlagBreakCommand), false)

	c.Registers.P.Write(value)
}

func (c *Cpu) PushUint16(value uint16) {
	c.PushByte(byte(value >> 8))
	c.PushByte(byte(value & 0xff))
}

func (c *Cpu) PopUint16() uint16 {
	return binary.LittleEndian.Uint16([]byte{c.PopByte(), c.PopByte()})
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

	c.Cycles += operation.MinCycles
	if operation.CanHaveExtraCycles {
		c.Cycles += int(c.ExtraCycles)
	}
	if c.Memory.DmaTransfer {
		c.Memory.DmaTransfer = false
		// This might lead to issues down the road
		// https://wiki.nesdev.com/w/index.php/PPU_registers#OAMDMA
		c.Cycles += 513
	}

	c.ExtraCycles = 0
}

func diffrentPages(old, new uint16) bool {
	return old&0xFF00 != new&0xFF00
}

func (c *Cpu) GetOprandAddress(addressMode AddressMode) uint16 {
	byteOperand := c.Memory.ReadByteAt(c.Registers.PC + 1)

	switch addressMode {
	case AddressModeImmediate:
		return c.Registers.PC + 1

	case AddressModeZeroPage:
		return uint16(byteOperand)

	case AddressModeZeroPageX:
		return utility.WrapUint16(uint16(byteOperand)+uint16(c.Registers.X), 0xFF)

	case AddressModeZeroPageY:
		return utility.WrapUint16(uint16(byteOperand)+uint16(c.Registers.Y), 0xFF)

	case AddressModeAbsolute:
		return c.Memory.ReadUint16At(c.Registers.PC + 1)

	case AddressModeAbsoluteX:
		address := c.Memory.ReadUint16At(c.Registers.PC + 1)
		if diffrentPages(address, address+uint16(c.Registers.X)) {
			c.ExtraCycles++
		}
		return address + uint16(c.Registers.X)

	case AddressModeAbsoluteY:
		address := c.Memory.ReadUint16At(c.Registers.PC + 1)
		if diffrentPages(address, address+uint16(c.Registers.Y)) {
			c.ExtraCycles++
		}
		return address + uint16(c.Registers.Y)

	case AddressModeIndirect:
		// Guess who spent 5 hours staring at this fucking thing
		// Only to find out it's a bug with the 6502 https://atariage.com/forums/topic/72382-6502-indirect-addressing-ff-behavior/
		address := uint16(c.Memory.ReadUint16At(c.Registers.PC + 1))
		buffer := make([]byte, 2)
		binary.LittleEndian.PutUint16(buffer, address)
		secondAddress := binary.LittleEndian.Uint16([]byte{
			buffer[0] + 1,
			buffer[1],
		})
		return binary.LittleEndian.Uint16([]byte{
			c.Memory.ReadByteAt(address),
			c.Memory.ReadByteAt(secondAddress),
		})

	case AddressModeIndirectX:
		address := uint16(byteOperand) + uint16(c.Registers.X)
		return binary.LittleEndian.Uint16([]byte{
			c.Memory.ReadByteAt(utility.WrapUint16(address, 0xFF)),
			c.Memory.ReadByteAt(utility.WrapUint16(address+1, 0xFF)),
		})

	case AddressModeIndirectY:
		indirect := binary.LittleEndian.Uint16([]byte{
			c.Memory.ReadByteAt(utility.WrapUint16(uint16(byteOperand), 0xFF)),
			c.Memory.ReadByteAt(utility.WrapUint16(uint16(byteOperand)+1, 0xFF)),
		})
		address := indirect + uint16(c.Registers.Y)
		if diffrentPages(address, indirect) {
			c.ExtraCycles++
		}
		return address

	default:
		panic(fmt.Errorf("address mode not implmented"))
	}
}

func (c *Cpu) ReadByteByMode(mode AddressMode) byte {
	switch mode {
	case AddressModeAccumulator:
		return c.Registers.A
	default:
		address := c.GetOprandAddress(mode)
		return c.Memory.ReadByteAt(address)
	}
}

func (c *Cpu) WriteByteByMode(mode AddressMode, value byte) {
	switch mode {
	case AddressModeAccumulator:
		c.Registers.A = value
	default:
		address := c.GetOprandAddress(mode)
		c.Memory.WriteByteAt(address, value)
	}
}

func (c *Cpu) Interrupt(interrupt bus.InterruptType) {
	c.Interupt = true

	if interrupt == bus.InterruptTypeBreak {
		c.Registers.PC++
	}

	c.PushUint16(c.Registers.PC + 2)
	c.PushP()

	switch interrupt {
	case bus.InterruptTypeBreak:
		c.Registers.PC = c.Memory.ReadUint16At(irqOffset)
	case bus.InterruptTypeNim:
		c.Registers.PC = c.Memory.ReadUint16At(nimOffset)
	}
}
