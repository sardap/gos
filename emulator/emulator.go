package emulator

import (
	"io"
	"time"

	"github.com/sardap/gos/cpu"
	"github.com/sardap/gos/memory"
	"github.com/sardap/gos/ppu"
)

type Emulator struct {
	memory *memory.Memory
	ppu    *ppu.Ppu
	cpu    *cpu.Cpu
}

func Create() *Emulator {
	result := &Emulator{}
	result.memory = memory.Create()
	result.ppu = ppu.Create()
	result.cpu = cpu.CreateCpu(result.memory, result.ppu)

	return result
}

func (e *Emulator) LoadRom(r io.Reader) error {
	return e.memory.LoadRom(r)
}

func (e *Emulator) Step() {
	e.cpu.Excute()

	time.Sleep(time.Duration(10) * time.Millisecond)
}
