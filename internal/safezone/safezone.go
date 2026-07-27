// Package safezone defines safe zone entities for citizens to reach
package safezone

import (
	"hazard/internal/bounds"
	"hazard/internal/pathfinding"
	"log/slog"
	"maps"

	"github.com/google/uuid"
)

// Config configures simulation safe zones
type Config struct {
	Probability float32
	Count       int
	RadiusRange bounds.Range
}

// SafeZone represents an area safe from hazards for citizens to navigate to
type SafeZone struct {
	ID            uuid.UUID
	Position      pathfinding.Position
	Radius        int
	Cells         []pathfinding.Position
	HasCapacity   bool
	Occupants     []uuid.UUID
	occupiedCells map[pathfinding.Position]bool
}

// Create creates a safe zone at a random open position with a random radius
func Create(config Config, grid *pathfinding.Grid) (SafeZone, error) {
	radius := config.RadiusRange.Random()

	origin, err := grid.GetRandomOpenPosition(radius, radius, grid.Width-1-radius, grid.Height-1-radius)
	if err != nil {
		return SafeZone{}, err
	}

	// Mark all open cells as part of the safe zone
	var cells []pathfinding.Position
	for dx := -radius; dx <= radius; dx++ {
		for dy := -radius; dy <= radius; dy++ {
			pos := pathfinding.Position{X: origin.X + dx, Y: origin.Y + dy}
			if grid.InBounds(pos) && grid.GetCell(pos) == pathfinding.CellOpen {
				grid.UpdateCell(pos, pathfinding.CellSafeZone)
				cells = append(cells, pos)
			}
		}
	}

	id := uuid.New()

	slog.Info("safe zone created", "id", id, "origin", origin)

	return SafeZone{
		ID:            id,
		Position:      origin,
		Radius:        radius,
		Cells:         cells,
		HasCapacity:   true,
		Occupants:     []uuid.UUID{},
		occupiedCells: map[pathfinding.Position]bool{},
	}, nil
}

// AddOccupant adds a citizen to the safe zone, marking their cell as occupied.
// It first tries the citizen's arrival cell; if it's already taken, the first free cell is used.
// Returns the assigned position and whether the citizen was admitted.
func (s *SafeZone) AddOccupant(citizenID uuid.UUID, position pathfinding.Position, grid *pathfinding.Grid) (pathfinding.Position, bool) {
	if !s.HasCapacity {
		return pathfinding.Position{}, false
	}

	if s.occupiedCells == nil {
		s.occupiedCells = make(map[pathfinding.Position]bool)
	}

	assignPos := position
	if s.occupiedCells[assignPos] {
		found := false
		for _, cell := range s.Cells {
			if !s.occupiedCells[cell] {
				assignPos = cell
				found = true
				break
			}
		}
		if !found {
			return pathfinding.Position{}, false
		}
	}

	s.Occupants = append(s.Occupants, citizenID)
	s.occupiedCells[assignPos] = true
	grid.UpdateCell(assignPos, pathfinding.CellEscapedCitizen)

	if len(s.Occupants) == len(s.Cells) {
		s.HasCapacity = false
	}

	return assignPos, true
}

// Copy returns a deep copy of the SafeZone
func (s *SafeZone) Copy() SafeZone {
	cells := make([]pathfinding.Position, len(s.Cells))
	copy(cells, s.Cells)

	occupants := make([]uuid.UUID, len(s.Occupants))
	copy(occupants, s.Occupants)

	var occupiedCells map[pathfinding.Position]bool
	if s.occupiedCells != nil {
		occupiedCells = make(map[pathfinding.Position]bool, len(s.occupiedCells))
		maps.Copy(occupiedCells, s.occupiedCells)
	}

	return SafeZone{
		ID:            s.ID,
		Position:      s.Position,
		Radius:        s.Radius,
		Cells:         cells,
		HasCapacity:   s.HasCapacity,
		Occupants:     occupants,
		occupiedCells: occupiedCells,
	}
}
