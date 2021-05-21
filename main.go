package main

import (
	"os"

	"github.com/sardap/gos/emulator"
)

func main() {
	e := emulator.Create()

	func() {
		f, err := os.Open("assets\\test_roms\\other\\nestest.nes")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		if err := e.LoadRom(f); err != nil {
			panic(err)
		}
	}()

	for {
		e.Step()
	}
}
