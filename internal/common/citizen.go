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
	TargetSafeZone     *SafeZone
	Path               []pf.Position
	CurrentPathIndex   int
	PreviousCellType   pf.CellType
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
func CreateCitizens(citizenCountRange PositiveRange, grid *pf.Grid, safeZoneLocations map[pf.Position]*SafeZone) []Citizen {
	citizenCount := citizenCountRange.Random()
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

		err = citizen.FindNearestSafeZone(grid, safeZoneLocations)
		if err != nil {
			log.Printf("error updating citizen %v path: %v", citizen.ID, err)
			continue
		}

		citizen.PreviousCellType = grid.GetCell(startPosition)
		grid.UpdateCell(startPosition, pf.CellCitizen)

		citizens = append(citizens, citizen)
	}

	return citizens
}

// FindNearestSafeZone finds a path from the citizen's current position to the nearest safe zone with capacity
func (c *Citizen) FindNearestSafeZone(grid *pf.Grid, safeZoneLocations map[pf.Position]*SafeZone) error {
	isGoal := func(pos pf.Position) bool {
		if grid.GetCell(pos) != pf.CellSafeZone {
			return false
		}
		safeZone, ok := safeZoneLocations[pos]
		if !ok {
			return false
		}
		return safeZone.HasCapacity
	}
	path, err := pf.FindPathToGoal(grid, c.CurrentPosition, isGoal)

	if err != nil {
		return err
	}

	pathDestination := path[len(path)-1]
	safeZoneDestination, ok := safeZoneLocations[pathDestination]
	if !ok {
		return fmt.Errorf("unable to find safe zone for location %v", pathDestination)
	}

	c.Path = path
	c.CurrentDestination = pathDestination
	c.TargetSafeZone = safeZoneDestination
	c.CurrentPathIndex = 0
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
func (c *Citizen) IncrementLocation(grid *pf.Grid) (bool, bool) {

	if c.Status == CitizenEscaped || c.Status == CitizenDead {
		return false, c.Status == CitizenEscaped
	}

	if len(c.Path) == 0 || c.CurrentPathIndex < 0 || c.CurrentPathIndex >= len(c.Path) {
		return false, false
	}

	hasMoved := false

	if c.CurrentPathIndex < len(c.Path)-1 {
		c.Status = CitizenNavigating

		// Revert cell to previous type
		grid.UpdateCell(c.CurrentPosition, c.PreviousCellType)

		// Update current position
		c.CurrentPathIndex++
		c.CurrentPosition = c.Path[c.CurrentPathIndex]
		hasMoved = true

		// Store previous cell type and update current cell
		c.PreviousCellType = grid.GetCell(c.CurrentPosition)
		grid.UpdateCell(c.CurrentPosition, pf.CellCitizen)
	}

	return hasMoved, c.CurrentPathIndex == len(c.Path)-1
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
		TargetSafeZone:     c.TargetSafeZone,
		Path:               path,
		CurrentPathIndex:   c.CurrentPathIndex,
		PreviousCellType:   c.PreviousCellType,
	}
}
