package cart

import (
	"fmt"

	"github.com/sardap/gos/pkg/utility"
)

var (
	ErrInvalidRom = fmt.Errorf("invalid rom header")
)

type MirrorType int

const (
	MirrorTypeVertical MirrorType = iota
	MirrorTypeHorizontal
)

type ControlByte1 struct {
	MapperLowerBits      byte
	FourScreenVramLayout bool
	Trainer512Byte       bool // 0x7000 - 0x71FF
	BatteryRam           bool // 0x6000 - 0x7FFF
	MirrorType           MirrorType
}

func CreateControlByte1(data byte) *ControlByte1 {
	result := &ControlByte1{}

	mapperLower := byte(0)
	mapperLower = utility.SetBit(mapperLower, 0, utility.BitSet(data, 4))
	mapperLower = utility.SetBit(mapperLower, 1, utility.BitSet(data, 5))
	mapperLower = utility.SetBit(mapperLower, 2, utility.BitSet(data, 6))
	mapperLower = utility.SetBit(mapperLower, 3, utility.BitSet(data, 7))
	result.MapperLowerBits = mapperLower

	result.FourScreenVramLayout = utility.BitSet(3, data)
	result.Trainer512Byte = utility.BitSet(2, data)
	result.BatteryRam = utility.BitSet(1, data)

	if utility.BitSet(0, data) {
		result.MirrorType = MirrorTypeVertical
	} else {
		result.MirrorType = MirrorTypeHorizontal
	}

	return result
}

type INesFormatType int

const (
	INesFormatType1 INesFormatType = iota
	INesFormatType2
)

type ControlByte2 struct {
	MapperHigherBits byte
	INesFormat       INesFormatType
}

func CreateControlByte2(data byte) (*ControlByte2, error) {
	result := &ControlByte2{}

	mapperHigher := byte(0)
	mapperHigher = utility.SetBit(mapperHigher, 4, utility.BitSet(data, 4))
	mapperHigher = utility.SetBit(mapperHigher, 5, utility.BitSet(data, 5))
	mapperHigher = utility.SetBit(mapperHigher, 6, utility.BitSet(data, 6))
	mapperHigher = utility.SetBit(mapperHigher, 7, utility.BitSet(data, 7))
	result.MapperHigherBits = mapperHigher

	if utility.BitSet(data, 3) && !utility.BitSet(data, 2) {
		result.INesFormat = INesFormatType2
	} else if !utility.BitSet(data, 3) && !utility.BitSet(data, 2) {
		result.INesFormat = INesFormatType1
	} else {
		return nil, ErrInvalidRom
	}

	return result, nil
}

type CartInfo struct {
	PrgRomBanks  byte
	ChrRomBanks  byte
	ControlByte1 *ControlByte1
	ControlByte2 *ControlByte2
	PrgRamLength byte
}

type Cart interface {
	WriteBytesPrg(value []byte) error
	ReadByteChrAt(address uint16) byte
	WriteBytesChr(value []byte) error
	WriteByteAt(address uint16, value byte)
	ReadByteAt(address uint16) byte
}

type TestCart struct {
	data [0x10000]byte
}

func (c *TestCart) WriteBytesPrg(value []byte) error {
	return nil
}

func (c *TestCart) ReadByteChrAt(offset uint16) byte {
	return c.data[offset]
}

func (c *TestCart) WriteBytesChr(value []byte) error {
	return nil
}

func (c *TestCart) WriteByteAt(address uint16, value byte) {
	c.data[address] = value
}

func (c *TestCart) ReadByteAt(address uint16) byte {
	return c.data[address]
}

type NRom struct {
	Prg       []byte
	Chr       []byte
	mapper    byte
	mirroring MirrorType
}

func CreateCart(info CartInfo) *NRom {
	result := &NRom{}

	result.mirroring = info.ControlByte1.MirrorType
	result.mapper = info.ControlByte2.MapperHigherBits | info.ControlByte1.MapperLowerBits

	return result
}

func (c *NRom) ReadByteChrAt(offset uint16) byte {
	return c.Chr[offset]
}

func (c *NRom) WriteBytesPrg(value []byte) error {
	c.Prg = append(c.Prg, value...)
	return nil
}

func (c *NRom) WriteBytesChr(value []byte) error {
	c.Chr = append(c.Chr, value...)
	return nil
}

func (c *NRom) WriteByteAt(address uint16, value byte) {
	switch {
	// PRG ram
	case address >= 0x6000 && address <= 0x7FFF:
		c.Prg[address-0x6000] = value
	case address >= 0x8000 && address <= 0xBFFF:
		c.Prg[address-0x8000] = value
	case address >= 0xC000 && address <= 0xFFFF:
		c.Prg[address-0xC000] = value
	default:
		panic("fuck")
	}
}

func (c *NRom) ReadByteAt(address uint16) byte {
	switch {
	// PRG ram
	case address >= 0x6000 && address <= 0x7FFF:
		return c.Prg[address-0x6000]
	case address >= 0x8000 && address <= 0xBFFF:
		return c.Prg[address-0x8000]
	case address >= 0xC000 && address <= 0xFFFF:
		return c.Prg[address-0xC000]
	default:
		panic(fmt.Sprintf("%04X", address))
	}
}
