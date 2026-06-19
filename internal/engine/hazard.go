package engine

import (
	pf "hazard/internal/pathfinding"
	"math/rand"
)

// Hazard in a simulation
type Hazard struct {
	Type     HazardType
	Duration int
	Origin   pf.Position
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

func createHazard(config hazardConfig, grid pf.Grid) Hazard {
	durationMin := config.HazardDurationRange[0]
	durationMax := config.HazardDurationRange[1]
	duration := durationMin + rand.Intn(durationMax-durationMin+1)

	origin := grid.GetRandomPosition()

	grid.UpdateCell(origin, pf.CellHazard)

	hazard := Hazard{
		Duration: duration,
		Type:     randomHazardType(),
		Origin:   origin,
	}

	return hazard
}
