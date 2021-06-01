package palette_test

import (
	"image/color"
	"os"
	"testing"

	"github.com/sardap/gos/palette"
	"github.com/stretchr/testify/assert"
)

var (
	expectedMap = map[int]color.Color{
		0:  color.RGBA{82, 82, 82, 255},
		1:  color.RGBA{1, 26, 81, 255},
		2:  color.RGBA{15, 15, 101, 255},
		3:  color.RGBA{35, 6, 99, 255},
		4:  color.RGBA{54, 3, 75, 255},
		5:  color.RGBA{64, 4, 38, 255},
		6:  color.RGBA{63, 9, 4, 255},
		7:  color.RGBA{50, 19, 0, 255},
		8:  color.RGBA{31, 32, 0, 255},
		9:  color.RGBA{11, 42, 0, 255},
		10: color.RGBA{0, 47, 0, 255},
		11: color.RGBA{0, 46, 10, 255},
		12: color.RGBA{0, 38, 45, 255},
		13: color.RGBA{0, 0, 0, 255},
		14: color.RGBA{0, 0, 0, 255},
		15: color.RGBA{0, 0, 0, 255},
		16: color.RGBA{160, 160, 160, 255},
		17: color.RGBA{30, 74, 157, 255},
		18: color.RGBA{56, 55, 188, 255},
		19: color.RGBA{88, 40, 184, 255},
		20: color.RGBA{117, 33, 148, 255},
		21: color.RGBA{132, 35, 92, 255},
		22: color.RGBA{130, 46, 36, 255},
		23: color.RGBA{111, 63, 0, 255},
		24: color.RGBA{81, 82, 0, 255},
		25: color.RGBA{49, 99, 0, 255},
		26: color.RGBA{26, 107, 5, 255},
		27: color.RGBA{14, 105, 46, 255},
		28: color.RGBA{16, 92, 104, 255},
		29: color.RGBA{0, 0, 0, 255},
		30: color.RGBA{0, 0, 0, 255},
		31: color.RGBA{0, 0, 0, 255},
		32: color.RGBA{254, 255, 255, 255},
		33: color.RGBA{105, 158, 252, 255},
		34: color.RGBA{137, 135, 255, 255},
		35: color.RGBA{174, 118, 255, 255},
		36: color.RGBA{206, 109, 241, 255},
		37: color.RGBA{224, 112, 178, 255},
		38: color.RGBA{222, 124, 112, 255},
		39: color.RGBA{200, 145, 62, 255},
		40: color.RGBA{166, 167, 37, 255},
		41: color.RGBA{129, 186, 40, 255},
		42: color.RGBA{99, 196, 70, 255},
		43: color.RGBA{84, 193, 125, 255},
		44: color.RGBA{86, 179, 192, 255},
		45: color.RGBA{60, 60, 60, 255},
		46: color.RGBA{0, 0, 0, 255},
		47: color.RGBA{0, 0, 0, 255},
		48: color.RGBA{254, 255, 255, 255},
		49: color.RGBA{190, 214, 253, 255},
		50: color.RGBA{204, 204, 255, 255},
		51: color.RGBA{221, 196, 255, 255},
		52: color.RGBA{234, 192, 249, 255},
		53: color.RGBA{242, 193, 223, 255},
		54: color.RGBA{241, 199, 194, 255},
		55: color.RGBA{232, 208, 170, 255},
		56: color.RGBA{217, 218, 157, 255},
		57: color.RGBA{201, 226, 158, 255},
		58: color.RGBA{188, 230, 174, 255},
		59: color.RGBA{180, 229, 199, 255},
		60: color.RGBA{181, 223, 228, 255},
		61: color.RGBA{169, 169, 169, 255},
		62: color.RGBA{0, 0, 0, 255},
		63: color.RGBA{0, 0, 0, 255},
	}
)

func TestParse(t *testing.T) {
	t.Parallel()

	f, _ := os.Open("../assets/palettes/ntscpalette.pal")

	pal, err := palette.Parse(f)

	assert.Equal(t, nil, err)

	for k, v := range expectedMap {
		assert.Equal(t, pal.ColorForInt(k), v)
	}
}
