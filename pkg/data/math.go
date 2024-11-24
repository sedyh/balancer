package data

import "math"

func NextPowerOfTwo(x int) int {
	if x < 1 {
		return 1
	}
	return int(math.Pow(2, math.Ceil(math.Log2(float64(x)))))
}

func PrevPowerOfTwo(x int) int {
	if x < 1 {
		return 0
	}
	power := int64(1)
	for x > 1 {
		x >>= 1
		power <<= 1
	}
	return int(power)
}
