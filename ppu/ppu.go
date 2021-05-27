package ppu

import (
	"fmt"
)

var (
	ErrInvalidAddress = fmt.Errorf("invalid ppu address")
)

type Ppu struct {
	PatternTable0 [0x1000]byte // 0x0000 - 0x0FFF
	PatternTable1 [0x1000]byte // 0x1000 - 0x1FFF
	NameTable0    [0x0400]byte // 0x2000 - 0x23FF
	NameTable1    [0x0400]byte // 0x2400 - 0x27FF
	NameTable2    [0x0400]byte // 0x2800 - 0x2BFF
	NameTable3    [0x0400]byte // 0x2C00 - 0x2FFF
	PalRam        [0x0020]byte // 0x3F00 - 0x3F1F
}

func Create() *Ppu {
	return &Ppu{}
}

func (p *Ppu) Step(cycles int) {
	for i := 0; i < cycles*3; i++ {

	}
}

func (p *Ppu) WriteByteAt(address uint16, value byte) {
	switch address & 0xF000 {
	case 0x0000:
		p.PatternTable0[address] = value
	case 0x1000:
		p.PatternTable1[address-0x1000] = value
	case 0x2000:
		switch address & 0x0F00 {
		case 0x0000, 0x0100, 0x0200, 0x0300:
			p.NameTable0[address-0x2000] = value
		case 0x0400, 0x0500, 0x0600, 0x0700:
			p.NameTable1[address-0x2400] = value
		case 0x0800, 0x0900, 0x0A00, 0x0B00:
			p.NameTable2[address-0x2800] = value
		case 0x0C00, 0x0D00, 0x0E00, 0x0F00:
			p.NameTable3[address-0x2C00] = value
		}
	case 0x3000:
		switch {
		// Mirror of 0x2000 - 0x2EFF
		case address >= 0x3000 && address <= 0x3EFF:
			p.WriteByteAt(address-0x1000, value)
		case address >= 0x3F00 && address <= 0x3F1F:
			p.PalRam[address-0x3F00] = value
		// Mirror of 0x3F00 - 0x3F1F
		case address >= 0x3F20 && address <= 0x3FFF:
			p.WriteByteAt(address-0x0020, value)
		}
	default:
		p.WriteByteAt(address-0x3FFF, value)
	}
}

func (p *Ppu) ReadByteAt(address uint16) byte {
	switch address & 0xF000 {
	case 0x0000:
		return p.PatternTable0[address]
	case 0x1000:
		return p.PatternTable1[address-0x1000]
	case 0x2000:
		switch address & 0x0F00 {
		case 0x0000, 0x0100, 0x0200, 0x0300:
			return p.NameTable0[address-0x2000]
		case 0x0400, 0x0500, 0x0600, 0x0700:
			return p.NameTable1[address-0x2400]
		case 0x0800, 0x0900, 0x0A00, 0x0B00:
			return p.NameTable2[address-0x2800]
		case 0x0C00, 0x0D00, 0x0E00, 0x0F00:
			return p.NameTable3[address-0x2C00]
		}
	case 0x3000:
		switch {
		// Mirror of 0x2000 - 0x2EFF
		case address >= 0x3000 && address <= 0x3EFF:
			return p.ReadByteAt(address - 0x1000)
		case address >= 0x3F00 && address <= 0x3F1F:
			return p.PalRam[address-0x3F00]
		// Mirror of 0x3F00 - 0x3F1F
		case address >= 0x3F20 && address <= 0x3FFF:
			return p.ReadByteAt(address - 0x0020)
		}
	}

	return p.ReadByteAt(address - 0x3FFF)
}
