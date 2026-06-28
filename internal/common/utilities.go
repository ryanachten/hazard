package common

import "math/rand"

// RandIntInRange returns a random integer in range
func RandIntInRange(valueRange [2]int) int {
	low := valueRange[0]
	high := valueRange[1]
	return low + rand.Intn(high-low+1)
}

// RandValInSlice returns a random element from a non-empty slice
func RandValInSlice[T any](values []T) T {
	randomIndex := rand.Intn(len(values))
	return values[randomIndex]
}
