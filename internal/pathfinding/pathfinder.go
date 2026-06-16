// Package pathfinding provides grid-based pathfinding capabilities
package pathfinding

import "errors"

// Pathfinder implementation of a pathfinding algorithm
type Pathfinder interface {
	Name() string
	FindPath(grid Grid, from, to Position) ([]Position, error)
}

var directions = []Position{
	{X: 0, Y: -1}, // north
	{X: 0, Y: 1},  // south
	{X: -1, Y: 0}, // west
	{X: 1, Y: 0},  // east
}

// ErrDestinationUnreachable returned when unable to reach destination from source node
var ErrDestinationUnreachable = errors.New("destination unreachable")
