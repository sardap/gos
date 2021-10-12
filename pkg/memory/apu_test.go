package memory_test

import (
	"testing"

	"github.com/sardap/gos/pkg/memory"
	"github.com/stretchr/testify/assert"
)

func TestApuReadWrites(t *testing.T) {
	t.Parallel()

	a := memory.CreateApu()

	for i := uint16(0x4000); i <= 0x4017; i++ {
		// 0x4014 is just missing
		if i == 0x4014 || i == 0x4016 {
			continue
		}

		a.WriteByteAt(i, 0x02)
		assert.Equalf(t, byte(0x02), a.ReadByteAt(uint16(i)), "%02X", i)
	}
}
