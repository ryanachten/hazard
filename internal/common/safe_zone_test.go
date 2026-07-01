package common

import (
	pf "hazard/internal/pathfinding"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateSafeZone_PlacedAtOpenPosition(t *testing.T) {
	grid := pf.NewGrid(10, 10, pf.CellOpen)
	config := SafeZoneConfig{
		CountRange:  [2]int{1, 1},
		RadiusRange: [2]int{1, 1},
	}

	safeZone, err := CreateSafeZone(config, &grid)

	require.NoError(t, err)
	require.True(t, grid.InBounds(safeZone.Position))
	require.NotEmpty(t, safeZone.Cells)
	for _, cell := range safeZone.Cells {
		require.Equal(t, pf.CellSafeZone, grid.GetCell(cell),
			"cell %v must be marked as safe zone", cell)
	}
}

func TestCreateSafeZone_RadiusZeroMarksOnlyOrigin(t *testing.T) {
	grid := pf.NewGrid(5, 5, pf.CellOpen)
	config := SafeZoneConfig{
		CountRange:  [2]int{1, 1},
		RadiusRange: [2]int{0, 0},
	}

	safeZone, err := CreateSafeZone(config, &grid)

	require.NoError(t, err)
	require.Len(t, safeZone.Cells, 1)
	require.Equal(t, safeZone.Position, safeZone.Cells[0])
	require.Equal(t, pf.CellSafeZone, grid.GetCell(safeZone.Position))
}

func TestCreateSafeZone_DoesNotOverwriteNonOpenCells(t *testing.T) {
	grid := pf.NewGrid(5, 5, pf.CellOpen)
	grid.UpdateCell(pf.Position{X: 2, Y: 2}, pf.CellObstacle)
	config := SafeZoneConfig{
		CountRange:  [2]int{1, 1},
		RadiusRange: [2]int{2, 2},
	}

	safeZone, err := CreateSafeZone(config, &grid)

	require.NoError(t, err)
	require.Equal(t, pf.CellObstacle, grid.GetCell(pf.Position{X: 2, Y: 2}),
		"obstacle cell must not be overwritten by safe zone")
	for _, cell := range safeZone.Cells {
		require.NotEqual(t, pf.Position{X: 2, Y: 2}, cell,
			"safe zone should not claim obstacle cell")
	}
}

func TestCreateSafeZone_ReturnsErrorWhenNoOpenCells(t *testing.T) {
	grid := pf.NewGrid(2, 2, pf.CellObstacle)
	config := SafeZoneConfig{
		CountRange:  [2]int{1, 1},
		RadiusRange: [2]int{1, 1},
	}

	_, err := CreateSafeZone(config, &grid)

	require.Error(t, err)
	require.Contains(t, err.Error(), "no open cells")
}

func TestCreateSafeZone_RadiusMarksMultipleCells(t *testing.T) {
	grid := pf.NewGrid(5, 5, pf.CellOpen)
	config := SafeZoneConfig{
		CountRange:  [2]int{1, 1},
		RadiusRange: [2]int{2, 2},
	}

	safeZone, err := CreateSafeZone(config, &grid)

	require.NoError(t, err)
	require.Equal(t, 2, safeZone.Radius)
	// A radius-2 square from origin would cover up to (2r+1)^2 = 25 cells,
	// but the origin position is random, so we just check at least some cells are marked
	require.Greater(t, len(safeZone.Cells), 1)
}
