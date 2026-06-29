package common

import (
	pf "hazard/internal/pathfinding"
	"math/rand"

	"github.com/google/uuid"
)

// Hazard in a simulation
type Hazard struct {
	ID            uuid.UUID
	Type          HazardType
	CreatedAt     uint64
	Duration      int
	Origin        pf.Position
	CurrentRadius int
}

// HazardKind categorizes how a hazard behaves during simulation ticks.
type HazardKind string

const (
	// HazardKindExpanding hazard expands over time (i.e. flood)
	HazardKindExpanding HazardKind = "expanding"
	// HazardKindStrike hazard strikes specific cells (i.e. lightning)
	HazardKindStrike HazardKind = "strike"
	// HazardKindGlobal hazard impacts entire map (i.e. earthquake)
	HazardKindGlobal HazardKind = "global"
)

// HazardType is a registry-level type identifier.
type HazardType struct {
	Name string
	Kind HazardKind
}

// HazardTypes registry for defining hazard metadata
var HazardTypes = map[string]HazardType{
	"fire":  {Name: "fire", Kind: HazardKindExpanding},
	"flood": {Name: "flood", Kind: HazardKindExpanding},
	"lava":  {Name: "lava", Kind: HazardKindExpanding},
}

func randomHazardType() HazardType {
	keys := make([]string, 0, len(HazardTypes))
	for k := range HazardTypes {
		keys = append(keys, k)
	}
	k := keys[rand.Intn(len(keys))]
	return HazardTypes[k]
}

// CreateHazard creates a new hazard at a random open position on the grid
func CreateHazard(config HazardConfig, grid *pf.Grid) (Hazard, error) {
	duration := RandIntInRange(config.DurationRange)

	origin, err := grid.GetRandomOpenPosition()
	if err != nil {
		return Hazard{}, err
	}

	grid.UpdateCell(origin, pf.CellHazard)

	hazard := Hazard{
		ID:       uuid.New(),
		Duration: duration,
		Type:     randomHazardType(),
		Origin:   origin,
	}

	return hazard, nil
}

// ExpandHazard increases the hazard radius and marks affected cells on the grid
func (h *Hazard) ExpandHazard(grid *pf.Grid) []pf.Position {
	h.CurrentRadius++
	return h.updateCells(grid, pf.CellOpen, pf.CellHazard)
}

// RemoveHazard clears the hazard's cells from the grid
func (h *Hazard) RemoveHazard(grid *pf.Grid) []pf.Position {
	return h.updateCells(grid, pf.CellHazard, pf.CellOpen)
}

func (h *Hazard) updateCells(grid *pf.Grid, oldType, newType pf.CellType) []pf.Position {
	updatedCells := []pf.Position{}

	for dx := -h.CurrentRadius; dx <= h.CurrentRadius; dx++ {
		for dy := -h.CurrentRadius; dy <= h.CurrentRadius; dy++ {
			pos := pf.Position{X: h.Origin.X + dx, Y: h.Origin.Y + dy}
			if grid.InBounds(pos) && grid.GetCell(pos) == oldType {
				grid.UpdateCell(pos, newType)
				updatedCells = append(updatedCells, pos)
			}
		}
	}

	return updatedCells
}
