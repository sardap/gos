package memory

import (
	"encoding/binary"

	"github.com/pkg/errors"

	nesmath "github.com/sardap/gos/math"
	"github.com/sardap/gos/ppu"
)

type PpuFlag byte

const (
	// https://wiki.nesdev.com/w/index.php/PPU_registers#PPUCTRL
	PpuFlagCtrlVblankInterval     PpuFlag = 7
	PpuFlagCtrlMasterSlaveSel     PpuFlag = 6
	PpuFlagCtrlSpirteSize         PpuFlag = 5
	PpuFlagCtrlBackgroundAddress  PpuFlag = 4
	PpuFlagCtrlSpirtePatternTable PpuFlag = 3
	PpuFlagCtrlVramAddress        PpuFlag = 2
	// https://wiki.nesdev.com/w/index.php/PPU_registers#PPUMASK
	PpuFlagMaskEmphasizeBlue          PpuFlag = 7
	PpuFlagMaskEmphasizeGreen         PpuFlag = 6
	PpuFlagMaskEmphasizeRed           PpuFlag = 5
	PpuFlagMaskShowSprites            PpuFlag = 4
	PpuFlagMaskShowBackground         PpuFlag = 3
	PpuFlagMaskShowSpirtesLeftmost    PpuFlag = 2
	PpuFlagMaskShowBackgroundLeftmost PpuFlag = 1
	PpuFlagMaskGreyScale              PpuFlag = 0
	// https://wiki.nesdev.com/w/index.php/PPU_registers#PPUSTATUS
	PpuFlagStatusVerticalBlank  PpuFlag = 7
	PpuFlagStatusSpirt0Hit      PpuFlag = 6
	PpuFlagStatusSpirteOverflow PpuFlag = 5
)

type PpuRegister struct {
	val byte
}

func (p *PpuRegister) Write(value byte) {
	p.val = value
}

func (p *PpuRegister) Read() byte {
	return p.val
}

func (p *PpuRegister) BitSet(flag PpuFlag) bool {
	return nesmath.BitSet(p.Read(), byte(flag))
}

type PpuCtrl struct {
	PpuRegister
}

func (p *PpuCtrl) VramIncrement() byte {
	if nesmath.BitSet(p.val, 2) {
		// Down
		return 32
	}

	// Across
	return 1
}

func (p *PpuCtrl) NameTableAddress() byte {
	flag := byte(0)
	flag = nesmath.SetBit(flag, 0, nesmath.BitSet(p.val, 0))
	flag = nesmath.SetBit(flag, 1, nesmath.BitSet(p.val, 1))

	return flag
}

type PpuWrite struct {
	Address uint16
	Value   byte
}

// The Ppu and CPU are on diffrent bus so I also seprated the objects
type PpuRegisters struct {
	Ctrl       *PpuCtrl     // 0x2000
	Mask       *PpuRegister // 0x2001
	Status     *PpuRegister // 0x2002
	OamAddress *PpuRegister // 0x2003
	OamData    *PpuRegister // 0x2004
	Scroll     *PpuRegister // 0x2005
	Address    *PpuRegister // 0x2006
	Data       *PpuRegister // 0x2007

	Ppu          *ppu.Ppu
	addressLatch byte
}

func CreatePpuRegisters() *PpuRegisters {
	return &PpuRegisters{
		Ctrl:       &PpuCtrl{},
		Mask:       &PpuRegister{},
		Status:     &PpuRegister{},
		OamAddress: &PpuRegister{},
		OamData:    &PpuRegister{},
		Scroll:     &PpuRegister{},
		Address:    &PpuRegister{},
		Data:       &PpuRegister{},
	}
}

func (p *PpuRegisters) WriteByteAt(address uint16, value byte) {
	switch address {
	case 0x2000:
		p.Ctrl.Write(value)
	case 0x2001:
		p.Mask.Write(value)
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
		p.Ppu.WriteByteAt(address, value)
		p.Address.Write(p.Address.Read() + p.Ctrl.VramIncrement())
	default:
		panic(errors.Wrapf(ErrInvalidAddress, "0x%04X", address))
	}
}

func (p *PpuRegisters) ReadByteAt(address uint16) byte {
	switch address {
	case 0x2002:
		return p.Status.Read()
	case 0x2004:
		return p.OamData.Read()
	case 0x2007:
		address := binary.LittleEndian.Uint16([]byte{p.addressLatch, p.Address.Read()})
		result := p.Ppu.ReadByteAt(address)
		p.Address.Write(p.Address.Read() + p.Ctrl.VramIncrement())
		return result
	default:
		panic(errors.Wrapf(ErrInvalidAddress, "0x%04X", address))
	}
}
