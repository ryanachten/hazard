package pathfinding

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDijkstra_FindPath(t *testing.T) {
	testFindPath(t, &Dijkstra{})
}

func TestAStar_FindPath(t *testing.T) {
	testFindPath(t, &AStar{})
}

func testFindPath(t *testing.T, algo Pathfinder) {
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
			path, err := algo.FindPath(&test.grid, test.from, test.to)

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

func TestNewGrid_PanicsOnInvalidDimensions(t *testing.T) {
	require.Panics(t, func() { NewGrid(0, 1, CellOpen) })
	require.Panics(t, func() { NewGrid(1, 0, CellOpen) })
	require.Panics(t, func() { NewGrid(-1, 1, CellOpen) })
}

func TestGrid_GetRandomOpenPosition(t *testing.T) {
	t.Run("open grid returns valid position", func(t *testing.T) {
		grid := NewGrid(5, 5, CellOpen)
		pos, err := grid.GetRandomOpenPosition()
		require.NoError(t, err)
		require.True(t, grid.InBounds(pos))
		require.Equal(t, CellOpen, grid.GetCell(pos))
	})

	t.Run("fully obstructed grid returns error", func(t *testing.T) {
		grid := NewGrid(3, 3, CellObstacle)
		_, err := grid.GetRandomOpenPosition()
		require.Error(t, err)
		require.Contains(t, err.Error(), "no open cells available")
	})

	t.Run("single open cell is found", func(t *testing.T) {
		grid := NewGrid(3, 3, CellObstacle)
		grid.Cells[1][1] = CellOpen
		pos, err := grid.GetRandomOpenPosition()
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

func TestAStar_Name(t *testing.T) {
	require.Equal(t, "a*", (&AStar{}).Name())
}

func TestDijkstra_Name(t *testing.T) {
	require.Equal(t, "dijkstra", (&Dijkstra{}).Name())
}
