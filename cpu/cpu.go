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
