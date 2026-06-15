// Package engine defines the simulation engine
package engine

// Citizen subjected to a hazard
type Citizen struct {
	Status CitizenStatus
}

// CitizenStatus defines the state of citizen activity
type CitizenStatus string

const (
	// CitizenIdle state pre-hazard taking place
	CitizenIdle CitizenStatus = "idle"
	// CitizenNavigating when moving towards a given destination
	CitizenNavigating CitizenStatus = "navigating"
)
