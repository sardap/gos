package memory

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

const (
	StackOffset = 0x0100
)

var (
	ErrInvalidRom     = fmt.Errorf("invalid rom header")
	ErrInvalidAddress = fmt.Errorf("invalid address")
)

type Memory struct {
	iRam         [0x0800]byte
	cart         Cart
	PpuRegisters *PpuRegisters
}

func Create() *Memory {
	return &Memory{
		PpuRegisters: CreatePpuRegisters(),
	}
}

func (m *Memory) SetCart(cart Cart) {
	m.cart = cart
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
		panic(fmt.Errorf("APU and IO regsiters not created"))
	//Funky APU and IO
	case address >= 0x4018 && address <= 0x401F:
		panic(fmt.Errorf("funky APU and IO not created"))
	//Cart space: PRG, ROM, PRG, RAM and mappers
	case address >= 0x4020 && address <= 0xFFFF:
		m.cart.WriteByteAt(address, value)
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
		panic(fmt.Errorf("APU and IO regsiters not created"))
	//Funky APU and IO
	case address >= 0x4018 && address <= 0x401F:
		panic(fmt.Errorf("funky APU and IO not created"))
	//Cart space: PRG, ROM, PRG, RAM and mappers
	case address >= 0x4020 && address <= 0xFFFF:
		return m.cart.ReadByteAt(address)
	}

	panic(fmt.Errorf("invalid address"))
}

func (m *Memory) ReadUint16At(address uint16) uint16 {
	return binary.LittleEndian.Uint16(
		[]byte{m.ReadByteAt(address), m.ReadByteAt(address + 1)})
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

func (m *Memory) LoadRom(r io.Reader) error {
	buffer := bytesQueue{
		r: r,
	}

	info := CartInfo{}

	dotNesHeaderPrefix := []byte{0x4E, 0x45, 0x53, 0x1A}
	if cartPrefix, err := buffer.PopN(4); err != nil || !bytes.Equal(dotNesHeaderPrefix, cartPrefix) {
		return ErrInvalidRom
	}

	var err error
	info.PrgRomBanks, _ = buffer.Pop()
	info.ChrRomBanks, _ = buffer.Pop()
	controlByte1, _ := buffer.Pop()
	info.ControlByte1 = createControlByte1(controlByte1)
	controlByte2, _ := buffer.Pop()
	info.ControlByte2, err = createControlByte2(controlByte2)
	if err != nil {
		return err
	}
	info.PrgRamLength, _ = buffer.Pop()

	m.cart = createCart(info)

	buffer.PopN(7)

	for i := 0; i < int(info.PrgRomBanks); i++ {
		data, err := buffer.PopN(16384)
		if err != nil {
			return ErrInvalidRom
		}
		m.cart.WriteBytesPrg(data)
	}

	for i := 0; i < int(info.ChrRomBanks); i++ {
		data, err := buffer.PopN(8192)
		if err != nil {
			return ErrInvalidRom
		}
		m.cart.WriteBytesPrg(data)
	}

	return nil
}
