package pathfinding

import (
	"errors"
	"math/rand"
)

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
	// CellHazard represents a cell occupied by a hazard
	CellHazard
)

// Grid (2D) for positioning simulation entities
type Grid struct {
	Width  int
	Height int
	Cells  [][]CellType
}

// NewGrid creates a new Grid with the given dimensions, filled with the specified cell type.
func NewGrid(width, height int, variant CellType) Grid {
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

// GetRandomOpenPosition returns a random unoccupied position within the grid bounds
func (g *Grid) GetRandomOpenPosition() (Position, error) {
	// Conduct n number of attempts to randomly find an open cell
	for range 10 {
		p := Position{X: rand.Intn(g.Width), Y: rand.Intn(g.Height)}
		if g.GetCell(p) == CellOpen {
			return p, nil
		}
	}
	// If random assignment fails, we scan for an open cell
	for y := range g.Height {
		for x := range g.Width {
			p := Position{X: x, Y: y}
			if g.GetCell(p) == CellOpen {
				return p, nil
			}
		}
	}

	return Position{}, errors.New("no open cells available")
}

// InBounds returns true if the given position is within the grid bounds
func (g *Grid) InBounds(p Position) bool {
	return p.X >= 0 && p.X < g.Width && p.Y >= 0 && p.Y < g.Height
}

// GetCell returns type at a give position
func (g *Grid) GetCell(p Position) CellType {
	return g.Cells[p.Y][p.X]
}

// UpdateCell to a given cell type
func (g *Grid) UpdateCell(p Position, cellType CellType) {
	g.Cells[p.Y][p.X] = cellType
}
