package memory_test

import (
	"testing"

	"github.com/sardap/gos/bus"
	"github.com/sardap/gos/memory"
	"github.com/sardap/gos/ppu"
	"github.com/stretchr/testify/assert"
)

func createMemory() *memory.Memory {
	b := &bus.Bus{}

	result := memory.Create(b)
	ppu.Create(b)

	return result
}

func TestRamMirroing(t *testing.T) {
	t.Parallel()

	m := createMemory()

	for i := uint16(0x0000); i < 0x07FF; i++ {
		value := byte(0x10)
		m.WriteByteAt(i, value)

		assert.Equal(t, value, m.ReadByteAt(i))
		assert.Equal(t, value, m.ReadByteAt(i+0x0800))
		assert.Equal(t, value, m.ReadByteAt(i+0x1000))
	}
}
