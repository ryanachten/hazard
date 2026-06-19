// Package engine defines the simulation engine
package engine

import (
	pf "hazard/internal/pathfinding"
	"log"
	"math/rand"
)

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

func createCitizens(citizenCountRange [2]int, grid pf.Grid, safeZone pf.Position) []Citizen {
	citizenMin := citizenCountRange[0]
	citizenMax := citizenCountRange[1]
	citizenCount := citizenMin + rand.Intn(citizenMax-citizenMin+1)
	citizens := make([]Citizen, citizenCount)

	pathfinder := pf.AStar{}

	for i := range citizens {
		startPosition := grid.GetRandomPosition()
		path, err := pathfinder.FindPath(grid, startPosition, safeZone)
		if err != nil {
			log.Printf("Error creating path for citizen %v: %v", i, err)
			citizens[i] = Citizen{
				Status:           CitizenIdle,
				CurrentPath:      []pf.Position{startPosition},
				CurrentPathIndex: 0,
			}
			continue
		}

		citizens[i] = Citizen{
			Status:           CitizenIdle,
			CurrentPath:      path,
			CurrentPathIndex: 0,
		}
	}

	return citizens
}
