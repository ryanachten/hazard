// Package hazard defines how hazards behave in the simulation
package hazard

import (
	r "hazard/internal/numrange"
	pf "hazard/internal/pathfinding"
	"math/rand/v2"

	"github.com/google/uuid"
)

// Config configures simulation hazards
type Config struct {
	Probability   float32
	Count         int
	DurationRange r.PositiveRange
}

// Hazard in a simulation
type Hazard struct {
	ID            uuid.UUID
	Type          Type
	CreatedAt     uint64
	Duration      int
	Origin        pf.Position
	CurrentRadius int
}

// Kind categorizes how a hazard behaves during simulation ticks.
type Kind string

const (
	// KindExpanding hazard expands over time (i.e. flood)
	KindExpanding Kind = "expanding"
	// KindStrike hazard strikes specific cells (i.e. lightning)
	KindStrike Kind = "strike"
	// KindGlobal hazard impacts entire map (i.e. earthquake)
	KindGlobal Kind = "global"
)

// Type is a registry-level type identifier.
type Type string

const (
	// FireHazard fire-type hazard
	FireHazard Type = "fire"
	// FloodHazard flood-type hazard
	FloodHazard Type = "flood"
	// LavaHazard lava-type hazard
	LavaHazard Type = "lava"
)

var hazardTypes = []Type{
	FireHazard,
	FloodHazard,
	LavaHazard,
}

// Create creates a new hazard at a random open position on the grid
func Create(config Config, grid *pf.Grid) (Hazard, error) {
	duration := config.DurationRange.Random()

	origin, err := grid.GetRandomOpenPosition(0, 0, grid.Width-1, grid.Height-1)
	if err != nil {
		return Hazard{}, err
	}

	grid.UpdateCell(origin, pf.CellHazard)

	hazard := Hazard{
		ID:       uuid.New(),
		Duration: duration,
		Type:     hazardTypes[rand.IntN(len(hazardTypes))],
		Origin:   origin,
	}

	return hazard, nil
}

// expandTypes define what cells can be overwritten during expansion
var expandTypes = map[pf.CellType]struct{}{
	pf.CellOpen:    {},
	pf.CellCitizen: {},
}

// Expand increases the hazard radius and marks affected cells on the grid
func (h *Hazard) Expand(grid *pf.Grid) []pf.Position {
	h.CurrentRadius++
	return h.updateCells(grid, expandTypes, pf.CellHazard)
}

// removalTypes define what cells can be overwritten during removal
var removalTypes = map[pf.CellType]struct{}{
	pf.CellHazard: {},
}

// Remove clears the hazard's cells from the grid
func (h *Hazard) Remove(grid *pf.Grid) []pf.Position {
	return h.updateCells(grid, removalTypes, pf.CellOpen)
}

func (h *Hazard) updateCells(grid *pf.Grid, oldTypes map[pf.CellType]struct{}, newType pf.CellType) []pf.Position {
	updatedCells := []pf.Position{}

	for dx := -h.CurrentRadius; dx <= h.CurrentRadius; dx++ {
		for dy := -h.CurrentRadius; dy <= h.CurrentRadius; dy++ {
			pos := pf.Position{X: h.Origin.X + dx, Y: h.Origin.Y + dy}
			if grid.InBounds(pos) {
				if _, validType := oldTypes[grid.GetCell(pos)]; validType {
					grid.UpdateCell(pos, newType)
					updatedCells = append(updatedCells, pos)
				}
			}
		}
	}

	return updatedCells
}
