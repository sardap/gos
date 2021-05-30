package bus

type Element interface {
	WriteByteAt(address uint16, value byte)
	ReadByteAt(address uint16) byte
}

type Ppu interface {
	Element
	WriteByteToOAM(offset byte, val byte)
	GetStatus() byte
	WriteMask(val byte)
	WriteController(val byte)
}

type Memory interface {
	Element
}

type InterruptType byte

const (
	InterruptTypeBreak InterruptType = iota
	InterruptTypeNim
)

type Cpu interface {
	Interrupt(interrupt InterruptType)
}

type Bus struct {
	Ppu Ppu
	Mem Memory
	Cpu Cpu
}

func (b *Bus) AddPpu(ppu Ppu) {
	b.Ppu = ppu
}

func (b *Bus) AddMemory(memory Memory) {
	b.Mem = memory
}
