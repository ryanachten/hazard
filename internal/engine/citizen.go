// Package engine defines the simulation engine
package engine

import pf "hazard/internal/pathfinding"

// Citizen subjected to a hazard
type Citizen struct {
	Status           CitizenStatus
	CurrentPath      []pf.Position
	CurrentPathIndex int
}

// CitizenStatus defines the state of citizen activity
type CitizenStatus string

const (
	// CitizenIdle state pre-hazard taking place
	CitizenIdle CitizenStatus = "idle"
	// CitizenNavigating when moving towards a given destination
	CitizenNavigating CitizenStatus = "navigating"
	// CitizenEscaped when the citizen has reached a safe zone
	CitizenEscaped CitizenStatus = "escaped"
	// CitizenDead when the citizen has been overtaken by a hazard
	CitizenDead CitizenStatus = "dead"
)
