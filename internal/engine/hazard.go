package engine

import (
	pf "hazard/internal/pathfinding"
	"math/rand"
)

// Hazard in a simulation
type Hazard struct {
	Type          HazardType
	CreatedAt     uint64
	Duration      int
	Origin        pf.Position
	CurrentRadius int
}

// HazardConfig configures a hazard
type HazardConfig struct {
	DurationRange [2]int
	Probability   float32
	MaxHazards    int
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

func createHazard(config HazardConfig, grid *pf.Grid) (Hazard, error) {
	durationMin := config.DurationRange[0]
	durationMax := config.DurationRange[1]
	duration := durationMin + rand.Intn(durationMax-durationMin+1)

	origin, err := grid.GetRandomOpenPosition()
	if err != nil {
		return Hazard{}, err
	}

	grid.UpdateCell(origin, pf.CellHazard)

	hazard := Hazard{
		Duration: duration,
		Type:     randomHazardType(),
		Origin:   origin,
	}

	return hazard, nil
}

func (h *Hazard) expandHazard(grid *pf.Grid) {
	h.CurrentRadius++
	h.updateHazardCells(grid, pf.CellHazard)
}

func (h *Hazard) removeHazard(grid *pf.Grid) {
	h.updateHazardCells(grid, pf.CellOpen)
}

func (h *Hazard) updateHazardCells(grid *pf.Grid, cellType pf.CellType) {
	for _, direction := range pf.Directions {
		newPosition := pf.Position{
			X: h.Origin.X + direction.X*h.CurrentRadius,
			Y: h.Origin.Y + direction.Y*h.CurrentRadius,
		}

		if grid.InBounds(newPosition) {
			grid.UpdateCell(newPosition, cellType)
		}
	}
}
