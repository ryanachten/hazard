// Package citizen defines how citizens behave in the simulation
package citizen

import (
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"hazard/internal/pathfinding"
	"hazard/internal/random"
	"hazard/internal/safezone"
)

// Citizen subjected to a hazard
type Citizen struct {
	ID                 uuid.UUID
	Status             Status
	CurrentPosition    pathfinding.Position
	CurrentDestination pathfinding.Position
	TargetSafeZone     *safezone.SafeZone
	Path               []pathfinding.Position
	CurrentPathIndex   int
	PreviousCellType   pathfinding.CellType
}

// Status defines the state of citizen activity
type Status string

const (
	// StatusIdle state pre-hazard taking place
	StatusIdle Status = "idle"
	// StatusNavigating when moving towards a given destination
	StatusNavigating Status = "navigating"
	// StatusEscaped when the citizen has reached a safe zone
	StatusEscaped Status = "escaped"
	// StatusDead when the citizen has been overtaken by a hazard
	StatusDead Status = "dead"
)

// CreateCitizens instantiates citizens in a grid
func CreateCitizens(citizenCount int, grid *pathfinding.Grid, safeZoneLocations map[pathfinding.Position]*safezone.SafeZone) []Citizen {
	citizens := make([]Citizen, 0, citizenCount)

	for range citizenCount {
		startPosition, err := grid.GetRandomOpenPosition(0, 0, grid.Width-1, grid.Height-1)
		if err != nil {
			slog.Warn("error getting position for citizen", "err", err)
			continue
		}

		citizen := Citizen{
			ID:               uuid.New(),
			Status:           StatusIdle,
			CurrentPosition:  startPosition,
			CurrentPathIndex: 0,
		}

		err = citizen.FindNearestSafeZone(grid, safeZoneLocations)
		if err != nil {
			slog.Warn("error updating citizen", "citizenID", citizen.ID, "err", err)
			continue
		}

		citizen.PreviousCellType = grid.GetCell(startPosition)
		grid.UpdateCell(startPosition, pathfinding.CellCitizen)

		citizens = append(citizens, citizen)
	}

	return citizens
}

// FindNearestSafeZone finds a path from the citizen's current position to the nearest safe zone with capacity
func (c *Citizen) FindNearestSafeZone(grid *pathfinding.Grid, safeZoneLocations map[pathfinding.Position]*safezone.SafeZone) error {
	isGoal := func(pos pathfinding.Position) bool {
		if grid.GetCell(pos) != pathfinding.CellSafeZone {
			return false
		}
		safeZone, ok := safeZoneLocations[pos]
		if !ok {
			return false
		}
		return safeZone.HasCapacity
	}
	path, err := pathfinding.FindPathToGoal(grid, c.CurrentPosition, isGoal)

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

// RecalculatePath recalculates the path from the citizen's current position to their destination
func (c *Citizen) RecalculatePath(grid *pathfinding.Grid) error {
	path, err := pathfinding.FindPath(grid, c.CurrentPosition, c.CurrentDestination)
	if err != nil {
		return fmt.Errorf("pathfinding failed: %w", err)
	}

	c.Path = path
	c.CurrentPathIndex = 0
	return nil
}

// IncrementLocation moves the citizen one step along their path and updates their status
func (c *Citizen) IncrementLocation(grid *pathfinding.Grid) (bool, bool) {
	if c.Status == StatusEscaped || c.Status == StatusDead {
		return false, c.Status == StatusEscaped
	}

	if len(c.Path) == 0 || c.CurrentPathIndex < 0 || c.CurrentPathIndex >= len(c.Path) {
		hasMoved := c.randomWalk(grid)
		return hasMoved, false
	}

	hasMoved := false

	if c.CurrentPathIndex < len(c.Path)-1 {
		c.Status = StatusNavigating
		c.CurrentPathIndex++
		c.updatePosition(c.Path[c.CurrentPathIndex], grid)
		hasMoved = true
	}

	return hasMoved, c.CurrentPathIndex == len(c.Path)-1
}

// randomWalk walks 1 step in a random direction
func (c *Citizen) randomWalk(grid *pathfinding.Grid) bool {
	curPos := c.CurrentPosition
	var validDirections []pathfinding.Position

	for _, direction := range pathfinding.Directions {
		dir := pathfinding.Position{
			X: curPos.X + direction.X,
			Y: curPos.Y + direction.Y,
		}
		if grid.InBounds(dir) && !pathfinding.AvoidableCellType[grid.GetCell(dir)] {
			validDirections = append(validDirections, dir)
		}
	}

	if len(validDirections) == 0 {
		return false
	}

	newPos := random.ValInSlice(validDirections)
	c.updatePosition(newPos, grid)

	return true
}

func (c *Citizen) updatePosition(newPos pathfinding.Position, grid *pathfinding.Grid) {
	grid.UpdateCell(c.CurrentPosition, c.PreviousCellType)

	c.CurrentPosition = newPos

	c.PreviousCellType = grid.GetCell(c.CurrentPosition)
	grid.UpdateCell(c.CurrentPosition, pathfinding.CellCitizen)
}

// Copy returns a deep copy of the Citizen
func (c *Citizen) Copy() Citizen {
	path := make([]pathfinding.Position, len(c.Path))
	copy(path, c.Path)

	var safeZone *safezone.SafeZone
	if c.TargetSafeZone != nil {
		sz := c.TargetSafeZone.Copy()
		safeZone = &sz
	}

	return Citizen{
		ID:                 c.ID,
		Status:             c.Status,
		CurrentPosition:    c.CurrentPosition,
		CurrentDestination: c.CurrentDestination,
		TargetSafeZone:     safeZone,
		Path:               path,
		CurrentPathIndex:   c.CurrentPathIndex,
		PreviousCellType:   c.PreviousCellType,
	}
}
