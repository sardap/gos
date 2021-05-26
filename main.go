package main

import (
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sardap/gos/emulator"
)

const (
	WindowWidth  = 1200
	WindowHeight = 720
)

type Gos struct {
}

func (g *Gos) Layout(outsideWidth, outsideHeight int) (int, int) {
	return WindowWidth, WindowHeight
}

func (g *Gos) Update() error {
	return nil
}

func (g *Gos) Draw(screen *ebiten.Image) {
}

func main() {
	e := emulator.Create()

	func() {
		f, err := os.Open("assets\\test_roms\\nestest.nes")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		if err := e.LoadRom(f); err != nil {
			panic(err)
		}
	}()

	ebiten.SetWindowSize(WindowWidth, WindowHeight)
	ebiten.SetWindowTitle("Go Boy")

	if err := ebiten.RunGame(&Gos{}); err != nil {
		log.Fatal(err)
	}
}
