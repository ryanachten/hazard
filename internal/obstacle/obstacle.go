// Package obstacle defines how obstacles behave in the simulation
package obstacle

import (
	"log/slog"

	"github.com/google/uuid"

	"hazard/internal/bounds"
	"hazard/internal/pathfinding"
)

// Config configures simulation obstacles
type Config struct {
	CountRange bounds.Range
	SizeRange  bounds.PositiveRange
}

// Obstacle like a building which must be avoided but can be destroyed
type Obstacle struct {
	ID    uuid.UUID
	Cells []pathfinding.Position
}

// CreateObstacles randomly on a grid
func CreateObstacles(config Config, grid *pathfinding.Grid) []Obstacle {
	obstacleCount := config.CountRange.Random()
	var obstacles []Obstacle

	for range obstacleCount {
		height := config.SizeRange.Random()
		width := config.SizeRange.Random()
		origin, err := grid.GetRandomOpenPosition(width, height, grid.Width-1-width, grid.Height-1-height)

		if err != nil {
			slog.Warn("error creating obstacle", "err", err)
			continue
		}

		// Mark all open cells as part of the obstacle
		var cells []pathfinding.Position
		for dx := -width; dx <= width; dx++ {
			for dy := -height; dy <= height; dy++ {
				pos := pathfinding.Position{X: origin.X + dx, Y: origin.Y + dy}
				if grid.InBounds(pos) && grid.GetCell(pos) == pathfinding.CellOpen {
					grid.UpdateCell(pos, pathfinding.CellObstacle)
					cells = append(cells, pos)
				}
			}
		}

		obstacles = append(obstacles, Obstacle{
			ID:    uuid.New(),
			Cells: cells,
		})
	}

	return obstacles
}

// Copy creates a copy of an obstacle
func (o *Obstacle) Copy() Obstacle {
	cells := make([]pathfinding.Position, len(o.Cells))
	copy(cells, o.Cells)

	return Obstacle{
		ID:    o.ID,
		Cells: cells,
	}
}
