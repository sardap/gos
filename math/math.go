package math

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

func WrapUint16(val, max uint16) uint16 {
	if val > max {
		return val - max - 1
	}

	return val
}
