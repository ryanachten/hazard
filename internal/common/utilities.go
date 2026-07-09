package common

import "math/rand"

// RandomFloat returns a random float within range
func RandomFloat(low, high float64) float64 {
	return low + rand.Float64()*(high-low)
}

// RandValInSlice returns a random element from a non-empty slice
func RandValInSlice[T any](values []T) T {
	randomIndex := rand.Intn(len(values))
	return values[randomIndex]
}
