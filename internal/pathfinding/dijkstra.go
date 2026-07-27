package pathfinding

import (
	"container/heap"
	"math"
	"slices"
)

// GoalPredicate defines when a goal has been reached
type GoalPredicate func(pos Position) bool

// FindPathToGoal finds shortest path from a starting point to any location matching a given predicate
func FindPathToGoal(grid *Grid, from Position, isGoal GoalPredicate) ([]Position, error) {
	if !grid.InBounds(from) {
		return nil, ErrPositionOutOfBounds
	}

	// If the source is already at goal, then short circuit
	if isGoal(from) {
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

	// Process queue until we have fulfilled the goal
	var goalPosition Position
	found := false
	for pq.Len() > 0 {
		curItem, _ := heap.Pop(&pq).(*Item)
		curPos := curItem.value

		if isGoal(curPos) {
			goalPosition = curPos
			found = true
			break
		}

		// If it's not the shortest distance for this node, skip it
		if curItem.priority > distances[curPos.Y][curPos.X] {
			continue
		}

		// Explore neighbouring elements of index
		for _, direction := range Directions {
			y := curPos.Y + direction.Y
			x := curPos.X + direction.X

			// Bounds check
			if y < 0 || y > grid.Height-1 || x < 0 || x > grid.Width-1 {
				continue
			}

			// Skip avoidable cells
			if AvoidableCellType[grid.Cells[y][x]] {
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
					priority: distances[y][x], // cost = distance
				})
			}
		}
	}

	if !found {
		return nil, ErrDestinationUnreachable
	}

	curNode := traversalPath[goalPosition.Y][goalPosition.X]

	result := make([]Position, 0)
	result = append(result, goalPosition)

	for curNode != nil {
		result = append(result, *curNode)
		curNode = traversalPath[curNode.Y][curNode.X]
	}

	slices.Reverse(result)

	return result, nil
}
