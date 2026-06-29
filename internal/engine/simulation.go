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
	EventEmitter         events.EventEmitter
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
func NewSimulation(config c.SimulationConfig) (Simulation, error) {
	grid := pf.NewGrid(config.Width, config.Height, pf.CellOpen)

	safeZone, err := c.CreateSafeZone(config.SafeZone, &grid)
	if err != nil {
		return Simulation{}, err
	}

	var simulation = Simulation{
		Config:       config,
		State:        SimulationCreated,
		TickCount:    0,
		Grid:         &grid,
		MaxHazards:   c.RandIntInRange(config.Hazard.CountRange),
		MaxSafeZones: c.RandIntInRange(config.SafeZone.CountRange),
		SafeZones:    []c.SafeZone{safeZone},
		Citizens:     c.CreateCitizens(config.CitizenCountRange, &grid),
		EventEmitter: &events.InMemoryEventLog{},
	}

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
		err := s.EventEmitter.SimulationCompleted(s.getEventMetadata())
		if err != nil {
			log.Printf("error emitting SimulationCompleted event: %v", err)
		}
		s.State = SimulationCompleted
	}
}

// Events captured during simulation
func (s *Simulation) Events() []events.SimulationEvent {
	return s.EventEmitter.Events()
}

func (s *Simulation) updateOrRemoveHazards() {
	for i := len(s.Hazards) - 1; i >= 0; i-- {
		if s.TickCount > s.Hazards[i].CreatedAt+uint64(s.Hazards[i].Duration) {
			hazardID := s.Hazards[i].ID
			updatedCells := s.Hazards[i].RemoveHazard(s.Grid)
			s.Hazards = slices.Delete(s.Hazards, i, i+1)
			err := s.EventEmitter.HazardDissipated(hazardID, updatedCells, s.getEventMetadata())
			if err != nil {
				log.Printf("error emitting HazardDissipated event: %v", err)
			}
		} else {
			updatedCells := s.Hazards[i].ExpandHazard(s.Grid)
			err := s.EventEmitter.HazardExpanded(s.Hazards[i].ID, updatedCells, s.getEventMetadata())
			if err != nil {
				log.Printf("error emitting HazardExpanded event: %v", err)
			}
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

	err = s.EventEmitter.HazardEmerged(hazard.ID, hazard.Origin, s.getEventMetadata())
	if err != nil {
		log.Printf("error emitting HazardEmerged event: %v", err)
	}
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

	err = s.EventEmitter.SafeZoneEmerged(safeZone.ID, safeZone.Cells, s.getEventMetadata())
	if err != nil {
		log.Printf("error emitting SafeZoneEmerged event: %v", err)
	}

	return true
}

func (s *Simulation) removeDeadCitizen(citizenIndex int) bool {
	if s.Grid.GetCell(s.Citizens[citizenIndex].CurrentPosition) != pf.CellHazard {
		return false
	}

	err := s.EventEmitter.CitizenDied(s.Citizens[citizenIndex].ID, s.getEventMetadata())
	if err != nil {
		log.Printf("error emitting CitizenDied event: %v", err)
		return false
	}

	s.Citizens[citizenIndex].Status = c.CitizenDead
	s.DeadCitizensCount++
	return true
}

func (s *Simulation) updateCitizenPath(citizenIndex int, safeZoneCreated bool) {
	pathUpdated := false

	// If new safe zone added, determine which safe zone is closest
	if safeZoneCreated {
		err := s.Citizens[citizenIndex].FindNearestSafeZone(s.Grid)
		if err != nil {
			log.Printf("error finding safe zone for citizen %v path: %v", citizenIndex, err)
			return
		}
		pathUpdated = true
	} else {
		// Check if any path intersects with hazards and needs recalculating
		for _, pos := range s.Citizens[citizenIndex].Path {
			if s.Grid.Cells[pos.Y][pos.X] == pf.CellHazard {
				err := s.Citizens[citizenIndex].UpdatePath(s.Grid)
				if err != nil {
					log.Printf("error updating citizen %v path: %v", citizenIndex, err)
				}
				pathUpdated = true
				break
			}
		}
	}

	if pathUpdated {
		err := s.EventEmitter.CitizenPathUpdated(s.Citizens[citizenIndex].ID, s.Citizens[citizenIndex].Path, s.getEventMetadata())
		if err != nil {
			log.Printf("error emitting CitizenPathUpdated event: %v", err)
		}
	}
}

func (s *Simulation) updateCitizenLocation(citizenIndex int) {
	hasMoved := s.Citizens[citizenIndex].IncrementLocation()
	if hasMoved {
		err := s.EventEmitter.CitizenMoved(s.Citizens[citizenIndex].ID, s.Citizens[citizenIndex].CurrentPosition, s.getEventMetadata())
		if err != nil {
			log.Printf("error emitting CitizenMoved event: %v", err)
		}
	}

	if s.Citizens[citizenIndex].Status == c.CitizenEscaped {
		s.EscapedCitizensCount++
		err := s.EventEmitter.CitizenEscaped(s.Citizens[citizenIndex].ID, s.getEventMetadata())
		if err != nil {
			log.Printf("error emitting CitizenEscaped event: %v", err)
		}
	}
}

func (s *Simulation) getEventMetadata() events.EventMetadata {
	return events.EventMetadata{
		SimulationID: s.ID,
		Tick:         s.TickCount,
	}
}
