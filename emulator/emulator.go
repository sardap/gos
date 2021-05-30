package emulator

import (
	"io"

	"github.com/sardap/gos/bus"
	"github.com/sardap/gos/cpu"
	"github.com/sardap/gos/memory"
	"github.com/sardap/gos/ppu"
)

type Emulator struct {
	Memory *memory.Memory
	Ppu    *ppu.Ppu
	Cpu    *cpu.Cpu
	bus    *bus.Bus
}

func Create() *Emulator {
	result := &Emulator{}
	result.bus = &bus.Bus{}
	result.Memory = memory.Create(result.bus)
	result.Ppu = ppu.Create(result.bus)
	result.Cpu = cpu.CreateCpu(result.Memory, result.bus)

	return result
}

func (e *Emulator) LoadRom(r io.Reader) error {
	return e.Memory.LoadRom(r)
}

func (e *Emulator) Step() {
	e.Cpu.Cycles = 0
	e.Cpu.Excute()

	e.Ppu.Step(e.Cpu.Cycles)
}
