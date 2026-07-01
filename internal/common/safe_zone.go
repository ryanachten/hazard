package common

import (
	pf "hazard/internal/pathfinding"

	"github.com/google/uuid"
)

// SafeZone represents an area safe from hazards for citizens to navigate to
type SafeZone struct {
	ID       uuid.UUID
	Position pf.Position
	Radius   int
	Cells    []pf.Position
}

// CreateSafeZone creates a safe zone at a random open position with a random radius
func CreateSafeZone(config SafeZoneConfig, grid *pf.Grid) (SafeZone, error) {
	radius := RandIntInRange(config.RadiusRange)

	origin, err := grid.GetRandomOpenPosition(radius, radius, grid.Width-1-radius, grid.Height-1-radius)
	if err != nil {
		return SafeZone{}, err
	}

	// Mark all open cells as part of the safe zone
	var cells []pf.Position
	for dx := -radius; dx <= radius; dx++ {
		for dy := -radius; dy <= radius; dy++ {
			pos := pf.Position{X: origin.X + dx, Y: origin.Y + dy}
			if grid.InBounds(pos) && grid.GetCell(pos) == pf.CellOpen {
				grid.UpdateCell(pos, pf.CellSafeZone)
				cells = append(cells, pos)
			}
		}
	}

	return SafeZone{
		ID:       uuid.New(),
		Position: origin,
		Radius:   radius,
		Cells:    cells,
	}, nil
}

// Copy returns a deep copy of the SafeZone
func (s *SafeZone) Copy() SafeZone {
	cells := make([]pf.Position, len(s.Cells))
	copy(cells, s.Cells)
	return SafeZone{
		ID:       s.ID,
		Position: s.Position,
		Radius:   s.Radius,
		Cells:    cells,
	}
}
