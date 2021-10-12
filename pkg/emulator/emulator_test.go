package emulator_test

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
	"testing"

	"github.com/sardap/gos/pkg/cart"
	"github.com/sardap/gos/pkg/cpu"
	"github.com/sardap/gos/pkg/emulator"
	"github.com/stretchr/testify/assert"
)

var (
	nesTestPath         = "test_roms"
	nesTestRomPath      = filepath.Join(nesTestPath, "nestest.nes")
	nesTestValidLogPath = filepath.Join(nesTestPath, "nestest-valid.txt")
	oamReadPath         = filepath.Join(nesTestPath, "oam_read.nes")
	testRomMutex        = &sync.Mutex{}
	numRegex            = regexp.MustCompile("[ ,a-zA-Z]")
)

func init() {
	log.SetOutput(ioutil.Discard)

	// Download nestest
	_, err := os.Stat(nesTestPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.Mkdir(nesTestPath, os.ModeDir); err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}

	_, err = os.Stat(nesTestRomPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = download(
				nesTestRomPath,
				"https://www.qmtpro.com/~nes/misc/nestest.nes",
			)
			if err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}

	_, err = os.Stat(nesTestValidLogPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = download(
				nesTestValidLogPath,
				"https://www.qmtpro.com/~nes/misc/nestest.log",
			)
			if err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}

	_, err = os.Stat(oamReadPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = downloadAndExtract(
				oamReadPath,
				"http://blargg.8bitalley.com/parodius/nes-tests/oam_read.zip",
				"oam_read/oam_read.nes",
			)
			if err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}
}

func downloadAndExtract(path, url, extractPath string) error {
	var zipFile bytes.Buffer

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	io.Copy(&zipFile, res.Body)

	reader := bytes.NewReader(zipFile.Bytes())
	zipR, err := zip.NewReader(reader, reader.Size())
	if err != nil {
		fmt.Printf("invalid zip given\n")
		os.Exit(1)
	}

	var rc io.ReadCloser

	for _, file := range zipR.File {
		if file.Name != extractPath {
			continue
		}

		rc, _ = file.Open()
		defer rc.Close()
		break
	}

	target, err := os.Create(path)
	if err != nil {
		return err
	}
	defer target.Close()

	io.Copy(target, rc)

	return err
}

func download(path, url string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	_, err = io.Copy(f, res.Body)

	return err
}

type status struct {
	PC       uint16
	Opcode   byte
	A        byte
	X        byte
	Y        byte
	P        byte
	SP       byte
	PpuLeft  int
	PpuRight int
	Cyc      int64
}

func (n status) String() string {
	return fmt.Sprintf(
		"PC:%04X Opcode:%02X A:%02X X:%02X Y:%02X P:%02X SP:%02X Cyc:%d",
		n.PC, n.Opcode, n.A, n.X, n.Y, n.P, n.SP, n.Cyc,
	)
}

func parseNesTestLine(line string) status {
	result := status{}

	r := bufio.NewReader(bytes.NewBufferString(line))

	// PC
	buffer := make([]byte, 4)
	r.Read(buffer)
	value, _ := strconv.ParseInt(string(buffer), 16, 32)
	result.PC = uint16(value)

	// Opcode
	r.ReadByte()
	r.ReadByte()
	buffer = make([]byte, 2)
	r.Read(buffer)
	value, _ = strconv.ParseInt(string(buffer), 16, 32)
	result.Opcode = byte(value)

	// A
	r.ReadString(':')
	buffer = make([]byte, 2)
	r.Read(buffer)
	value, _ = strconv.ParseInt(string(buffer), 16, 32)
	result.A = byte(value)

	// X
	r.ReadString(':')
	buffer = make([]byte, 2)
	r.Read(buffer)
	value, _ = strconv.ParseInt(string(buffer), 16, 32)
	result.X = byte(value)

	// Y
	r.ReadString(':')
	buffer = make([]byte, 2)
	r.Read(buffer)
	value, _ = strconv.ParseInt(string(buffer), 16, 32)
	result.Y = byte(value)

	// P
	r.ReadString(':')
	buffer = make([]byte, 2)
	r.Read(buffer)
	value, _ = strconv.ParseInt(string(buffer), 16, 32)
	result.P = byte(value)

	// SP
	r.ReadString(':')
	buffer = make([]byte, 2)
	r.Read(buffer)
	value, _ = strconv.ParseInt(string(buffer), 16, 32)
	result.SP = byte(value)

	// Ppu
	r.ReadString(':')
	buffer, _ = r.ReadBytes(',')
	buffer = numRegex.ReplaceAll(buffer, []byte(""))
	value, _ = strconv.ParseInt(string(buffer), 10, 32)
	result.PpuLeft = int(value)

	buffer, _ = r.ReadBytes('C')
	buffer = numRegex.ReplaceAll(buffer, []byte(""))
	value, _ = strconv.ParseInt(string(buffer), 10, 32)
	result.PpuRight = int(value)

	// Cyc
	r.ReadString(':')
	buffer, _ = r.ReadBytes(' ')
	value, _ = strconv.ParseInt(string(buffer), 10, 64)
	result.Cyc = value
	return result
}

func emulatorToTestLine(e *emulator.Emulator, cycles int64) status {
	return status{
		PC:     e.Cpu.Registers.PC,
		Opcode: e.Memory.ReadByteAt(e.Cpu.Registers.PC),
		A:      e.Cpu.Registers.A,
		X:      e.Cpu.Registers.X,
		Y:      e.Cpu.Registers.Y,
		P:      e.Cpu.Registers.P.Read(),
		SP:     e.Cpu.Registers.SP,
		Cyc:    cycles,
	}
}

func assertFlags(t *testing.T, expect, act byte) {
	expectP := cpu.CreateFlagRegister(expect)
	actP := cpu.CreateFlagRegister(act)

	assert.Equal(t, expectP.ReadFlag(cpu.FlagNegative), actP.ReadFlag(cpu.FlagNegative), "Negative")
	assert.Equal(t, expectP.ReadFlag(cpu.FlagOverflow), actP.ReadFlag(cpu.FlagOverflow), "Overflow")
	assert.Equal(t, expectP.ReadFlag(cpu.FlagDecimal), actP.ReadFlag(cpu.FlagDecimal), "Decimal")
	assert.Equal(t, expectP.ReadFlag(cpu.FlagInteruprtDisable), actP.ReadFlag(cpu.FlagInteruprtDisable), "Interuprt Disable")
	assert.Equal(t, expectP.ReadFlag(cpu.FlagZero), actP.ReadFlag(cpu.FlagZero), "Zero")
	assert.Equal(t, expectP.ReadFlag(cpu.FlagCarry), actP.ReadFlag(cpu.FlagCarry), "Carry")
}

func TestNestestRom(t *testing.T) {
	t.Parallel()

	e := emulator.Create()
	e.PpuEnabled = false

	var scanner *bufio.Scanner
	func() {
		testRomMutex.Lock()
		defer testRomMutex.Unlock()
		// Test loading rom
		romBytes, _ := os.ReadFile(nesTestRomPath)
		e.LoadRom(bytes.NewBuffer(romBytes))

		// Run
		testRomLog, _ := os.ReadFile(nesTestValidLogPath)
		scanner = bufio.NewScanner(bytes.NewBuffer(testRomLog))
	}()

	cycles := int64(7)
	lineNum := 1
	for scanner.Scan() && !t.Failed() {
		line := scanner.Text()
		statusValid := parseNesTestLine(string(line))
		statusEmu := emulatorToTestLine(e, cycles)

		fmt.Printf("line %04d Valid %s\n", lineNum, statusValid)
		fmt.Printf("line %04d Mine  %s\n", lineNum, statusEmu)

		if lineNum == 36 {
			fmt.Printf("fuck\n")
		}

		opcode := statusEmu.Opcode
		// Program Counters
		assert.Equalf(t, statusValid.Opcode, opcode, "Line:%d Opcode:%02X Opcode", lineNum, opcode)
		assert.Equalf(t, statusValid.PC, statusEmu.PC, "Line:%d Opcode:%02X Program Counter", lineNum, opcode)
		assert.Equalf(t, statusValid.Cyc, cycles, "Line:%d Opcode:%02X Cycles", lineNum, opcode)
		// Regsiters
		assert.Equalf(t, statusValid.A, statusEmu.A, "Line:%d Opcode:%02X A", lineNum, opcode)
		assert.Equalf(t, statusValid.X, statusEmu.X, "Line:%d Opcode:%02X X", lineNum, opcode)
		assert.Equalf(t, statusValid.Y, statusEmu.Y, "Line:%d Opcode:%02X Y", lineNum, opcode)
		assert.Equalf(t, statusValid.P, statusEmu.P, "Line:%d Opcode:%02X P", lineNum, opcode)
		assertFlags(t, statusValid.P, statusEmu.P)
		// PPU
		// assert.Equalf(t, statusValid.PpuLeft, statusEmu.PpuLeft, "Line:%d Opcode:%02X Ppu Left", lineNum, opcode)
		// assert.Equalf(t, statusValid.PpuRight, statusEmu.PpuRight, "Line:%d Opcode:%02X Ppu Rightt", lineNum, opcode)

		e.Step()
		cycles += int64(e.Cpu.Cycles)

		lineNum++
	}
}

func TestOamRead(t *testing.T) {
	log.SetOutput(os.Stdout)

	t.Parallel()

	e := emulator.Create()
	e.PpuEnabled = false

	func() {
		testRomMutex.Lock()
		defer testRomMutex.Unlock()
		// Test loading rom
		romBytes, _ := os.ReadFile(oamReadPath)
		e.LoadRom(bytes.NewBuffer(romBytes))
	}()

	for {
		e.Step()
	}
}

func TestRtsTrick(t *testing.T) {
	t.Parallel()

	e := emulator.Create()
	e.Bus.Cart = &cart.TestCart{}
	e.Cpu.Registers.PC = 0xC0E0

	// 0xC0E0 JSR $8000
	e.Memory.WriteByteAt(0xC0E0, 0x20)
	e.Memory.WriteByteAt(0xC0E1, 0x00)
	e.Memory.WriteByteAt(0xC0E2, 0x80)
	// 0xC0E3 LDX #$00
	e.Memory.WriteByteAt(0xC0E3, 0xA2)
	e.Memory.WriteByteAt(0xC0E4, 0x00)
	// 0x8000 LDA #$0F
	e.Memory.WriteByteAt(0x8000, 0xA9)
	e.Memory.WriteByteAt(0x8001, 0x0F)
	// 0x8002 STA #$1015
	e.Memory.WriteByteAt(0x8002, 0x8D)
	e.Memory.WriteByteAt(0x8003, 0x15)
	e.Memory.WriteByteAt(0x8004, 0x10)
	// 0x8005 RTS
	e.Memory.WriteByteAt(0x8005, 0x60)

	// Jsr
	e.Step()
	assert.Equal(t, uint16(0x8000), e.Cpu.Registers.PC)

	// LDA
	e.Step()
	assert.Equal(t, byte(0x0F), e.Cpu.Registers.A)

	// STA
	e.Step()
	assert.Equal(t, byte(0x0F), e.Memory.ReadByteAt(0x1015))

	// RTS
	e.Step()
	assert.Equal(t, uint16(0xC0E3), e.Cpu.Registers.PC)
}
