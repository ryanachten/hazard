package random

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFloat(t *testing.T) {
	tests := []struct {
		name string
		low  float64
		high float64
	}{
		{name: "zero range", low: 5.0, high: 5.0},
		{name: "small range", low: 0.0, high: 1.0},
		{name: "negative to positive", low: -1.0, high: 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for range 100 {
				result := Float(tt.low, tt.high)
				require.GreaterOrEqual(t, result, tt.low)
				require.LessOrEqual(t, result, tt.high)
			}
		})
	}
}

func TestRandValInSlice(t *testing.T) {
	values := []int{10, 20, 30, 40, 50}

	seen := make(map[int]bool)
	for range 100 {
		result := ValInSlice(values)
		require.Contains(t, values, result)
		seen[result] = true
	}

	require.GreaterOrEqual(t, len(seen), 2, "random selection should cover multiple values over 100 trials")
}

func TestRandValInSlice_SingleElement(t *testing.T) {
	values := []string{"only"}
	result := ValInSlice(values)
	require.Equal(t, "only", result)
}
