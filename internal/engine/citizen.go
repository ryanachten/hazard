// Package engine defines the simulation engine
package engine

import (
	"fmt"
	pf "hazard/internal/pathfinding"
	"log"
	"math/rand"
)

// Citizen subjected to a hazard
type Citizen struct {
	ID               int
	Status           CitizenStatus
	Path             []pf.Position
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

func createCitizens(citizenCountRange [2]int, grid *pf.Grid, safeZone pf.Position) []Citizen {
	citizenMin := citizenCountRange[0]
	citizenMax := citizenCountRange[1]
	citizenCount := citizenMin + rand.Intn(citizenMax-citizenMin+1)
	citizens := make([]Citizen, citizenCount)

	pathfinder := pf.AStar{}

	for i := range citizens {
		startPosition, err := grid.GetRandomOpenPosition()
		if err != nil {
			log.Printf("error creating citizen %v: %v", i, err)
			continue
		}

		citizens[i] = Citizen{
			ID:               i,
			Status:           CitizenIdle,
			CurrentPathIndex: 0,
			pathfinder:       &pathfinder,
		}
		err = citizens[i].updatePath(grid, startPosition, safeZone)
		if err != nil {
			log.Printf("error updating citizen %v path: %v", i, err)
		}
	}

	return citizens
}

func (c *Citizen) updatePath(grid *pf.Grid, start, end pf.Position) error {
	var err error
	for range 3 {
		var path []pf.Position
		path, err = c.pathfinder.FindPath(grid, start, end)
		if err == nil {
			c.Path = path
			c.CurrentPathIndex = 0
			return nil
		}
	}

	return fmt.Errorf("pathfinding failed after 3 attempts: %w", err)
}
