package memory

import (
	"encoding/binary"
	"fmt"
)

const (
	StackOffset = 0x0100
)

type Memory struct {
	iRam [0x07D0]byte
}

func Create() *Memory {
	return &Memory{}
}

func (m *Memory) WriteByte(address uint16, value byte) {
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
		panic(fmt.Errorf("PPU regsiters not created"))
	//Mirror of PPU repeats every 8 bytes
	case address >= 0x2008 && address <= 0x3FFF:
		panic(fmt.Errorf("PPU regsiters not created"))
	//APU and IO
	case address >= 0x4000 && address <= 0x4017:
		panic(fmt.Errorf("APU and IO regsiters not created"))
	//Funky APU and IO
	case address >= 0x4018 && address <= 0x401F:
		panic(fmt.Errorf("funky APU and IO not created"))
	//Cart space: PRG, ROM, PRG, RAM and mappers
	case address >= 0x4020 && address <= 0xFFFF:
		panic(fmt.Errorf("cart space not created"))
	}
}

func (m *Memory) WriteShort(address, value uint16) {
	m.WriteByte(address, byte(value&0x00FF))
	m.WriteByte(address+1, byte(value>>8))
}

func (m *Memory) ReadByte(address uint16) byte {
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
		panic(fmt.Errorf("PPU regsiters not created"))
	//Mirror of PPU repeats every 8 bytes
	case address >= 0x2008 && address <= 0x3FFF:
		panic(fmt.Errorf("PPU regsiters not created"))
	//APU and IO
	case address >= 0x4000 && address <= 0x4017:
		panic(fmt.Errorf("APU and IO regsiters not created"))
	//Funky APU and IO
	case address >= 0x4018 && address <= 0x401F:
		panic(fmt.Errorf("funky APU and IO not created"))
	//Cart space: PRG, ROM, PRG, RAM and mappers
	case address >= 0x4020 && address <= 0xFFFF:
		panic(fmt.Errorf("cart space not created"))
	}

	panic(fmt.Errorf("invalid address"))
}

func (m *Memory) ReadUint16(address uint16) uint16 {
	return binary.LittleEndian.Uint16(
		[]byte{m.ReadByte(address), m.ReadByte(address + 1)})
}
