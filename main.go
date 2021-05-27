package main

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/namsral/flag"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sardap/gos/emulator"
)

const (
	WindowWidth  = 1200
	WindowHeight = 720
)

type Gos struct {
	emu *emulator.Emulator
}

func (g *Gos) Layout(outsideWidth, outsideHeight int) (int, int) {
	return WindowWidth, WindowHeight
}

func (g *Gos) Update() error {
	g.emu.Step()
	return nil
}

func (g *Gos) Draw(screen *ebiten.Image) {
}

type args struct {
	romPath string
}

func parseArgs() (*args, error) {
	result := &args{}

	flag.StringVar(&result.romPath, "rom", "", "path of rom to launch")
	flag.Parse()

	if result.romPath == "" {
		return nil, fmt.Errorf("missing rom args")
	}

	return result, nil
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

	g := &Gos{}
	g.emu = emulator.Create()

	args, err := parseArgs()
	if err != nil {
		fmt.Printf("Error: %s\nTry -h\n", err.Error())
		os.Exit(1)
	}

	func() {
		// Load whole rom into ram because it's a nes game who gives a shit
		rom, err := os.ReadFile(args.romPath)
		if err != nil {
			fmt.Printf("cannot open %s\n", args.romPath)
			os.Exit(1)
		}

		fType := http.DetectContentType(rom)
		switch strings.TrimPrefix(fType, "application/") {
		case "x-gzip":
			func() {
				gzipR, err := gzip.NewReader(bytes.NewReader(rom))
				if err != nil {
					fmt.Printf("invalid zip given\n")
					os.Exit(1)
				}
				defer gzipR.Close()

				var buffer bytes.Buffer
				buffer.ReadFrom(gzipR)

				rom = buffer.Bytes()
			}()
		case "zip":
			func() {
				reader := bytes.NewReader(rom)
				zipR, err := zip.NewReader(reader, reader.Size())
				if err != nil {
					fmt.Printf("invalid zip given\n")
					os.Exit(1)
				}

				if len(zipR.File) != 1 {
					fmt.Printf("zip must contain only one file\n")
					os.Exit(1)
				}

				rc, err := zipR.File[0].Open()
				if err != nil {
					fmt.Printf("malformed target file in zip\n")
					os.Exit(1)
				}
				defer rc.Close()

				var buffer bytes.Buffer
				io.Copy(&buffer, rc)

				rom = buffer.Bytes()
			}()

		case "octet-stream":
		}
		fmt.Printf("%s\n", fType)

		g.emu.LoadRom(bytes.NewBuffer(rom))
	}()

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
