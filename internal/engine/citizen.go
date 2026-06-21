// Package engine defines the simulation engine
package engine

import (
	"fmt"
	pf "hazard/internal/pathfinding"
	"log"
)

// Citizen subjected to a hazard
type Citizen struct {
	ID                 int
	Status             CitizenStatus
	CurrentPosition    pf.Position
	CurrentDestination pf.Position
	Path               []pf.Position
	CurrentPathIndex   int
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

func createCitizens(citizenCountRange [2]int, grid *pf.Grid) []Citizen {
	citizenCount := randIntInRange(citizenCountRange)
	citizens := make([]Citizen, citizenCount)

	for i := range citizens {
		startPosition, err := grid.GetRandomOpenPosition()
		if err != nil {
			log.Printf("error creating citizen %v: %v", i, err)
			continue
		}

		citizens[i] = Citizen{
			ID:               i,
			Status:           CitizenIdle,
			CurrentPosition:  startPosition,
			CurrentPathIndex: 0,
		}

		err = citizens[i].findNearestSafeZone(grid)
		if err != nil {
			log.Printf("error updating citizen %v path: %v", i, err)
		}
	}

	return citizens
}

func (c *Citizen) findNearestSafeZone(grid *pf.Grid) error {
	dijkstra := pf.Dijkstra{}

	isGoal := func(pos pf.Position) bool {
		return grid.GetCell(pos) == pf.CellSafeZone
	}
	path, err := dijkstra.FindPathToGoal(grid, c.CurrentPosition, isGoal)

	if err != nil {
		return err
	}

	c.Path = path
	c.CurrentDestination = path[len(path)-1]
	return nil
}

func (c *Citizen) updatePath(grid *pf.Grid) error {
	aStar := pf.AStar{}

	var err error
	for range 3 {
		var path []pf.Position
		path, err = aStar.FindPath(grid, c.CurrentPosition, c.CurrentDestination)
		if err == nil {
			c.Path = path
			c.CurrentPathIndex = 0
			return nil
		}
	}

	return fmt.Errorf("pathfinding failed after 3 attempts: %w", err)
}

func (c *Citizen) incrementLocation() {
	if c.CurrentPathIndex < len(c.Path)-1 {
		c.Status = CitizenNavigating
		c.CurrentPathIndex++
		c.CurrentPosition = c.Path[c.CurrentPathIndex]

		log.Printf("Citizen %v now at %v of %v steps", c.ID, c.CurrentPathIndex, len(c.Path))
	}
}
