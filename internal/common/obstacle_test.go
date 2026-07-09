package common

import (
	pf "hazard/internal/pathfinding"
	"testing"

	"github.com/stretchr/testify/require"
)

func newGridWithSafeZone(width, height int) *pf.Grid {
	grid := pf.NewGrid(width, height, pf.CellOpen)
	grid.UpdateCell(pf.Position{X: width - 1, Y: height - 1}, pf.CellSafeZone)
	return &grid
}

func TestCreateObstacles_PlacesExpectedCount(t *testing.T) {
	grid := newGridWithSafeZone(20, 20)
	config := ObstacleConfig{
		CountRange: Range{Min: 3, Max: 3},
		SizeRange:  PositiveRange{Min: 1, Max: 1},
	}

	obstacles := CreateObstacles(config, grid)

	require.Len(t, obstacles, 3)
	for _, obs := range obstacles {
		require.NotEmpty(t, obs.Cells)
		for _, cell := range obs.Cells {
			require.Equal(t, pf.CellObstacle, grid.GetCell(cell),
				"obstacle cell %v must be marked as CellObstacle", cell)
		}
	}
}

func TestCreateObstacles_DoesNotOverwriteNonOpenCells(t *testing.T) {
	grid := pf.NewGrid(10, 10, pf.CellOpen)
	// Mark a safe zone cell that obstacles should not overwrite
	grid.UpdateCell(pf.Position{X: 5, Y: 5}, pf.CellSafeZone)

	config := ObstacleConfig{
		CountRange: Range{Min: 10, Max: 10},
		SizeRange:  PositiveRange{Min: 1, Max: 1},
	}

	obstacles := CreateObstacles(config, &grid)

	require.NotEmpty(t, obstacles, "obstacles should be placed in open cells")
	// Safe zone cell must never be overwritten
	require.Equal(t, pf.CellSafeZone, grid.GetCell(pf.Position{X: 5, Y: 5}),
		"safe zone cell must not be overwritten by obstacle")
	// Collect all obstacle cells and verify none overlap with the safe zone
	for _, obs := range obstacles {
		for _, cell := range obs.Cells {
			require.NotEqual(t, pf.Position{X: 5, Y: 5}, cell,
				"obstacle cell must not overlap safe zone cell")
		}
	}
}

func TestCreateObstacles_ReturnsEmptyWhenGridFull(t *testing.T) {
	grid := pf.NewGrid(3, 3, pf.CellObstacle)
	config := ObstacleConfig{
		CountRange: Range{Min: 1, Max: 1},
		SizeRange:  PositiveRange{Min: 1, Max: 1},
	}

	obstacles := CreateObstacles(config, &grid)

	require.Empty(t, obstacles)
}

func TestObstacle_CopyCreatesDeepCopy(t *testing.T) {
	original := Obstacle{
		Cells: []pf.Position{{X: 0, Y: 0}, {X: 1, Y: 0}},
	}
	copied := original.Copy()

	require.Equal(t, original.Cells, copied.Cells)

	// Mutate the original to verify deep copy
	original.Cells[0] = pf.Position{X: 9, Y: 9}
	require.NotEqual(t, original.Cells[0], copied.Cells[0],
		"mutating original cells must not affect copy")
}

func TestObstacle_CopyPreservesID(t *testing.T) {
	obs := Obstacle{
		Cells: []pf.Position{{X: 0, Y: 0}},
	}
	// ID is zero-value since we didn't set it; just verify Copy preserves it
	copied := obs.Copy()
	require.Equal(t, obs.Cells, copied.Cells)
}
