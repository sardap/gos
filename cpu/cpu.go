package cpu

import (
	"github.com/sardap/gos/memory"
	"github.com/sardap/gos/ppu"
)

type Cpu struct {
	Registers   *Registers
	Memory      *memory.Memory
	Ppu         *ppu.Ppu
	Ticks       int
	ExtraCycles byte
	Interupt    bool
}

func CreateCpu(mem *memory.Memory, ppu *ppu.Ppu) *Cpu {
	return &Cpu{
		Registers: CreateRegisters(),
		Memory:    mem,
		Ppu:       ppu,
		Ticks:     0,
	}
}

func (c *Cpu) PushByte(value byte) {
	c.Memory.WriteByteAt(memory.StackOffset+uint16(c.Registers.SP), value)
	c.Registers.SP--
}

func (c *Cpu) PopByte() byte {
	c.Registers.SP++
	result := c.Memory.ReadByteAt(memory.StackOffset + uint16(c.Registers.SP))
	return result
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
