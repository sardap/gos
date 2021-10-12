package testing

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	"github.com/sardap/gos/pkg/cpu"
)

type operation struct {
	re     *regexp.Regexp
	opcode byte
}

var (
	operations []operation
)

func init() {
	operations = make([]operation, 0)

	for k, v := range cpu.GetOpcodes() {
		pattern := v.Name
		pattern = strings.Replace(pattern, "#oper", `\#(?P<IOprand>\d\d)`, 1)
		switch v.AddressMode {
		case cpu.AddressModeZeroPage, cpu.AddressModeZeroPageX, cpu.AddressModeZeroPageY:
			pattern = strings.Replace(pattern, "oper", `\$(?P<ZOprand>\d\d)`, 1)
		case cpu.AddressModeAbsolute, cpu.AddressModeAbsoluteX, cpu.AddressModeAbsoluteY:
			pattern = strings.Replace(pattern, "oper", `\$(?P<AOprand>\d\d\d\d)`, 1)
		}
		pattern = strings.Replace(pattern, "(oper,X)", `\(\$(?P<IXOprand>\d\d),X\)`, 1)
		pattern = strings.Replace(pattern, "(oper),Y", `\(\$(?P<IYOprand>\d\d)\),Y`, 1)

		re := regexp.MustCompile(fmt.Sprintf("^%s$", pattern))

		operations = append(operations, operation{re, k})
	}
}

func DumbAssemble(assembly string) ([]byte, error) {
	var result bytes.Buffer

	for _, line := range strings.Split(assembly, "\n") {
		line = strings.ReplaceAll(line, "\n", "")

		matchFound := false

		tmp := operations

		for _, op := range tmp {
			matches := op.re.FindAllStringSubmatch(line, -1)

			if len(matches) == 0 {
				continue
			}

			matchFound = true

			result.WriteByte(op.opcode)

			keys := op.re.SubexpNames()

			for i, match := range matches[0] {
				switch keys[i] {
				case "IOprand", "ZOprand", "AOprand", "IXOprand", "IYOprand":
					data, _ := hex.DecodeString(match)
					result.Write([]byte(data))
				}
			}
		}

		if !matchFound {
			return nil, fmt.Errorf("invalid assembly no match found %s", line)
		}
	}

	return result.Bytes(), nil
}
