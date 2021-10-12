package utility

import "encoding/binary"

func SetBit(n byte, pos byte, value bool) byte {
	if value {
		n |= (1 << pos)
	} else {
		mask := byte(^(1 << pos))
		n &= mask
	}

	return n
}

func BitSet(val byte, pos byte) bool {
	return (val & (1 << pos)) > 0
}

// Who needs generics lamo
func Wrapbyte(val, max byte) byte {
	if val > max {
		return val - max - 1
	}

	return val
}

func WrapUint16(val, max uint16) uint16 {
	if val > max {
		return val - max - 1
	}

	return val
}

func CombineToUint16(lower, higher byte) uint16 {
	return binary.LittleEndian.Uint16([]byte{lower, higher})
}
