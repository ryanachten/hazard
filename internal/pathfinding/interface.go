// Package pathfinding provides grid-based pathfinding capabilities
package pathfinding

// Position on a grid
type Position struct {
	X int
	Y int
}

type cellType int

const (
	cellOpen     cellType = iota // Passable
	cellObstacle                 // Blocked
)

type grid struct {
	Width  int
	Height int
	Cells  [][]cellType
}

// Pathfinder implementation of a pathfinding algorithm
type Pathfinder interface {
	Name() string
	FindPath(grid grid, from, to Position) ([]Position, error)
}
