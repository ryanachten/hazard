package pathfinding

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAStar_FindPath(t *testing.T) {
	tests := []struct {
		name        string
		grid        Grid
		from        Position
		to          Position
		expectedLen int
		expectedErr error
	}{
		{
			name:        "adjacent cells",
			grid:        NewGrid(3, 3, CellOpen),
			from:        Position{X: 0, Y: 0},
			to:          Position{X: 1, Y: 0},
			expectedLen: 2,
		},
		{
			name:        "non-trivial navigation",
			grid:        NewGrid(6, 6, CellOpen),
			from:        Position{X: 0, Y: 0},
			to:          Position{X: 5, Y: 5},
			expectedLen: 11,
		},
		{
			name:        "blocked by obstacle",
			grid:        NewGrid(2, 2, CellObstacle),
			from:        Position{X: 0, Y: 0},
			to:          Position{X: 1, Y: 0},
			expectedErr: ErrDestinationUnreachable,
		},
		{
			name:        "to and from are the same position",
			grid:        NewGrid(3, 3, CellOpen),
			from:        Position{X: 0, Y: 0},
			to:          Position{X: 0, Y: 0},
			expectedLen: 1,
		},
		{
			name:        "from is out of bounds (negative)",
			grid:        NewGrid(3, 3, CellOpen),
			from:        Position{X: -1, Y: 0},
			to:          Position{X: 1, Y: 0},
			expectedErr: ErrPositionOutOfBounds,
		},
		{
			name:        "from is out of bounds (exceeds width)",
			grid:        NewGrid(3, 3, CellOpen),
			from:        Position{X: 5, Y: 0},
			to:          Position{X: 1, Y: 0},
			expectedErr: ErrPositionOutOfBounds,
		},
		{
			name:        "to is out of bounds (negative)",
			grid:        NewGrid(3, 3, CellOpen),
			from:        Position{X: 0, Y: 0},
			to:          Position{X: -1, Y: -1},
			expectedErr: ErrPositionOutOfBounds,
		},
		{
			name:        "to is out of bounds (exceeds height)",
			grid:        NewGrid(3, 3, CellOpen),
			from:        Position{X: 0, Y: 0},
			to:          Position{X: 0, Y: 5},
			expectedErr: ErrPositionOutOfBounds,
		}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path, err := FindPath(&test.grid, test.from, test.to)

			if test.expectedErr != nil {
				require.ErrorIs(t, err, test.expectedErr)
				return
			}
			require.NoError(t, err)

			require.Len(t, path, test.expectedLen, "path length mismatch")
			require.Equal(t, test.from, path[0], "path should start at from")
			require.Equal(t, test.to, path[len(path)-1], "path should end at to")
		})
	}
}

func TestDijkstra_FindPathToGoal(t *testing.T) {
	t.Run("adjacent cells", func(t *testing.T) {
		grid := NewGrid(3, 3, CellOpen)
		grid.Cells[0][1] = CellSafeZone
		path, err := FindPathToGoal(&grid, Position{X: 0, Y: 0}, func(pos Position) bool {
			return grid.GetCell(pos) == CellSafeZone
		})
		require.NoError(t, err)
		require.Len(t, path, 2)
		require.Equal(t, Position{X: 0, Y: 0}, path[0])
		require.Equal(t, Position{X: 1, Y: 0}, path[len(path)-1])
	})

	t.Run("non-trivial navigation", func(t *testing.T) {
		grid := NewGrid(6, 6, CellOpen)
		grid.Cells[5][5] = CellSafeZone
		path, err := FindPathToGoal(&grid, Position{X: 0, Y: 0}, func(pos Position) bool {
			return grid.GetCell(pos) == CellSafeZone
		})
		require.NoError(t, err)
		require.Len(t, path, 11)
		require.Equal(t, Position{X: 0, Y: 0}, path[0])
		require.Equal(t, Position{X: 5, Y: 5}, path[len(path)-1])
	})

	t.Run("from is already at goal", func(t *testing.T) {
		grid := NewGrid(3, 3, CellOpen)
		grid.Cells[0][0] = CellSafeZone
		path, err := FindPathToGoal(&grid, Position{X: 0, Y: 0}, func(pos Position) bool {
			return grid.GetCell(pos) == CellSafeZone
		})
		require.NoError(t, err)
		require.Len(t, path, 1)
		require.Equal(t, Position{X: 0, Y: 0}, path[0])
	})

	t.Run("from is out of bounds (negative)", func(t *testing.T) {
		grid := NewGrid(3, 3, CellOpen)
		_, err := FindPathToGoal(&grid, Position{X: -1, Y: 0}, func(pos Position) bool {
			return grid.GetCell(pos) == CellSafeZone
		})
		require.ErrorIs(t, err, ErrPositionOutOfBounds)
	})

	t.Run("from is out of bounds (exceeds width)", func(t *testing.T) {
		grid := NewGrid(3, 3, CellOpen)
		_, err := FindPathToGoal(&grid, Position{X: 5, Y: 0}, func(pos Position) bool {
			return grid.GetCell(pos) == CellSafeZone
		})
		require.ErrorIs(t, err, ErrPositionOutOfBounds)
	})

	t.Run("goal unreachable (no safe zone)", func(t *testing.T) {
		grid := NewGrid(2, 2, CellOpen)
		_, err := FindPathToGoal(&grid, Position{X: 0, Y: 0}, func(pos Position) bool {
			return grid.GetCell(pos) == CellSafeZone
		})
		require.ErrorIs(t, err, ErrDestinationUnreachable)
	})

	t.Run("blocked by obstacle with safe zone present", func(t *testing.T) {
		grid := NewGrid(3, 3, CellOpen)
		grid.Cells[0][1] = CellObstacle
		grid.Cells[1][0] = CellObstacle
		grid.Cells[2][2] = CellSafeZone
		_, err := FindPathToGoal(&grid, Position{X: 0, Y: 0}, func(pos Position) bool {
			return grid.GetCell(pos) == CellSafeZone
		})
		require.ErrorIs(t, err, ErrDestinationUnreachable)
	})
}

func TestGrid_GetRandomOpenPosition(t *testing.T) {
	t.Run("open grid returns valid position", func(t *testing.T) {
		grid := NewGrid(5, 5, CellOpen)
		pos, err := grid.GetRandomOpenPosition(0, 0, grid.Width-1, grid.Height-1)
		require.NoError(t, err)
		require.True(t, grid.InBounds(pos))
		require.Equal(t, CellOpen, grid.GetCell(pos))
	})

	t.Run("fully obstructed grid returns error", func(t *testing.T) {
		grid := NewGrid(3, 3, CellObstacle)
		_, err := grid.GetRandomOpenPosition(0, 0, grid.Width-1, grid.Height-1)
		require.Error(t, err)
		require.Contains(t, err.Error(), "no open cells available")
	})

	t.Run("single open cell is found", func(t *testing.T) {
		grid := NewGrid(3, 3, CellObstacle)
		grid.Cells[1][1] = CellOpen
		pos, err := grid.GetRandomOpenPosition(0, 0, grid.Width-1, grid.Height-1)
		require.NoError(t, err)
		require.Equal(t, Position{X: 1, Y: 1}, pos)
	})
}

func TestGrid_GetCellAndUpdateCell(t *testing.T) {
	grid := NewGrid(3, 3, CellOpen)
	pos := Position{X: 1, Y: 1}

	require.Equal(t, CellOpen, grid.GetCell(pos))

	grid.UpdateCell(pos, CellObstacle)
	require.Equal(t, CellObstacle, grid.GetCell(pos))

	grid.UpdateCell(pos, CellHazard)
	require.Equal(t, CellHazard, grid.GetCell(pos))

	grid.UpdateCell(pos, CellOpen)
	require.Equal(t, CellOpen, grid.GetCell(pos))
}

func TestFindPath_AvoidsCellCitizen(t *testing.T) {
	grid := NewGrid(3, 3, CellOpen)
	grid.Cells[1][1] = CellCitizen

	path, err := FindPath(&grid, Position{X: 0, Y: 0}, Position{X: 2, Y: 2})
	require.NoError(t, err)
	require.NotContains(t, path, Position{X: 1, Y: 1},
		"path must avoid CellCitizen cell")
}

func TestFindPath_AvoidsCellDeadCitizen(t *testing.T) {
	// CellDeadCitizen is NOT in AvoidableCellType, so pathfinding should pass through it.
	// Bottleneck grid: only route from (0,0) to (2,2) passes through (1,1).
	//
	//	O X O //nolint:dupword // ASCII art grid diagram
	//	O . O //nolint:dupword // ASCII art grid diagram
	//	O X O //nolint:dupword // ASCII art grid diagram
	grid := NewGrid(3, 3, CellOpen)
	grid.Cells[0][1] = CellObstacle
	grid.Cells[1][1] = CellDeadCitizen
	grid.Cells[2][1] = CellObstacle

	path, err := FindPath(&grid, Position{X: 0, Y: 0}, Position{X: 2, Y: 2})
	require.NoError(t, err)
	require.Contains(t, path, Position{X: 1, Y: 1},
		"path must pass through CellDeadCitizen since it is not avoidable and is the only route")
}

func TestFindPathToGoal_AvoidsCellCitizen(t *testing.T) {
	grid := NewGrid(3, 3, CellOpen)
	grid.Cells[1][1] = CellCitizen
	grid.Cells[2][2] = CellSafeZone

	path, err := FindPathToGoal(&grid, Position{X: 0, Y: 0}, func(pos Position) bool {
		return grid.GetCell(pos) == CellSafeZone
	})
	require.NoError(t, err)
	require.NotContains(t, path, Position{X: 1, Y: 1},
		"path to safe zone must avoid CellCitizen cell")
}

func TestGetRandomOpenPosition_SkipsCellCitizen(t *testing.T) {
	grid := NewGrid(3, 3, CellCitizen)
	grid.Cells[1][1] = CellOpen

	pos, err := grid.GetRandomOpenPosition(0, 0, grid.Width-1, grid.Height-1)
	require.NoError(t, err)
	require.Equal(t, Position{X: 1, Y: 1}, pos,
		"GetRandomOpenPosition should treat CellCitizen as blocked and pick the only open cell")
}
