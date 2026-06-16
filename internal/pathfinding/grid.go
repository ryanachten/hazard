package pathfinding

import "math/rand"

// Position on a grid
type Position struct {
	X int
	Y int
}

type CellType int

const (
	CellOpen     CellType = iota // Passable
	CellObstacle                 // Blocked
)

// Grid (2D) for positioning simulation entities
type Grid struct {
	Width  int
	Height int
	Cells  [][]CellType
}

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

func (g *Grid) GetRandomPosition() Position {
	x := rand.Intn(g.Width)
	y := rand.Intn(g.Height)

	return Position{
		X: x,
		Y: y,
	}
}
