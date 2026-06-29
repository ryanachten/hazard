package common

import "math/rand"

// RandIntInRange returns a random integer in range
func RandIntInRange(valueRange [2]int) int {
	low := valueRange[0]
	high := valueRange[1]
	return low + rand.Intn(high-low+1)
}

// RandomFloat returns a random float within range
func RandomFloat(low, high float64) float64 {
	return low + rand.Float64()*(high-low)
}

// RandValInSlice returns a random element from a non-empty slice
func RandValInSlice[T any](values []T) T {
	randomIndex := rand.Intn(len(values))
	return values[randomIndex]
}
