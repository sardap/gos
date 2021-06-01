package ppu

import (
	"fmt"

	"github.com/sardap/gos/bus"
	nesmath "github.com/sardap/gos/math"
)

const (
	vblankCycleCount = 2273
)

var (
	ErrInvalidAddress = fmt.Errorf("invalid ppu address")
)

type PpuFlag byte

const (
	// https://wiki.nesdev.com/w/index.php/PPU_registers#PPUSTATUS
	PpuFlagStatusVerticalBlank  PpuFlag = 7
	PpuFlagStatusSpirt0Hit      PpuFlag = 6
	PpuFlagStatusSpirteOverflow PpuFlag = 5
	// https://wiki.nesdev.com/w/index.php/PPU_registers#PPUMASK
	PpuFlagMaskEmphasizeBlue          PpuFlag = 7
	PpuFlagMaskEmphasizeGreen         PpuFlag = 6
	PpuFlagMaskEmphasizeRed           PpuFlag = 5
	PpuFlagMaskShowSprites            PpuFlag = 4
	PpuFlagMaskShowBackground         PpuFlag = 3
	PpuFlagMaskShowSpirtesLeftmost    PpuFlag = 2
	PpuFlagMaskShowBackgroundLeftmost PpuFlag = 1
	PpuFlagMaskGreyScale              PpuFlag = 0
	// https://wiki.nesdev.com/w/index.php/PPU_registers#PPUCTRL
	PpuFlagCtrlNmi                PpuFlag = 7
	PpuFlagCtrlMasterSlaveSel     PpuFlag = 6
	PpuFlagCtrlSpirteSize         PpuFlag = 5
	PpuFlagCtrlBackgroundAddress  PpuFlag = 4
	PpuFlagCtrlSpirtePatternTable PpuFlag = 3
	PpuFlagCtrlVramAddress        PpuFlag = 2
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

func (p *Register) FlagSet(flag PpuFlag) bool {
	return nesmath.BitSet(p.Read(), byte(flag))
}

func (p *Register) SetFlag(flag PpuFlag, value bool) {
	p.Write(nesmath.SetBit(p.Read(), byte(flag), value))
}

type Memory struct {
	PatternTable0 [0x1000]byte // 0x0000 - 0x0FFF
	PatternTable1 [0x1000]byte // 0x1000 - 0x1FFF
	NameTable0    [0x0400]byte // 0x2000 - 0x23FF
	NameTable1    [0x0400]byte // 0x2400 - 0x27FF
	NameTable2    [0x0400]byte // 0x2800 - 0x2BFF
	NameTable3    [0x0400]byte // 0x2C00 - 0x2FFF
	PalRam        [0x0020]byte // 0x3F00 - 0x3F1F
}

type state byte

const (
	stateRendering state = iota
	stateVblanking
)

type Ppu struct {
	Mem Memory

	controller *Register
	status     *Register
	mask       *Register
	oam        [256]byte

	paletteDirty    bool
	bus             *bus.Bus
	state           state
	vblankCountdown int
	even            bool
	x               int
	y               int
}

func Create(b *bus.Bus) *Ppu {
	result := &Ppu{
		status:     &Register{},
		mask:       &Register{},
		controller: &Register{},
		state:      stateRendering,
	}
	b.AddPpu(result)
	result.bus = b
	return result
}

func (p *Ppu) PaletteDirty() bool {
	result := p.paletteDirty
	p.paletteDirty = false
	return result
}

func (p *Ppu) WriteByteToOAM(offset byte, val byte) {
	p.oam[offset] = val
}

func (p *Ppu) GetStatus() byte {
	return p.status.Read()
}

func (p *Ppu) WriteMask(val byte) {
	p.mask.Write(val)
}

func (p *Ppu) WriteController(val byte) {
	p.controller.Write(val)
}

func (p *Ppu) Step(cycles int) {
	for i := 0; i < cycles*3; i++ {
		p.step()
	}
}

func (p *Ppu) RendingEnabled() bool {
	return p.mask.FlagSet(PpuFlagMaskShowSprites) || p.mask.FlagSet(PpuFlagMaskShowBackground)
}

func (p *Ppu) startVblank() {
	p.controller.SetFlag(PpuFlagCtrlNmi, true)
	p.status.SetFlag(PpuFlagStatusVerticalBlank, true)
	p.vblankCountdown = vblankCycleCount
	p.state = stateVblanking
	p.bus.Cpu.Interrupt(bus.InterruptTypeNim)
}

func (p *Ppu) endVblank() {
	oldVblank := p.status.FlagSet(PpuFlagStatusVerticalBlank)
	p.status.SetFlag(PpuFlagStatusVerticalBlank, false)
	p.controller.SetFlag(PpuFlagCtrlNmi, oldVblank)
	p.state = stateRendering
}

func (p *Ppu) step() {
	switch p.state {
	case stateRendering:
		// Rending disabled
		if !p.RendingEnabled() {
			p.y++
			if p.y > 262 {
				p.y = 0
				p.startVblank()
				return
			}
		}
	case stateVblanking:
		p.vblankCountdown--
		if p.vblankCountdown < 0 {
			p.endVblank()
		}
	}
}

func (p *Ppu) WriteByteAt(address uint16, value byte) {
	switch address & 0xF000 {
	case 0x2000:
		switch address & 0x0F00 {
		case 0x0000, 0x0100, 0x0200, 0x0300:
			p.Mem.NameTable0[address-0x2000] = value
		case 0x0400, 0x0500, 0x0600, 0x0700:
			p.Mem.NameTable1[address-0x2400] = value
		case 0x0800, 0x0900, 0x0A00, 0x0B00:
			p.Mem.NameTable2[address-0x2800] = value
		case 0x0C00, 0x0D00, 0x0E00, 0x0F00:
			p.Mem.NameTable3[address-0x2C00] = value
		}
	case 0x3000:
		switch {
		// Mirror of 0x2000 - 0x2EFF
		case address >= 0x3000 && address <= 0x3EFF:
			p.WriteByteAt(address-0x1000, value)
		case address >= 0x3F00 && address <= 0x3F1F:
			p.paletteDirty = true
			p.Mem.PalRam[address-0x3F00] = value
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
		return p.bus.Cart.ReadByteChrAt(address)
	case 0x1000:
		return p.bus.Cart.ReadByteChrAt(address)
	case 0x2000:
		switch address & 0x0F00 {
		case 0x0000, 0x0100, 0x0200, 0x0300:
			return p.Mem.NameTable0[address-0x2000]
		case 0x0400, 0x0500, 0x0600, 0x0700:
			return p.Mem.NameTable1[address-0x2400]
		case 0x0800, 0x0900, 0x0A00, 0x0B00:
			return p.Mem.NameTable2[address-0x2800]
		case 0x0C00, 0x0D00, 0x0E00, 0x0F00:
			return p.Mem.NameTable3[address-0x2C00]
		}
	case 0x3000:
		switch {
		// Mirror of 0x2000 - 0x2EFF
		case address >= 0x3000 && address <= 0x3EFF:
			return p.ReadByteAt(address - 0x1000)
		case address >= 0x3F00 && address <= 0x3F1F:
			return p.Mem.PalRam[address-0x3F00]
		// Mirror of 0x3F00 - 0x3F1F
		case address >= 0x3F20 && address <= 0x3FFF:
			return p.ReadByteAt(address - 0x0020)
		}
	}

	return p.ReadByteAt(address - 0x3FFF)
}
