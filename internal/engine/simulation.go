package engine

import (
	c "hazard/internal/common"
	"hazard/internal/events"
	pf "hazard/internal/pathfinding"
	"log"
	"math/rand"
	"slices"

	"github.com/google/uuid"
)

// Simulation engine for hazards
type Simulation struct {
	ID                   uuid.UUID
	Config               c.SimulationConfig
	State                SimulationState
	TickCount            uint64
	Grid                 *pf.Grid
	Citizens             []c.Citizen
	DeadCitizensCount    int
	EscapedCitizensCount int
	MaxHazards           int
	Hazards              []c.Hazard
	MaxSafeZones         int
	SafeZones            []c.SafeZone
	eventBus             *events.EventBus
	safeZoneLocations    map[pf.Position]*c.SafeZone
}

// SimulationState phases of a simulation
type SimulationState string

const (
	// SimulationCreated simulation created but not running
	SimulationCreated SimulationState = "created"
	// SimulationRunning simulation running
	SimulationRunning SimulationState = "running"
	// SimulationPaused simulation paused
	SimulationPaused SimulationState = "paused"
	// SimulationCompleted simulation completed
	SimulationCompleted SimulationState = "completed"
)

// NewSimulation creates a simulation based on configuration
func NewSimulation(width, height int, config c.SimulationConfig, eventBus *events.EventBus) (Simulation, error) {
	grid := pf.NewGrid(width, height, pf.CellOpen)

	safeZone, err := c.CreateSafeZone(config.SafeZone, &grid)
	if err != nil {
		return Simulation{}, err
	}

	safeZoneLocations := make(map[pf.Position]*c.SafeZone)
	for _, cell := range safeZone.Cells {
		safeZoneLocations[cell] = &safeZone
	}

	obstacles := c.CreateObstacles(config.Obstacle, &grid)

	var simulation = Simulation{
		ID:                uuid.New(),
		Config:            config,
		State:             SimulationRunning,
		TickCount:         0,
		Grid:              &grid,
		MaxHazards:        config.Hazard.CountRange.Random(),
		MaxSafeZones:      config.SafeZone.CountRange.Random(),
		SafeZones:         []c.SafeZone{safeZone},
		Citizens:          c.CreateCitizens(config.CitizenCountRange, &grid, safeZoneLocations),
		eventBus:          eventBus,
		safeZoneLocations: safeZoneLocations,
	}

	simulation.eventBus.SimulationCreated(
		events.SimulationCreatedPayload{
			Grid:      simulation.Grid.Copy(),
			Citizens:  simulation.Citizens,
			SafeZones: simulation.SafeZones,
			Obstacles: obstacles,
		},
		events.EventMetadata{
			SimulationID: simulation.ID,
			Tick:         simulation.TickCount,
		})

	return simulation, nil
}

// Tick increments a simulation by one tick
func (s *Simulation) Tick() {
	if s.State == SimulationPaused || s.State == SimulationCompleted {
		return
	}
	s.State = SimulationRunning

	s.updateOrRemoveHazards()
	s.generateIntermittentHazard()

	safeZoneCreated := s.generateIntermittentSafeZone()

	for i := range s.Citizens {

		if s.Citizens[i].Status == c.CitizenDead || s.Citizens[i].Status == c.CitizenEscaped {
			continue
		}

		isDead := s.removeDeadCitizen(i)
		if isDead {
			continue
		}

		s.updateCitizenPath(i, safeZoneCreated)
		s.updateCitizenLocation(i)
	}

	s.TickCount++

	if len(s.Citizens) > 0 && s.DeadCitizensCount+s.EscapedCitizensCount == len(s.Citizens) {
		s.eventBus.SimulationCompleted(s.getEventMetadata())
		s.State = SimulationCompleted
	}
}

// ProcessCommand handles a simulation command, updating state accordingly.
func (s *Simulation) ProcessCommand(cmd events.SimulationCommand) {
	switch cmd.CommandType {
	case events.PauseSimulation:
		if s.State == SimulationRunning {
			s.State = SimulationPaused
		} else {
			s.State = SimulationRunning
		}
	}
}

func (s *Simulation) updateOrRemoveHazards() {
	for i := len(s.Hazards) - 1; i >= 0; i-- {
		if s.TickCount > s.Hazards[i].CreatedAt+uint64(s.Hazards[i].Duration) {
			hazardID := s.Hazards[i].ID
			updatedCells := s.Hazards[i].RemoveHazard(s.Grid)
			s.Hazards = slices.Delete(s.Hazards, i, i+1)
			s.eventBus.HazardDissipated(hazardID, updatedCells, s.getEventMetadata())
		} else {
			updatedCells := s.Hazards[i].ExpandHazard(s.Grid)
			s.eventBus.HazardExpanded(s.Hazards[i].ID, updatedCells, s.getEventMetadata())
		}
	}
}

func (s *Simulation) generateIntermittentHazard() {
	hazardConfig := s.Config.Hazard
	if len(s.Hazards) >= s.MaxHazards || rand.Float32() > hazardConfig.Probability {
		return
	}

	hazard, err := c.CreateHazard(hazardConfig, s.Grid)
	if err != nil {
		log.Printf("error creating hazard: %v", err)
		return
	}

	hazard.CreatedAt = s.TickCount
	s.Hazards = append(s.Hazards, hazard)

	s.eventBus.HazardEmerged(hazard.ID, events.HazardEmergedPayload{
		Type:     hazard.Type,
		Position: hazard.Origin,
	}, s.getEventMetadata())
}

func (s *Simulation) generateIntermittentSafeZone() bool {
	safeZoneConfig := s.Config.SafeZone
	if len(s.SafeZones) >= s.MaxSafeZones || rand.Float32() > safeZoneConfig.Probability {
		return false
	}

	safeZone, err := c.CreateSafeZone(safeZoneConfig, s.Grid)
	if err != nil {
		log.Printf("error creating safe zone: %v", err)
		return false
	}

	s.SafeZones = append(s.SafeZones, safeZone)

	for _, cell := range safeZone.Cells {
		s.safeZoneLocations[cell] = &safeZone
	}

	s.eventBus.SafeZoneEmerged(safeZone.ID, events.SafeZoneEmergedPayload{
		ID:    safeZone.ID,
		Cells: safeZone.Cells,
	}, s.getEventMetadata())

	return true
}

func (s *Simulation) removeDeadCitizen(citizenIndex int) bool {
	if s.Citizens[citizenIndex].Status == c.CitizenEscaped {
		return false
	}

	if s.Grid.GetCell(s.Citizens[citizenIndex].CurrentPosition) != pf.CellHazard {
		return false
	}

	s.Citizens[citizenIndex].Status = c.CitizenDead
	s.Grid.UpdateCell(s.Citizens[citizenIndex].CurrentPosition, pf.CellDeadCitizen)
	s.DeadCitizensCount++
	s.eventBus.CitizenDied(s.Citizens[citizenIndex].ID, s.getEventMetadata())

	return true
}

func (s *Simulation) updateCitizenPath(citizenIndex int, safeZoneCreated bool) {
	pathUpdated := false

	// If no safe zone assigned, new safe zone added, or the target has no capacity
	// determine which available safe zone is closest
	targetSafeZone := s.Citizens[citizenIndex].TargetSafeZone
	if safeZoneCreated || targetSafeZone == nil || !targetSafeZone.HasCapacity {
		if err := s.Citizens[citizenIndex].FindNearestSafeZone(s.Grid, s.safeZoneLocations); err != nil {
			s.Citizens[citizenIndex].Path = nil
		}
		pathUpdated = true
	} else {
		// Check if next cell intersects with avoidable cell types and needs recalculating
		curIndex := s.Citizens[citizenIndex].CurrentPathIndex
		nextIndex := curIndex + 1
		if nextIndex < len(s.Citizens[citizenIndex].Path) {
			pos := s.Citizens[citizenIndex].Path[nextIndex]
			if pf.AvoidableCellType[s.Grid.GetCell(pos)] {
				if err := s.Citizens[citizenIndex].RecalculatePath(s.Grid); err != nil {
					// Destination itself is blocked (e.g. occupied by another citizen).
					// Find a new safe zone with capacity.
					if err := s.Citizens[citizenIndex].FindNearestSafeZone(s.Grid, s.safeZoneLocations); err != nil {
						s.Citizens[citizenIndex].Path = nil
					}
				}
				pathUpdated = true
			}
		}
	}

	if pathUpdated {
		s.eventBus.CitizenPathUpdated(s.Citizens[citizenIndex].ID, s.Citizens[citizenIndex].Path, s.getEventMetadata())
	}
}

func (s *Simulation) updateCitizenLocation(citizenIndex int) {
	hasMoved, hasEscaped := s.Citizens[citizenIndex].IncrementLocation(s.Grid)
	if hasMoved {
		s.eventBus.CitizenMoved(s.Citizens[citizenIndex].ID, s.Citizens[citizenIndex].CurrentPosition, s.getEventMetadata())
	}

	if hasEscaped {
		citizen := &s.Citizens[citizenIndex]
		safeZone := s.safeZoneLocations[citizen.CurrentPosition]
		assignedPosition, hasCapacity := safeZone.AddOccupant(citizen.ID, citizen.CurrentPosition, s.Grid)

		if !hasCapacity {
			s.updateCitizenPath(citizenIndex, true)
		} else {
			s.EscapedCitizensCount++
			citizen.Status = c.CitizenEscaped

			if assignedPosition != citizen.CurrentPosition {
				s.Grid.UpdateCell(citizen.CurrentPosition, citizen.PreviousCellType)
				citizen.CurrentPosition = assignedPosition
			}

			s.eventBus.CitizenEscaped(citizen.ID, events.CitizenEscapedPayload{
				SafeZoneID:       safeZone.ID,
				AssignedPosition: assignedPosition,
			}, s.getEventMetadata())
		}
	}
}

func (s *Simulation) getEventMetadata() events.EventMetadata {
	return events.EventMetadata{
		SimulationID: s.ID,
		Tick:         s.TickCount,
	}
}
