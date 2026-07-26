package ranging

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateRange(t *testing.T) {
	t.Run("valid range", func(t *testing.T) {
		err := ValidateRange(0, 5)
		require.NoError(t, err)
	})

	t.Run("negative min", func(t *testing.T) {
		err := ValidateRange(-1, 5)
		require.Error(t, err)
		require.Contains(t, err.Error(), "min must be at least 0")
	})

	t.Run("min greater than max", func(t *testing.T) {
		err := ValidateRange(5, 3)
		require.Error(t, err)
		require.Contains(t, err.Error(), "must be less than or equal to")
	})
}

func TestValidatePositiveRange(t *testing.T) {
	t.Run("valid positive range", func(t *testing.T) {
		err := ValidatePositiveRange(1, 5)
		require.NoError(t, err)
	})

	t.Run("zero min", func(t *testing.T) {
		err := ValidatePositiveRange(0, 5)
		require.Error(t, err)
		require.Contains(t, err.Error(), "min must be positive")
	})

	t.Run("negative min", func(t *testing.T) {
		err := ValidatePositiveRange(-1, 5)
		require.Error(t, err)
		require.Contains(t, err.Error(), "min must be positive")
	})

	t.Run("min greater than max", func(t *testing.T) {
		err := ValidatePositiveRange(5, 3)
		require.Error(t, err)
		require.Contains(t, err.Error(), "must be less than or equal to")
	})
}

func TestRangeRandom(t *testing.T) {
	tests := []struct {
		name string
		r    Range
	}{
		{name: "zero range", r: Range{Min: 5, Max: 5}},
		{name: "positive range", r: Range{Min: 1, Max: 10}},
		{name: "starting at zero", r: Range{Min: 0, Max: 5}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for range 100 {
				result := tt.r.Random()
				require.GreaterOrEqual(t, result, tt.r.Min)
				require.LessOrEqual(t, result, tt.r.Max)
			}
		})
	}
}

func TestPositiveRangeRandom(t *testing.T) {
	tests := []struct {
		name string
		r    PositiveRange
	}{
		{name: "zero range", r: PositiveRange{Min: 5, Max: 5}},
		{name: "positive range", r: PositiveRange{Min: 1, Max: 10}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for range 100 {
				result := tt.r.Random()
				require.GreaterOrEqual(t, result, tt.r.Min)
				require.LessOrEqual(t, result, tt.r.Max)
			}
		})
	}
}
