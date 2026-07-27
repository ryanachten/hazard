// Package hazard defines how hazards behave in the simulation
package hazard

import (
	"math/rand"

	"github.com/google/uuid"

	"hazard/internal/bounds"
	"hazard/internal/pathfinding"
)

// Config configures simulation hazards
type Config struct {
	Probability   float32
	Count         int
	DurationRange bounds.PositiveRange
}

// Hazard in a simulation
type Hazard struct {
	ID            uuid.UUID
	Type          Type
	CreatedAt     uint64
	Duration      int
	Origin        pathfinding.Position
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
func Create(config Config, grid *pathfinding.Grid) (Hazard, error) {
	duration := config.DurationRange.Random()

	origin, err := grid.GetRandomOpenPosition(0, 0, grid.Width-1, grid.Height-1)
	if err != nil {
		return Hazard{}, err
	}

	grid.UpdateCell(origin, pathfinding.CellHazard)

	hz := Hazard{
		ID:       uuid.New(),
		Duration: duration,
		Type:     hazardTypes[rand.Intn(len(hazardTypes))],
		Origin:   origin,
	}

	return hz, nil
}

// expandTypes define what cells can be overwritten during expansion
var expandTypes = map[pathfinding.CellType]struct{}{
	pathfinding.CellOpen:    {},
	pathfinding.CellCitizen: {},
}

// Expand increases the hazard radius and marks affected cells on the grid
func (h *Hazard) Expand(grid *pathfinding.Grid) []pathfinding.Position {
	h.CurrentRadius++
	return h.updateCells(grid, expandTypes, pathfinding.CellHazard)
}

// removalTypes define what cells can be overwritten during removal
var removalTypes = map[pathfinding.CellType]struct{}{
	pathfinding.CellHazard: {},
}

// Remove clears the hazard's cells from the grid
func (h *Hazard) Remove(grid *pathfinding.Grid) []pathfinding.Position {
	return h.updateCells(grid, removalTypes, pathfinding.CellOpen)
}

func (h *Hazard) updateCells(grid *pathfinding.Grid, oldTypes map[pathfinding.CellType]struct{}, newType pathfinding.CellType) []pathfinding.Position {
	updatedCells := []pathfinding.Position{}

	for dx := -h.CurrentRadius; dx <= h.CurrentRadius; dx++ {
		for dy := -h.CurrentRadius; dy <= h.CurrentRadius; dy++ {
			pos := pathfinding.Position{X: h.Origin.X + dx, Y: h.Origin.Y + dy}
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
