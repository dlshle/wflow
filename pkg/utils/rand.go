package utils

import (
	"math/rand"
	"time"
)

func RandomNumber(lower, upper int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(upper-lower) + lower
}
