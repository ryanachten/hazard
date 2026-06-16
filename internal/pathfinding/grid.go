package pathfinding

import "math/rand"

// Position on a grid
type Position struct {
	X int
	Y int
}

// CellType represents the type of a cell on the grid
type CellType int

const (
	// CellOpen represents a passable cell
	CellOpen CellType = iota
	// CellObstacle represents a blocked cell
	CellObstacle
)

// Grid (2D) for positioning simulation entities
type Grid struct {
	Width  int
	Height int
	Cells  [][]CellType
}

// NewGrid creates a new Grid with the given dimensions, filled with the specified cell type.
// Panics if width or height are less than or equal to zero.
func NewGrid(width, height int, variant CellType) Grid {
	if width <= 0 || height <= 0 {
		panic("grid width and height must be greater than zero")
	}
	cells := make([][]CellType, height)
	for y := range height {
		cells[y] = make([]CellType, width)

		for x := range width {
			cells[y][x] = variant
		}
	}

	return Grid{
		Width:  width,
		Height: height,
		Cells:  cells,
	}
}

// GetRandomPosition returns a random position within the grid bounds
func (g *Grid) GetRandomPosition() Position {
	x := rand.Intn(g.Width)
	y := rand.Intn(g.Height)

	return Position{
		X: x,
		Y: y,
	}
}

// InBounds returns true if the given position is within the grid bounds
func (g *Grid) InBounds(p Position) bool {
	return p.X >= 0 && p.X < g.Width && p.Y >= 0 && p.Y < g.Height
}
