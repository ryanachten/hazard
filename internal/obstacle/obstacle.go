// Package obstacle defines how obstacles behave in the simulation
package obstacle

import (
	pf "hazard/internal/pathfinding"
	r "hazard/internal/ranging"
	"log/slog"

	"github.com/google/uuid"
)

// Config configures simulation obstacles
type Config struct {
	CountRange r.Range
	SizeRange  r.PositiveRange
}

// Obstacle like a building which must be avoided but can be destroyed
type Obstacle struct {
	ID    uuid.UUID
	Cells []pf.Position
}

// CreateObstacles randomly on a grid
func CreateObstacles(config Config, grid *pf.Grid) []Obstacle {
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
		var cells []pf.Position
		for dx := -width; dx <= width; dx++ {
			for dy := -height; dy <= height; dy++ {
				pos := pf.Position{X: origin.X + dx, Y: origin.Y + dy}
				if grid.InBounds(pos) && grid.GetCell(pos) == pf.CellOpen {
					grid.UpdateCell(pos, pf.CellObstacle)
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
	cells := make([]pf.Position, len(o.Cells))
	copy(cells, o.Cells)

	return Obstacle{
		ID:    o.ID,
		Cells: cells,
	}
}
