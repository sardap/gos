package cpu

import nesmath "github.com/sardap/gos/math"

const (
	StartingA  = 0
	StartingX  = 0
	StartingY  = 0
	StartingPC = 0
	StartingSP = 0xFF
	StartingP  = 0
)

type ByteRegister struct {
	val byte
}

func (r *ByteRegister) Write(value byte) {
	r.val = value
}

func (r *ByteRegister) Read() byte {
	return r.val
}

type FlagRegister struct {
	*ByteRegister
}

func CreateFlagRegister(value byte) *ByteRegister {
	return &ByteRegister{val: value}
}

type Flag byte

const (
	FlagCarry Flag = iota
	FlagZero
	FlagInteruprtDisable
	FlagDecimal
	FlagBreakCommand
	FlagUnsued
	FlagOverflow
	FlagNegative
)

func (f *FlagRegister) SetFlag(flag Flag, value bool) {
	f.val = nesmath.SetBit(f.val, byte(flag), value)
}

func (f *FlagRegister) ReadFlag(flag Flag) bool {
	return nesmath.BitSet(f.val, byte(flag))
}

func (f *FlagRegister) ReadFlagByte(flag Flag) byte {
	if f.ReadFlag(flag) {
		return 1
	}

	return 0
}

type Registers struct {
	A  byte
	X  byte
	Y  byte
	PC uint16
	SP byte
	P  *FlagRegister
}

func CreateRegisters() *Registers {
	return &Registers{
		A:  StartingA,
		X:  StartingX,
		Y:  StartingY,
		PC: StartingPC,
		SP: StartingSP,
		P:  &FlagRegister{ByteRegister: &ByteRegister{val: StartingP}},
	}
}
