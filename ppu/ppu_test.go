package ppu_test

import (
	"testing"

	"github.com/sardap/gos/ppu"
	"github.com/stretchr/testify/assert"
)

func createPpu() *ppu.Ppu {
	return &ppu.Ppu{}
}

func TestPpuMirroing(t *testing.T) {
	t.Parallel()

	p := createPpu()

	// Name tables
	for i := uint16(0x2000); i < 0x2EFF; i++ {
		assert.Equalf(t, byte(0), p.ReadByteAt(i), "%04X", i)
		assert.Equalf(t, byte(0), p.ReadByteAt(i+0x1000), "%04X", i)

		value := byte(0x10)
		p.WriteByteAt(i, value)

		assert.Equalf(t, value, p.ReadByteAt(i), "%04X", i)
		assert.Equalf(t, value, p.ReadByteAt(i+0x1000), "%04X", i)

		value = byte(0x15)
		p.WriteByteAt(i+0x1000, value)

		assert.Equalf(t, value, p.ReadByteAt(i), "%04X", i)
		assert.Equalf(t, value, p.ReadByteAt(i+0x1000), "%04X", i)
	}

	// PaltteeRam
	for i := uint16(0x3F00); i < 0x3F1F; i++ {
		assert.Equalf(t, byte(0), p.ReadByteAt(i), "%04X", i)
		assert.Equalf(t, byte(0), p.ReadByteAt(i+0x020), "%04X", i)

		value := byte(0x20)
		p.WriteByteAt(i, value)

		assert.Equalf(t, value, p.ReadByteAt(i), "%04X", i)
		assert.Equalf(t, value, p.ReadByteAt(i+0x020), "%04X", i)

		value = byte(0x25)
		p.WriteByteAt(i+0x020, value)

		assert.Equalf(t, value, p.ReadByteAt(i), "%04X", i)
		assert.Equalf(t, value, p.ReadByteAt(i+0x020), "%04X", i)

	}
}
