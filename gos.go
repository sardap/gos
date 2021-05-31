package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sardap/gos/emulator"
	nesmath "github.com/sardap/gos/math"
)

type Gos struct {
	emu      *emulator.Emulator
	tileMaps map[uint16]*ebiten.Image
	palettes map[int]map[int]color.Color
}

func (g *Gos) Layout(outsideWidth, outsideHeight int) (int, int) {
	return WindowWidth, WindowHeight
}

func (g *Gos) Update() error {
	g.emu.Step()
	return nil
}

func (g *Gos) Draw(screen *ebiten.Image) {

	options := &ebiten.DrawImageOptions{}
	options.GeoM.Scale(2, 2)
	options.GeoM.Translate(float64(0), float64(0))
	img := g.RenderPatternTable(0x0000)
	screen.DrawImage(img, options)

	options = &ebiten.DrawImageOptions{}
	options.GeoM.Scale(2, 2)
	options.GeoM.Translate(float64(img.Bounds().Max.X+100), float64(0))
	img = g.RenderPatternTable(0x1000)
	screen.DrawImage(img, options)
}

func (g *Gos) Color(palNumber int, num int) color.Color {
	if num > 3 || num < 0 {
		panic("invalid number")
	}

	return g.palettes[palNumber][num]
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
			left := g.emu.Bus.Ppu.ReadByteAt(address)
			right := g.emu.Bus.Ppu.ReadByteAt(address + 8)
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
