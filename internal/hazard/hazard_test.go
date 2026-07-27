package hazard

import (
	"hazard/internal/bounds"
	"hazard/internal/pathfinding"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCreateHazard(t *testing.T) {
	grid := pathfinding.NewGrid(10, 10, pathfinding.CellOpen)
	config := Config{
		DurationRange: bounds.PositiveRange{Min: 5, Max: 10},
		Probability:   0.5,
		Count:         0,
	}

	hazard, err := Create(config, &grid)
	require.NoError(t, err)
	require.GreaterOrEqual(t, hazard.Duration, 5)
	require.LessOrEqual(t, hazard.Duration, 10)
	require.Contains(t, []Type{FireHazard, FloodHazard, LavaHazard}, hazard.Type)
	require.True(t, grid.InBounds(hazard.Origin))
	require.Equal(t, pathfinding.CellHazard, grid.GetCell(hazard.Origin))
	require.NotEqual(t, uuid.Nil, hazard.ID)
}

func TestCreateHazard_NoOpenCells(t *testing.T) {
	grid := pathfinding.NewGrid(2, 2, pathfinding.CellObstacle)
	config := Config{
		DurationRange: bounds.PositiveRange{Min: 5, Max: 5},
	}

	_, err := Create(config, &grid)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no open cells available")
}

func TestRandomType(t *testing.T) {
	grid := pathfinding.NewGrid(20, 20, pathfinding.CellOpen)
	config := Config{
		DurationRange: bounds.PositiveRange{Min: 5, Max: 10},
	}

	seen := make(map[Type]bool)
	for range 100 {
		hazard, err := Create(config, &grid)
		require.NoError(t, err)
		require.Contains(t, []Type{FireHazard, FloodHazard, LavaHazard}, hazard.Type)
		seen[hazard.Type] = true
	}
	require.GreaterOrEqual(t, len(seen), 2)
}

func TestHazard_CellsBlockPathfinding(t *testing.T) {
	grid := pathfinding.NewGrid(5, 5, pathfinding.CellOpen)
	hazard := Hazard{
		ID:            uuid.New(),
		Type:          FireHazard,
		Origin:        pathfinding.Position{X: 2, Y: 2},
		CurrentRadius: 0,
	}

	grid.UpdateCell(pathfinding.Position{X: 2, Y: 2}, pathfinding.CellHazard)
	hazard.Expand(&grid)

	require.Equal(t, pathfinding.CellHazard, grid.GetCell(pathfinding.Position{X: 2, Y: 2}))
	require.Equal(t, pathfinding.CellHazard, grid.GetCell(pathfinding.Position{X: 2, Y: 1}))
	require.Equal(t, pathfinding.CellHazard, grid.GetCell(pathfinding.Position{X: 2, Y: 3}))
	require.Equal(t, pathfinding.CellHazard, grid.GetCell(pathfinding.Position{X: 1, Y: 2}))
	require.Equal(t, pathfinding.CellHazard, grid.GetCell(pathfinding.Position{X: 3, Y: 2}))

	path, err := pathfinding.FindPath(&grid, pathfinding.Position{X: 0, Y: 0}, pathfinding.Position{X: 4, Y: 4})
	require.NoError(t, err)
	for _, pos := range path {
		require.NotEqual(t, pathfinding.CellHazard, grid.GetCell(pos), "path must not enter hazard cell at %v", pos)
	}

	for y := range grid.Height {
		for x := range grid.Width {
			grid.Cells[y][x] = pathfinding.CellHazard
		}
	}
	_, err = pathfinding.FindPath(&grid, pathfinding.Position{X: 0, Y: 0}, pathfinding.Position{X: 4, Y: 4})
	require.ErrorIs(t, err, pathfinding.ErrDestinationUnreachable)
}
