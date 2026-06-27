package engine

import (
	pf "hazard/internal/pathfinding"

	"github.com/google/uuid"
)

// SafeZone represents an area safe from hazards for citizens to navigate to
type SafeZone struct {
	ID       uuid.UUID
	Position pf.Position
	Radius   int
}

func createSafeZone(config SafeZoneConfig, grid *pf.Grid) (SafeZone, error) {
	radius := randIntInRange(config.RadiusRange)

	origin, err := grid.GetRandomOpenPosition()
	if err != nil {
		return SafeZone{}, err
	}

	// Mark all open cells as part of the safe zone
	for dx := -radius; dx <= radius; dx++ {
		for dy := -radius; dy <= radius; dy++ {
			pos := pf.Position{X: origin.X + dx, Y: origin.Y + dy}
			if grid.InBounds(pos) && grid.GetCell(pos) == pf.CellOpen {
				grid.UpdateCell(pos, pf.CellSafeZone)
			}
		}
	}

	return SafeZone{
		Position: origin,
		Radius:   radius,
	}, nil
}
