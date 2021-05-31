package memory

import (
	"encoding/binary"
	"fmt"

	"github.com/sardap/gos/bus"
	nesmath "github.com/sardap/gos/math"
)

var (
	ErrInvalidAddress = fmt.Errorf("invalid address")
)

type Memory struct {
	iRam         [0x0800]byte
	PpuRegisters *PpuRegisters
	Apu          *Apu
	DmaTransfer  bool
	bus          *bus.Bus
}

func Create(b *bus.Bus) *Memory {
	result := &Memory{
		PpuRegisters: CreatePpuRegisters(b),
		Apu:          CreateApu(),
		bus:          b,
	}
	b.AddMemory(result)
	return result
}

func (m *Memory) WriteByteAt(address uint16, value byte) {
	switch {
	//Intenal Ram
	case address >= 0x0000 && address <= 0x07FF:
		m.iRam[address] = value
	//Mirror of Ram
	case address >= 0x0800 && address <= 0x0FFF:
		m.iRam[address-0x0800] = value
	//Mirror of Ram
	case address >= 0x1000 && address <= 0x17FF:
		m.iRam[address-0x1000] = value
	//Mirror of Ram
	case address >= 0x1800 && address <= 0x1FFF:
		m.iRam[address-0x1800] = value
	//PPU
	case address >= 0x2000 && address <= 0x2007:
		m.PpuRegisters.WriteByteAt(address, value)
	//Mirror of PPU repeats every 8 bytes
	case address >= 0x2008 && address <= 0x3FFF:
		m.PpuRegisters.WriteByteAt(0x2000+address%8, value)
	//APU and IO
	case address >= 0x4000 && address <= 0x4017:
		switch address {
		// OMA DMA
		case 0x4014:
			m.DmaTransfer = true
			baseReadAddress := nesmath.CombineToUint16(value, 0x00)
			for i := byte(0x00); i < 0xFF; i++ {
				m.bus.Ppu.WriteByteToOAM(i, m.ReadByteAt(baseReadAddress+uint16(i)))
			}
		default:
			m.Apu.WriteByteAt(address, value)
		}
	//Funky APU and IO
	case address >= 0x4018 && address <= 0x401F:
		panic(fmt.Errorf("funky APU and IO not created"))
	//Cart space: PRG, ROM, PRG, RAM and mappers
	case address >= 0x4020 && address <= 0xFFFF:
		m.bus.Cart.WriteByteAt(address, value)
	}
}

func (m *Memory) WriteUint16At(address, value uint16) {
	m.WriteByteAt(address, byte(value&0x00FF))
	m.WriteByteAt(address+1, byte(value>>8))
}

func (m *Memory) ReadByteAt(address uint16) byte {
	switch {
	//Intenal Ram
	case address >= 0x0000 && address <= 0x07FF:
		return m.iRam[address]
	//Mirror of Ram
	case address >= 0x0800 && address <= 0x0FFF:
		return m.iRam[address-0x0800]
	//Mirror of Ram
	case address >= 0x1000 && address <= 0x17FF:
		return m.iRam[address-0x1000]
	//Mirror of Ram
	case address >= 0x1800 && address <= 0x1FFF:
		return m.iRam[address-0x1800]
	//PPU
	case address >= 0x2000 && address <= 0x2007:
		return m.PpuRegisters.ReadByteAt(address)
	//Mirror of PPU repeats every 8 bytes
	case address >= 0x2008 && address <= 0x3FFF:
		return m.PpuRegisters.ReadByteAt(0x2000 + address%8)
	//APU and IO
	case address >= 0x4000 && address <= 0x4017:
		return m.Apu.ReadByteAt(address)
	//Funky APU and IO
	case address >= 0x4018 && address <= 0x401F:
		panic(fmt.Errorf("funky APU and IO not created"))
	//Cart space: PRG, ROM, PRG, RAM and mappers
	case address >= 0x4020 && address <= 0xFFFF:
		return m.bus.Cart.ReadByteAt(address)
	}

	panic(fmt.Errorf("invalid address"))
}

func (m *Memory) ReadUint16At(address uint16) uint16 {
	return binary.LittleEndian.Uint16(
		[]byte{m.ReadByteAt(address), m.ReadByteAt(address + 1)})
}
