package pathfinding

import (
	"container/heap"
	"math"
	"slices"
)

// AStar implementation of the pathfinding algorithm
type AStar struct{}

// Name returns the given name of a pathfinding algorithm
func (a *AStar) Name() string {
	return "a*"
}

// FindPath finds shortest path from a starting point to a given destination
func (a *AStar) FindPath(grid Grid, from, to Position) ([]Position, error) {
	if !grid.InBounds(from) || !grid.InBounds(to) {
		return nil, ErrPositionOutOfBounds
	}

	// If from and to are the same node, then short-circuit
	if from == to {
		return []Position{from}, nil
	}

	pq := make(PriorityQueue, 0)
	heap.Init(&pq)

	// Initialize distances to max value
	distances := make([][]int, grid.Height)
	for y := range grid.Height {
		distances[y] = make([]int, grid.Width)
		for x := range grid.Width {
			distances[y][x] = math.MaxInt
		}
	}

	// Initialize traversal path to record paths to cells
	traversalPath := make([][]*Position, grid.Height)
	for y := range grid.Height {
		traversalPath[y] = make([]*Position, grid.Width)
	}

	// Set source distance to 0 and add to priority queue
	distances[from.Y][from.X] = 0
	heap.Push(&pq, &Item{
		value:    from,
		priority: 0,
	})

	// Process queue until we have determined distances
	for pq.Len() > 0 {
		curItem := heap.Pop(&pq).(*Item)
		curPos := curItem.value

		if curPos == to {
			break
		}

		// Explore neighbouring elements of index
		for _, direction := range Directions {
			y := curPos.Y + direction.Y
			x := curPos.X + direction.X

			// Bounds check
			if y < 0 || y > grid.Height-1 || x < 0 || x > grid.Width-1 {
				continue
			}

			// Skip obstacles
			if grid.Cells[y][x] == CellObstacle {
				continue
			}

			cost := 1 // how much each step costs

			// If we've found a shorter distance to a neighbour...
			if distances[curPos.Y][curPos.X]+cost < distances[y][x] {
				// Update the shortest distance
				distances[y][x] = distances[curPos.Y][curPos.X] + cost

				// Update traversal path to reference current node
				traversalPath[y][x] = &curPos

				// Add neighbouring node to queue
				heap.Push(&pq, &Item{
					value: Position{
						X: x,
						Y: y,
					},
					priority: distances[y][x] + heuristic(Position{X: x, Y: y}, to),
				})
			}
		}
	}

	curNode := traversalPath[to.Y][to.X]
	if curNode == nil {
		return nil, ErrDestinationUnreachable
	}

	result := make([]Position, 0)
	result = append(result, to)

	for curNode != nil {
		result = append(result, *curNode)
		curNode = traversalPath[curNode.Y][curNode.X]
	}

	slices.Reverse(result)

	return result, nil
}

// Using Manhattan heuristic since we can only move left, right, up and down
func heuristic(position, destination Position) int {
	cost := math.Abs(float64(position.Y-destination.Y)) + math.Abs(float64(position.X-destination.X))
	return int(cost)
}
