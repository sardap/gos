package palette

import (
	"image/color"
	"io"
)

type Palette struct {
	mapping map[int]color.Color
}

func (p *Palette) ColorForInt(val int) color.Color {
	return p.mapping[val]
}

func Parse(r io.Reader) (*Palette, error) {
	result := &Palette{
		mapping: make(map[int]color.Color),
	}

	buffer := make([]byte, 3)
	for i := 0; i < 64*3; i += 3 {
		_, err := r.Read(buffer)
		if err != nil {
			return nil, err
		}

		result.mapping[i/3] = color.RGBA{
			R: buffer[0],
			G: buffer[1],
			B: buffer[2],
			A: 0xFF,
		}
	}

	return result, nil
}
