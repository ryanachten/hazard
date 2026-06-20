// Package engine defines the simulation engine
package engine

import (
	pf "hazard/internal/pathfinding"
	"log"
	"math/rand"
)

// Citizen subjected to a hazard
type Citizen struct {
	ID               int
	Status           CitizenStatus
	CurrentPath      []pf.Position
	CurrentPathIndex int
	pathfinder       pf.Pathfinder
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

func createCitizens(citizenCountRange [2]int, grid pf.Grid, safeZone pf.Position) []Citizen {
	citizenMin := citizenCountRange[0]
	citizenMax := citizenCountRange[1]
	citizenCount := citizenMin + rand.Intn(citizenMax-citizenMin+1)
	citizens := make([]Citizen, citizenCount)

	pathfinder := pf.AStar{}

	for i := range citizens {
		startPosition := grid.GetRandomPosition()
		citizen := Citizen{
			ID:               i,
			Status:           CitizenIdle,
			CurrentPathIndex: 0,
			pathfinder:       &pathfinder,
		}
		citizen.updatePath(grid, startPosition, safeZone, 0)
	}

	return citizens
}

func (c *Citizen) updatePath(grid pf.Grid, startPosition, endPosition pf.Position, attempts int) {
	path, err := c.pathfinder.FindPath(grid, startPosition, endPosition)

	if err != nil && attempts < 3 {
		log.Printf("Error creating path for citizen %v: %v", c.ID, err)

		c.updatePath(grid, startPosition, endPosition, attempts+1)
	}

	c.CurrentPath = path
	c.CurrentPathIndex = 0
}
