package encoding

import "math"

func ZOrderEncodeFloat(x, y float32, bitsNum uint8) int64 {
	xInt64 := int64(math.Float32bits(x))
	yInt64 := int64(math.Float32bits(y))
	var z int64
	for i := uint8(0); i < bitsNum; i++ {
		z |= ((xInt64 & (1 << i)) << i) | ((yInt64 & (1 << i)) << (i + 1))
	}
	return z
}

func ZOrderDecodeFloat(z int64, bitsNum uint8) (float32, float32) {
	var x, y int64 = 0, 0
	for i := uint8(0); i < bitsNum; i++ {
		x |= (z & (1 << (i * 2))) >> i
		y |= (z & (1 << (i*2 + 1))) >> (i + 1)
	}
	return math.Float32frombits(uint32(x)), math.Float32frombits(uint32(y))
}

func ZOrderEncodeFloat754(x, y float32, bitsNum uint8) int64 {
	xInt64 := float32ToInt64(x)
	yInt64 := float32ToInt64(y)
	var z int64
	for i := uint8(0); i < bitsNum; i++ {
		z |= ((xInt64 & (1 << i)) << i) | ((yInt64 & (1 << i)) << (i + 1))
	}
	return z
}

func ZOrderDecodeFloat754(z int64, bitsNum uint8) (float32, float32) {
	var x, y int64 = 0, 0
	for i := uint8(0); i < bitsNum; i++ {
		x |= (z & (1 << (i * 2))) >> i
		y |= (z & (1 << (i*2 + 1))) >> (i + 1)
	}
	return int64ToFloat32(x), int64ToFloat32(y)
}

func ZOrderEncodeInt(x, y int64, bitsNum uint8) int64 {
	var z int64
	for i := uint8(0); i < bitsNum; i++ {
		z |= ((x & (1 << i)) << i) | ((y & (1 << i)) << (i + 1))
	}
	return z
}

func ZOrderDecodeInt(z int64, bitsNum uint8) (int64, int64) {
	var x, y int64 = 0, 0
	for i := uint8(0); i < bitsNum; i++ {
		x |= (z & (1 << (i * 2))) >> i
		y |= (z & (1 << (i*2 + 1))) >> (i + 1)
	}
	return x, y
}

func COrderEncodeFloat(x, y float32, bitsNum uint8) int64 {
	xBits := math.Float32bits(x)
	yBits := math.Float32bits(y)
	c := uint64(xBits)<<bitsNum | uint64(yBits)
	return int64(c)
}

func COrderDecodeFloat(c int64, bitsNum uint8) (float32, float32) {
	mergedBits := uint64(c)
	xBits := mergedBits >> bitsNum
	yBits := mergedBits & ((1 << bitsNum) - 1)
	x := math.Float32frombits(uint32(xBits))
	y := math.Float32frombits(uint32(yBits))
	return x, y
}

func COrderEncodeInt(x, y int64, bitsNum uint8) int64 {
	c := x<<bitsNum | y
	return c
}

func COrderDecodeInt(c int64, bitsNum uint8) (int64, int64) {
	x := c >> bitsNum
	y := c & ((1 << bitsNum) - 1)
	return x, y
}

func IEEEEncodeFloat(x, y float32, bitsNum uint8) int64 {
	xInt64 := int64(math.Float32bits(x))
	yInt64 := int64(math.Float32bits(y))

	var encoded int64
	//sign
	signLoc := uint8(31)
	encoded |= (((xInt64 >> signLoc) & 1) << 63) | (((yInt64 >> signLoc) & 1) << 62)
	//exponent
	encoded |= (((xInt64 >> 23) & 0xFF) << 54) | (((yInt64 >> 23) & 0xFF) << 46)
	//Mantissa
	encoded |= ((xInt64 & 0x7FFFFF) << 23) | ((yInt64) & 0x7FFFFF)

	return encoded
}

func IEEEDecodeFloat(encoded int64, bitsNum uint8) (float32, float32) {
	var x, y int64 = 0, 0

	//restore the sign bit
	x |= (encoded & (1 >> 63)) << 31
	y |= (encoded & (1 >> 62)) << 31

	//restore the exponent bits
	x |= ((encoded >> 54) & 0xFF) << 23
	y |= ((encoded >> 46) & 0xFF) << 23

	//restore the mantissa bits
	x |= (encoded >> 23) & 0x7FFFFF
	y |= encoded & 0x7FFFFF

	return math.Float32frombits(uint32(x)), math.Float32frombits(uint32(y))
}
