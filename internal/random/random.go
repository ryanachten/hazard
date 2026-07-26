// Package random provides random value generation utilities
package random

import "math/rand"

// Float returns a random float within range
func Float(low, high float64) float64 {
	return low + rand.Float64()*(high-low)
}

// ValInSlice returns a random element from a non-empty slice
func ValInSlice[T any](values []T) T {
	randomIndex := rand.Intn(len(values))
	return values[randomIndex]
}
