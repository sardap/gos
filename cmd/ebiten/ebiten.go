package main

import (
	"bytes"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sardap/gos/pkg/emulator"
	"github.com/sardap/gos/pkg/palette"
	"github.com/sardap/gos/pkg/utility"
)

var ntscpalettePalette = []byte{0x52, 0x52, 0x52, 0x1, 0x1a, 0x51, 0xf, 0xf, 0x65, 0x23, 0x6, 0x63, 0x36, 0x3, 0x4b, 0x40, 0x4, 0x26, 0x3f, 0x9, 0x4, 0x32, 0x13, 0x0, 0x1f, 0x20, 0x0, 0xb, 0x2a, 0x0, 0x0, 0x2f, 0x0, 0x0, 0x2e, 0xa, 0x0, 0x26, 0x2d, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xa0, 0xa0, 0xa0, 0x1e, 0x4a, 0x9d, 0x38, 0x37, 0xbc, 0x58, 0x28, 0xb8, 0x75, 0x21, 0x94, 0x84, 0x23, 0x5c, 0x82, 0x2e, 0x24, 0x6f, 0x3f, 0x0, 0x51, 0x52, 0x0, 0x31, 0x63, 0x0, 0x1a, 0x6b, 0x5, 0xe, 0x69, 0x2e, 0x10, 0x5c, 0x68, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xfe, 0xff, 0xff, 0x69, 0x9e, 0xfc, 0x89, 0x87, 0xff, 0xae, 0x76, 0xff, 0xce, 0x6d, 0xf1, 0xe0, 0x70, 0xb2, 0xde, 0x7c, 0x70, 0xc8, 0x91, 0x3e, 0xa6, 0xa7, 0x25, 0x81, 0xba, 0x28, 0x63, 0xc4, 0x46, 0x54, 0xc1, 0x7d, 0x56, 0xb3, 0xc0, 0x3c, 0x3c, 0x3c, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xfe, 0xff, 0xff, 0xbe, 0xd6, 0xfd, 0xcc, 0xcc, 0xff, 0xdd, 0xc4, 0xff, 0xea, 0xc0, 0xf9, 0xf2, 0xc1, 0xdf, 0xf1, 0xc7, 0xc2, 0xe8, 0xd0, 0xaa, 0xd9, 0xda, 0x9d, 0xc9, 0xe2, 0x9e, 0xbc, 0xe6, 0xae, 0xb4, 0xe5, 0xc7, 0xb5, 0xdf, 0xe4, 0xa9, 0xa9, 0xa9, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}

var (
	completePalette *palette.Palette
	blankPalette    = map[int]color.Color{
		0: color.RGBA{0, 0, 0, 0},
		1: color.RGBA{64, 64, 64, 255},
		2: color.RGBA{128, 128, 128, 255},
		3: color.RGBA{255, 255, 255, 255},
	}
)

func init() {
	var err error
	completePalette, err = palette.Parse(bytes.NewBuffer(ntscpalettePalette))
	if err != nil {
		panic(err)
	}
}

type Gos struct {
	emu         *emulator.Emulator
	tileMaps    map[uint16]*ebiten.Image
	paletteImgs map[uint16]*ebiten.Image
}

func (g *Gos) Layout(outsideWidth, outsideHeight int) (int, int) {
	return WindowWidth, WindowHeight
}

func (g *Gos) Update() error {
	g.emu.Step()
	return nil
}

func (g *Gos) Draw(screen *ebiten.Image) {

	var options *ebiten.DrawImageOptions
	var img *ebiten.Image
	var x int

	// palettes
	if g.emu.Ppu.PaletteDirty() {
		g.paletteImgs = make(map[uint16]*ebiten.Image)
	}
	options = &ebiten.DrawImageOptions{}
	options.GeoM.Scale(32, 32)
	options.GeoM.Translate(float64(x), float64(0))
	img = g.RenderPalette(0x3F01)
	screen.DrawImage(img, options)
	x += img.Bounds().Max.X

	x = 700

	// tiles
	options = &ebiten.DrawImageOptions{}
	options.GeoM.Scale(2, 2)
	options.GeoM.Translate(float64(x), float64(0))
	img = g.RenderPatternTable(0x0000)
	screen.DrawImage(img, options)
	x += img.Bounds().Max.X

	options = &ebiten.DrawImageOptions{}
	options.GeoM.Scale(2, 2)
	options.GeoM.Translate(float64(x+100), float64(0))
	img = g.RenderPatternTable(0x1000)
	screen.DrawImage(img, options)
	x += img.Bounds().Max.X
}

func (g *Gos) Color(palNumber int, num int) color.Color {
	if num > 3 || num < 0 {
		panic("invalid number")
	}

	return blankPalette[num]
}

func (g *Gos) RenderPalette(baseAddress uint16) *ebiten.Image {
	if g.paletteImgs[baseAddress] != nil {
		return g.paletteImgs[baseAddress]
	}

	palImg := ebiten.NewImage(4, 1)
	palImg.Set(0, 0, completePalette.ColorForInt(int(g.emu.Ppu.ReadByteAt(0x3F00))))
	for i := uint16(1); i < 4; i++ {
		value := g.emu.Ppu.ReadByteAt(baseAddress + i)
		palImg.Set(int(i), 0, completePalette.ColorForInt(int(value)))
	}

	g.paletteImgs[baseAddress] = palImg
	return g.paletteImgs[baseAddress]
}

func (g *Gos) RenderPatternTable(baseAddress uint16) *ebiten.Image {
	if g.tileMaps[baseAddress] == nil {
		g.tileMaps[baseAddress] = ebiten.NewImage(10*8, 256*8)
	}

	img := g.tileMaps[baseAddress]
	img.Fill(color.Transparent)

	palNumber := 0

	for i := uint16(0); i < 0x0FFF; i += 16 {
		tid := i / 16
		x := tid % 10 * 8
		y := tid / 10 * 8

		for j := uint16(0); j < 8; j++ {
			address := baseAddress + i + j
			left := g.emu.Ppu.ReadByteAt(address)
			right := g.emu.Ppu.ReadByteAt(address + 8)
			for k := byte(0); k < 8; k++ {
				var value byte
				value = utility.SetBit(value, 0, utility.BitSet(left, k))
				value = utility.SetBit(value, 1, utility.BitSet(right, k))
				subY := int(j)
				img.Set(int(x)+int(k), int(y)+subY, g.Color(palNumber, int(value)))
			}
		}
	}

	g.tileMaps[baseAddress] = img

	return g.tileMaps[baseAddress]
}
