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
	Bus    *bus.Bus

	PpuEnabled bool
}

func Create() *Emulator {
	result := &Emulator{}
	result.Bus = &bus.Bus{}
	result.Memory = memory.Create(result.Bus)
	result.Ppu = ppu.Create(result.Bus)
	result.Cpu = cpu.CreateCpu(result.Memory, result.Bus)
	result.PpuEnabled = true

	return result
}

func (e *Emulator) LoadRom(r io.Reader) error {
	return e.Bus.LoadRom(r)
}

func (e *Emulator) Step() {
	e.Cpu.Cycles = 0
	e.Cpu.Excute()

	if e.PpuEnabled {
		e.Ppu.Step(e.Cpu.Cycles)
	}
}
