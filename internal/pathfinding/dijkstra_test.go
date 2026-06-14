package pathfinding

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDijkstra_FindPath(t *testing.T) {
	tests := []struct {
		name        string
		grid        grid
		from        Position
		to          Position
		expectedLen int
		expectedErr error
	}{
		{
			name:        "adjacent cells",
			grid:        grid{Width: 3, Height: 3, Cells: makeGrid(3, 3, cellOpen)},
			from:        Position{X: 0, Y: 0},
			to:          Position{X: 1, Y: 0},
			expectedLen: 2,
		},
		{
			name:        "blocked by obstacle",
			grid:        grid{Width: 2, Height: 2, Cells: makeGrid(2, 2, cellObstacle)},
			from:        Position{X: 0, Y: 0},
			to:          Position{X: 1, Y: 0},
			expectedErr: ErrDestinationUnreachable,
		},
		{
			name:        "to and from are the same position",
			grid:        grid{Width: 3, Height: 3, Cells: makeGrid(3, 3, cellOpen)},
			from:        Position{X: 0, Y: 0},
			to:          Position{X: 0, Y: 0},
			expectedLen: 1,
		}}

	algo := &Dijkstra{}

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

func makeGrid(width, height int, variant cellType) [][]cellType {
	cells := make([][]cellType, height)
	for y := range height {
		cells[y] = make([]cellType, height)

		for x := range width {
			cells[y][x] = variant
		}
	}

	return cells
}
