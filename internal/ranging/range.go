// Package ranging defines numeric range utilities
package ranging

import (
	"fmt"
	"math/rand"
)

// Range represents an inclusive [min, max] range where min >= 0 and min <= max.
type Range struct {
	Min int
	Max int
}

// Validate checks that the range invariants hold for Range.
func Validate(minVal, maxVal int) error {
	if minVal < 0 {
		return fmt.Errorf("min must be at least 0, got %d", minVal)
	}
	if minVal > maxVal {
		return fmt.Errorf("min %d must be less than or equal to max %d", minVal, maxVal)
	}
	return nil
}

// Random returns a random integer within the range.
func (r Range) Random() int {
	return r.Min + rand.Intn(r.Max-r.Min+1)
}

// PositiveRange represents an inclusive [min, max] range where min > 0 and min <= max.
type PositiveRange struct {
	Min int
	Max int
}

// ValidatePositive checks that the range invariants hold for PositiveRange.
func ValidatePositive(minVal, maxVal int) error {
	if minVal <= 0 {
		return fmt.Errorf("min must be positive, got %d", minVal)
	}
	if minVal > maxVal {
		return fmt.Errorf("min %d must be less than or equal to max %d", minVal, maxVal)
	}
	return nil
}

// Random returns a random integer within the range.
func (r PositiveRange) Random() int {
	return r.Min + rand.Intn(r.Max-r.Min+1)
}
