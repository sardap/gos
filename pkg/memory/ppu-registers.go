package memory

import (
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/sardap/gos/pkg/bus"
	"github.com/sardap/gos/pkg/utility"
)

type PpuFlag byte

const (
	// https://wiki.nesdev.com/w/index.php/PPU_registers#PPUSTATUS
	PpuFlagStatusVerticalBlank  PpuFlag = 7
	PpuFlagStatusSpirt0Hit      PpuFlag = 6
	PpuFlagStatusSpirteOverflow PpuFlag = 5
)

type Register struct {
	val byte
}

func (p *Register) Write(value byte) {
	p.val = value
}

func (p *Register) Read() byte {
	return p.val
}

func (p *Register) BitSet(flag PpuFlag) bool {
	return utility.BitSet(p.Read(), byte(flag))
}

type PpuCtrl struct {
	Register
}

func (p *PpuCtrl) VramIncrement() byte {
	if utility.BitSet(p.val, 2) {
		// Down
		return 32
	}

	// Across
	return 1
}

func (p *PpuCtrl) NameTableAddress() byte {
	flag := byte(0)
	flag = utility.SetBit(flag, 0, utility.BitSet(p.val, 0))
	flag = utility.SetBit(flag, 1, utility.BitSet(p.val, 1))

	return flag
}

// The Ppu and CPU are on diffrent bus so I also seprated the objects
type PpuRegisters struct {
	Ctrl *PpuCtrl  // 0x2000
	Mask *Register // 0x2001
	// 0x2002 gotten from ppu over bus
	OamAddress *Register // 0x2003
	OamData    *Register // 0x2004
	Scroll     *Register // 0x2005
	Address    *Register // 0x2006
	Data       *Register // 0x2007

	bus          *bus.Bus
	addressLatch byte
}

func CreatePpuRegisters(b *bus.Bus) *PpuRegisters {
	return &PpuRegisters{
		Ctrl:       &PpuCtrl{},
		Mask:       &Register{},
		OamAddress: &Register{},
		OamData:    &Register{},
		Scroll:     &Register{},
		Address:    &Register{},
		Data:       &Register{},
		bus:        b,
	}
}

func (p *PpuRegisters) WriteByteAt(address uint16, value byte) {
	switch address {
	case 0x2000:
		p.bus.Ppu.WriteController(value)
	case 0x2001:
		p.bus.Ppu.WriteMask(value)
	case 0x2003:
		p.OamAddress.Write(value)
	case 0x2004:
		p.OamData.Write(value)
	case 0x2005:
		p.Scroll.Write(value)
	case 0x2006:
		p.addressLatch = p.Address.Read()
		p.Address.Write(value)
	case 0x2007:
		address := binary.LittleEndian.Uint16([]byte{p.addressLatch, p.Address.Read()})
		p.Data.Write(value)
		p.bus.Ppu.WriteByteAt(address, value)
		p.Address.Write(p.Address.Read() + p.Ctrl.VramIncrement())
	default:
		panic(errors.Wrapf(ErrInvalidAddress, "0x%04X", address))
	}
}

func (p *PpuRegisters) ReadByteAt(address uint16) byte {
	switch address {
	case 0x2002:
		return p.bus.Ppu.GetStatus()
	case 0x2004:
		return p.OamData.Read()
	case 0x2007:
		address := binary.LittleEndian.Uint16([]byte{p.addressLatch, p.Address.Read()})
		result := p.bus.Ppu.ReadByteAt(address)
		p.Address.Write(p.Address.Read() + p.Ctrl.VramIncrement())
		return result
	default:
		panic(errors.Wrapf(ErrInvalidAddress, "0x%04X", address))
	}
}
