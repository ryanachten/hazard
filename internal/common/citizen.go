// Package common define common types and utilities in the project
package common

import (
	"fmt"
	pf "hazard/internal/pathfinding"
	"log"

	"github.com/google/uuid"
)

// Citizen subjected to a hazard
type Citizen struct {
	ID                 uuid.UUID
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

// CreateCitizens instantiates citizens in a grid
func CreateCitizens(citizenCountRange [2]int, grid *pf.Grid) []Citizen {
	citizenCount := RandIntInRange(citizenCountRange)
	citizens := make([]Citizen, 0, citizenCount)

	for range citizenCount {
		startPosition, err := grid.GetRandomOpenPosition(0, 0, grid.Width-1, grid.Height-1)
		if err != nil {
			log.Printf("error getting position for citizen: %v", err)
			continue
		}

		citizen := Citizen{
			ID:               uuid.New(),
			Status:           CitizenIdle,
			CurrentPosition:  startPosition,
			CurrentPathIndex: 0,
		}

		err = citizen.FindNearestSafeZone(grid)
		if err != nil {
			log.Printf("error updating citizen %v path: %v", citizen.ID, err)
			continue
		}

		citizens = append(citizens, citizen)
	}

	return citizens
}

// FindNearestSafeZone finds a path from the citizen's current position to the nearest safe zone
func (c *Citizen) FindNearestSafeZone(grid *pf.Grid) error {
	isGoal := func(pos pf.Position) bool {
		return grid.GetCell(pos) == pf.CellSafeZone
	}
	path, err := pf.FindPathToGoal(grid, c.CurrentPosition, isGoal)

	if err != nil {
		return err
	}

	c.Path = path
	c.CurrentDestination = path[len(path)-1]
	return nil
}

// UpdatePath recalculates the path from the citizen's current position to their destination
func (c *Citizen) UpdatePath(grid *pf.Grid) error {
	path, err := pf.FindPath(grid, c.CurrentPosition, c.CurrentDestination)
	if err != nil {
		return fmt.Errorf("pathfinding failed: %w", err)
	}

	c.Path = path
	c.CurrentPathIndex = 0
	return nil
}

// IncrementLocation moves the citizen one step along their path and updates their status
func (c *Citizen) IncrementLocation() bool {
	hasMoved := false

	if c.Status == CitizenEscaped || c.Status == CitizenDead {
		return hasMoved
	}

	if c.CurrentPathIndex < len(c.Path)-1 {
		c.CurrentPathIndex++
		c.CurrentPosition = c.Path[c.CurrentPathIndex]
		hasMoved = true
	}

	if c.CurrentPathIndex == len(c.Path)-1 {
		c.Status = CitizenEscaped
	} else {
		c.Status = CitizenNavigating
	}

	return hasMoved
}

// Copy returns a deep copy of the Citizen
func (c *Citizen) Copy() Citizen {
	path := make([]pf.Position, len(c.Path))
	copy(path, c.Path)
	return Citizen{
		ID:                 c.ID,
		Status:             c.Status,
		CurrentPosition:    c.CurrentPosition,
		CurrentDestination: c.CurrentDestination,
		Path:               path,
		CurrentPathIndex:   c.CurrentPathIndex,
	}
}
