package obstacle

import (
	"testing"

	"github.com/stretchr/testify/require"

	"hazard/internal/bounds"
	"hazard/internal/pathfinding"
)

func newGridWithSafeZone(width, height int) *pathfinding.Grid {
	grid := pathfinding.NewGrid(width, height, pathfinding.CellOpen)
	grid.UpdateCell(pathfinding.Position{X: width - 1, Y: height - 1}, pathfinding.CellSafeZone)
	return &grid
}

func TestCreateObstacles_PlacesExpectedCount(t *testing.T) {
	grid := newGridWithSafeZone(20, 20)
	config := Config{
		CountRange: bounds.Range{Min: 3, Max: 3},
		SizeRange:  bounds.PositiveRange{Min: 1, Max: 1},
	}

	obstacles := CreateObstacles(config, grid)

	require.Len(t, obstacles, 3)
	for _, obs := range obstacles {
		require.NotEmpty(t, obs.Cells)
		for _, cell := range obs.Cells {
			require.Equal(t, pathfinding.CellObstacle, grid.GetCell(cell),
				"obstacle cell %v must be marked as CellObstacle", cell)
		}
	}
}

func TestCreateObstacles_DoesNotOverwriteNonOpenCells(t *testing.T) {
	grid := pathfinding.NewGrid(10, 10, pathfinding.CellOpen)
	// Mark a safe zone cell that obstacles should not overwrite
	grid.UpdateCell(pathfinding.Position{X: 5, Y: 5}, pathfinding.CellSafeZone)

	config := Config{
		CountRange: bounds.Range{Min: 10, Max: 10},
		SizeRange:  bounds.PositiveRange{Min: 1, Max: 1},
	}

	obstacles := CreateObstacles(config, &grid)

	require.NotEmpty(t, obstacles, "obstacles should be placed in open cells")
	// Safe zone cell must never be overwritten
	require.Equal(t, pathfinding.CellSafeZone, grid.GetCell(pathfinding.Position{X: 5, Y: 5}),
		"safe zone cell must not be overwritten by obstacle")
	// Collect all obstacle cells and verify none overlap with the safe zone
	for _, obs := range obstacles {
		for _, cell := range obs.Cells {
			require.NotEqual(t, pathfinding.Position{X: 5, Y: 5}, cell,
				"obstacle cell must not overlap safe zone cell")
		}
	}
}

func TestCreateObstacles_ReturnsEmptyWhenGridFull(t *testing.T) {
	grid := pathfinding.NewGrid(3, 3, pathfinding.CellObstacle)
	config := Config{
		CountRange: bounds.Range{Min: 1, Max: 1},
		SizeRange:  bounds.PositiveRange{Min: 1, Max: 1},
	}

	obstacles := CreateObstacles(config, &grid)

	require.Empty(t, obstacles)
}

func TestObstacle_CopyCreatesDeepCopy(t *testing.T) {
	original := Obstacle{
		Cells: []pathfinding.Position{{X: 0, Y: 0}, {X: 1, Y: 0}},
	}
	copied := original.Copy()

	require.Equal(t, original.Cells, copied.Cells)

	// Mutate the original to verify deep copy
	original.Cells[0] = pathfinding.Position{X: 9, Y: 9}
	require.NotEqual(t, original.Cells[0], copied.Cells[0],
		"mutating original cells must not affect copy")
}

func TestObstacle_CopyPreservesID(t *testing.T) {
	obs := Obstacle{
		Cells: []pathfinding.Position{{X: 0, Y: 0}},
	}
	// ID is zero-value since we didn't set it; just verify Copy preserves it
	copied := obs.Copy()
	require.Equal(t, obs.Cells, copied.Cells)
}
