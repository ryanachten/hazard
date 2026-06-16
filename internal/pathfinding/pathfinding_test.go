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
			path, err := algo.FindPath(test.grid, test.from, test.to)

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
