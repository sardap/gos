package emulator

import (
	"io"
	"time"

	"github.com/sardap/gos/cpu"
	"github.com/sardap/gos/memory"
)

type Emulator struct {
	cpu    *cpu.Cpu
	memory *memory.Memory
}

func Create() *Emulator {
	result := &Emulator{}
	result.memory = &memory.Memory{}
	result.cpu = cpu.CreateCpu(result.memory)

	return result
}

func (e *Emulator) LoadRom(r io.Reader) error {
	return e.memory.LoadRom(r)
}

func (e *Emulator) Step() {
	e.cpu.Excute()

	time.Sleep(time.Duration(10) * time.Millisecond)
}
