// Package pathfinding provides grid-based pathfinding capabilities
package pathfinding

import "errors"

// Directions to navigate a grid
var Directions = []Position{
	{X: 0, Y: -1}, // north
	{X: 0, Y: 1},  // south
	{X: -1, Y: 0}, // west
	{X: 1, Y: 0},  // east
}

// ErrDestinationUnreachable returned when unable to reach destination from source node
var ErrDestinationUnreachable = errors.New("destination unreachable")

// ErrPositionOutOfBounds returned when a position is outside grid bounds
var ErrPositionOutOfBounds = errors.New("position out of bounds")
