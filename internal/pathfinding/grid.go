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
	// CellOpen represents an unoccupied
	CellOpen CellType = iota
	// CellObstacle represents a cell occupied by a building
	CellObstacle
	// CellHazard represents a cell occupied by a hazard
	CellHazard
	// CellSafeZone represents a cell occupied by a safe zone
	CellSafeZone
	// CellCitizen represents a cell occupied by a citizen
	CellCitizen
	// CellEscapedCitizen represents cell occupied by a citizen escaped to a safe zone
	CellEscapedCitizen
	// CellDeadCitizen represents cell occupied by a citizen who has been killed
	CellDeadCitizen
)

// AvoidableCellType lookup for pathfinding to determine which cell types should be avoided
var AvoidableCellType = map[CellType]bool{
	CellOpen:           false,
	CellObstacle:       true,
	CellHazard:         true,
	CellSafeZone:       false,
	CellCitizen:        true,
	CellEscapedCitizen: true,
	CellDeadCitizen:    false,
}

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

// GetRandomOpenPosition returns a random unoccupied position within the given bounds (inclusive).
func (g *Grid) GetRandomOpenPosition(minX, minY, maxX, maxY int) (Position, error) {
	if minX > maxX || minY > maxY {
		return Position{}, errors.New("invalid bounds")
	}

	for range 10 {
		p := Position{X: rand.Intn(maxX-minX+1) + minX, Y: rand.Intn(maxY-minY+1) + minY}
		if g.GetCell(p) == CellOpen {
			return p, nil
		}
	}

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
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

// Copy grid cells into a new instance
func (g *Grid) Copy() Grid {
	cells := make([][]CellType, g.Height)
	for y := range g.Height {
		cells[y] = make([]CellType, g.Width)
		copy(cells[y], g.Cells[y])
	}
	return Grid{Width: g.Width, Height: g.Height, Cells: cells}
}
