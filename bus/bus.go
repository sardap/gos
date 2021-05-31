package bus

import (
	"bytes"
	"io"

	"github.com/sardap/gos/cart"
)

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
	Ppu  Ppu
	Mem  Memory
	Cpu  Cpu
	Cart cart.Cart
}

func (b *Bus) AddPpu(ppu Ppu) {
	b.Ppu = ppu
}

func (b *Bus) AddMemory(memory Memory) {
	b.Mem = memory
}

type bytesQueue struct {
	r io.Reader
}

func (b *bytesQueue) Pop() (byte, error) {
	buf := make([]byte, 1)
	_, err := b.r.Read(buf)
	return buf[0], err
}

func (b *bytesQueue) PopN(n int64) ([]byte, error) {
	buf := make([]byte, n)
	_, err := b.r.Read(buf)
	return buf, err
}

func (b *Bus) LoadRom(r io.Reader) error {
	buffer := bytesQueue{
		r: r,
	}

	info := cart.CartInfo{}

	dotNesHeaderPrefix := []byte{0x4E, 0x45, 0x53, 0x1A}
	if cartPrefix, err := buffer.PopN(4); err != nil || !bytes.Equal(dotNesHeaderPrefix, cartPrefix) {
		return cart.ErrInvalidRom
	}

	var err error
	info.PrgRomBanks, _ = buffer.Pop()
	info.ChrRomBanks, _ = buffer.Pop()
	controlByte1, _ := buffer.Pop()
	info.ControlByte1 = cart.CreateControlByte1(controlByte1)
	controlByte2, _ := buffer.Pop()
	info.ControlByte2, err = cart.CreateControlByte2(controlByte2)
	if err != nil {
		return err
	}
	info.PrgRamLength, _ = buffer.Pop()

	b.Cart = cart.CreateCart(info)

	buffer.PopN(7)

	for i := 0; i < int(info.PrgRomBanks); i++ {
		data, err := buffer.PopN(16384)
		if err != nil {
			return cart.ErrInvalidRom
		}
		b.Cart.WriteBytesPrg(data)
	}

	for i := 0; i < int(info.ChrRomBanks); i++ {
		data, err := buffer.PopN(8192)
		if err != nil {
			return cart.ErrInvalidRom
		}
		b.Cart.WriteBytesChr(data)
	}

	return nil
}
