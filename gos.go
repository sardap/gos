package main

import (
	"image/color"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sardap/gos/emulator"
	nesmath "github.com/sardap/gos/math"
	"github.com/sardap/gos/palette"
)

var (
	completePalette *palette.Palette
	blankPallete    = map[int]color.Color{
		0: color.RGBA{0, 0, 0, 0},
		1: color.RGBA{64, 64, 64, 255},
		2: color.RGBA{128, 128, 128, 255},
		3: color.RGBA{255, 255, 255, 255},
	}
)

func init() {
	f, err := os.Open("assets/palettes/ntscpalette.pal")
	if err != nil {
		panic(err)
	}

	completePalette, err = palette.Parse(f)
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

	// pallettes
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

	return blankPallete[num]
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
	if g.tileMaps[baseAddress] != nil {
		return g.tileMaps[baseAddress]
	}

	g.tileMaps[baseAddress] = ebiten.NewImage(10*8, 256*8)

	palNumber := 0

	for i := uint16(0); i < 0x0FFF; i += 16 {
		tile := ebiten.NewImage(8, 8)

		for j := uint16(0); j < 8; j++ {
			address := baseAddress + i + j
			left := g.emu.Ppu.ReadByteAt(address)
			right := g.emu.Ppu.ReadByteAt(address + 8)
			for k := byte(0); k < 8; k++ {
				var value byte
				value = nesmath.SetBit(value, 0, nesmath.BitSet(left, k))
				value = nesmath.SetBit(value, 1, nesmath.BitSet(right, k))
				y := int(j)
				tile.Set(int(k), y, g.Color(palNumber, int(value)))
			}
		}

		tid := i / 16

		options := &ebiten.DrawImageOptions{}
		options.GeoM.Scale(1, 1)
		x := tid % 10 * 8
		y := tid / 10 * 8
		options.GeoM.Translate(float64(x), float64(y))
		g.tileMaps[baseAddress].DrawImage(tile, options)
	}

	return g.tileMaps[baseAddress]
}
