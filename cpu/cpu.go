package cpu

import "github.com/sardap/gos/memory"

type Cpu struct {
	Registers  *Registers
	Memory     *memory.Memory
	Ticks      int
	ExtraTicks byte
}

func CreateCpu(mem *memory.Memory) *Cpu {
	return &Cpu{
		Registers: CreateRegisters(),
		Memory:    mem,
		Ticks:     0,
	}
}

func (c *Cpu) PushByte(value byte) {
	c.Memory.WriteByte(memory.StackOffset+uint16(c.Registers.SP), value)
	c.Registers.SP--
}

func (c *Cpu) PopByte() byte {
	result := c.Memory.ReadByte(memory.StackOffset + uint16(c.Registers.SP))
	c.Registers.SP++
	return result
}

func (c *Cpu) PushUint16(value uint16) {
	c.Memory.WriteShort(memory.StackOffset+uint16(c.Registers.SP), value)
	c.Registers.SP -= 2
}

func (c *Cpu) PopUint16() uint16 {
	c.Registers.SP += 2
	result := c.Memory.ReadUint16(memory.StackOffset + uint16(c.Registers.SP))
	return result
}
