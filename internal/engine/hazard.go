package engine

// Hazard in a simulation
type Hazard struct {
	Type  HazardType
	State HazardState
}

// HazardState defines the lifecycle of a hazard
type HazardState string

const (
	// HazardScheduled hazard is scheduled to appear in the future
	HazardScheduled HazardState = "scheduled"
	// HazardActive hazard is actively impacting map
	HazardActive HazardState = "active"
	// HazardDissipated hazard is no longer actively impacting map
	HazardDissipated HazardState = "dissipated"
)

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
