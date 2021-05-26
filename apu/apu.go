package apu

import (
	"fmt"
)

var (
	ErrInvalidAddress = fmt.Errorf("invalid apu address")
)

type SoundRegister struct {
	data [4]byte
}

func (s *SoundRegister) WriteByteAt(address uint16, value byte) {
	s.data[address] = value
}

func (s *SoundRegister) ReadByteAt(address uint16) byte {
	return s.data[address]
}

type Pluse struct {
	SoundRegister
}

type Triangle struct {
	SoundRegister
}

type Noise struct {
	SoundRegister
}

type Dmc struct {
	SoundRegister
}

type Apu struct {
	Pluse1        *Pluse    // 0x4000 - 0x4003
	Pluse2        *Pluse    // 0x4004 - 0x4007
	Triangle      *Triangle // 0x4008 - 0x400B
	Noise         *Noise    // 0x400C - 0x400F
	Dmc           *Dmc      // 0x4010 - 0x4013
	ChannelEnable byte      // 0x4015
	FrameCounter  byte      // 0x4017
}

func Create() *Apu {
	return &Apu{
		Pluse1:   &Pluse{},
		Pluse2:   &Pluse{},
		Triangle: &Triangle{},
		Noise:    &Noise{},
		Dmc:      &Dmc{},
	}
}

func (a *Apu) WriteByteAt(address uint16, value byte) {
	switch {
	case address >= 0x4000 && address <= 0x4003:
		a.Pluse1.WriteByteAt(address-0x4000, value)
	case address >= 0x4004 && address <= 0x4007:
		a.Pluse2.WriteByteAt(address-0x4004, value)
	case address >= 0x4008 && address <= 0x400B:
		a.Triangle.WriteByteAt(address-0x4008, value)
	case address >= 0x400C && address <= 0x400F:
		a.Noise.WriteByteAt(address-0x400C, value)
	case address >= 0x4010 && address <= 0x4013:
		a.Dmc.WriteByteAt(address-0x4010, value)
	case address == 0x4015:
		a.ChannelEnable = value
	case address == 0x4017:
		a.FrameCounter = value
	default:
		panic(ErrInvalidAddress)
	}
}

func (a *Apu) ReadByteAt(address uint16) byte {
	switch {
	case address >= 0x4000 && address <= 0x4003:
		return a.Pluse1.ReadByteAt(address - 0x4000)
	case address >= 0x4004 && address <= 0x4007:
		return a.Pluse2.ReadByteAt(address - 0x4004)
	case address >= 0x4008 && address <= 0x400B:
		return a.Triangle.ReadByteAt(address - 0x4008)
	case address >= 0x400C && address <= 0x400F:
		return a.Noise.ReadByteAt(address - 0x400C)
	case address >= 0x4010 && address <= 0x4013:
		return a.Dmc.ReadByteAt(address - 0x4010)
	case address == 0x4015:
		return a.ChannelEnable
	case address == 0x4017:
		return a.FrameCounter
	default:
		panic(ErrInvalidAddress)
	}
}
