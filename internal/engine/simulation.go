package engine

import (
	pf "hazard/internal/pathfinding"
	"log"
	"math/rand"
	"slices"
)

// Simulation engine for hazards
type Simulation struct {
	Config               SimulationConfig
	State                SimulationState
	TickCount            uint64
	Grid                 *pf.Grid
	Citizens             []Citizen
	DeadCitizensCount    int
	EscapedCitizensCount int
	MaxHazards           int
	Hazards              []Hazard
	MaxSafeZones         int
	SafeZones            []SafeZone
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
func NewSimulation(config SimulationConfig) (Simulation, error) {
	grid := pf.NewGrid(config.Width, config.Height, pf.CellOpen)

	safeZone, err := createSafeZone(config.SafeZone, &grid)
	if err != nil {
		return Simulation{}, err
	}

	return Simulation{
		Config:       config,
		State:        SimulationCreated,
		TickCount:    0,
		Grid:         &grid,
		MaxHazards:   randIntInRange(config.Hazard.CountRange),
		MaxSafeZones: randIntInRange(config.SafeZone.CountRange),
		SafeZones:    []SafeZone{safeZone},
		Citizens:     createCitizens(config.CitizenCountRange, &grid),
	}, nil
}

// Tick increments a simulation by one tick
func (s *Simulation) Tick() {
	if s.State == SimulationPaused || s.State == SimulationCompleted {
		return
	}
	s.State = SimulationRunning

	// Update or remove hazards
	hazardConfig := s.Config.Hazard
	for i := len(s.Hazards) - 1; i >= 0; i-- {
		if s.TickCount > s.Hazards[i].CreatedAt+uint64(s.Hazards[i].Duration) {
			s.Hazards[i].removeHazard(s.Grid)
			s.Hazards = slices.Delete(s.Hazards, i, i+1)
		} else {
			s.Hazards[i].expandHazard(s.Grid)
		}
	}

	// Randomly generate hazards within hazard limits
	if len(s.Hazards) < s.MaxHazards && rand.Float32() <= hazardConfig.Probability {
		hazard, err := createHazard(hazardConfig, s.Grid)
		if err != nil {
			log.Printf("error creating hazard: %v", err)
		} else {
			hazard.CreatedAt = s.TickCount
			s.Hazards = append(s.Hazards, hazard)
		}
	}

	// Randomly generate safe zones within hazard limits
	safeZoneConfig := s.Config.SafeZone
	safeZoneCreated := false
	if len(s.SafeZones) < s.MaxSafeZones && rand.Float32() <= safeZoneConfig.Probability {
		safeZone, err := createSafeZone(safeZoneConfig, s.Grid)
		if err != nil {
			log.Printf("error creating safe zone: %v", err)
		} else {
			s.SafeZones = append(s.SafeZones, safeZone)
			safeZoneCreated = true
		}
	}

	// Update citizen pathfinding state
	for i := range s.Citizens {

		// Skip dead and escaped citizens
		if s.Citizens[i].Status == CitizenDead || s.Citizens[i].Status == CitizenEscaped {
			continue
		}

		// If a hazard has caught up with a citizen, they're now dead
		if s.Grid.GetCell(s.Citizens[i].CurrentPosition) == pf.CellHazard {
			s.Citizens[i].Status = CitizenDead
			s.DeadCitizensCount++
			continue
		}

		// If new safe zone added, determine which safe zone is closest
		if safeZoneCreated {
			err := s.Citizens[i].findNearestSafeZone(s.Grid)
			if err != nil {
				log.Printf("error finding safe zone for citizen %v path: %v", i, err)
			}
		} else {
			// Check if any path intersects with hazards and needs recalculating
			for _, pos := range s.Citizens[i].Path {
				if s.Grid.Cells[pos.Y][pos.X] == pf.CellHazard {
					err := s.Citizens[i].updatePath(s.Grid)
					if err != nil {
						log.Printf("error updating citizen %v path: %v", i, err)
					}
					break
				}
			}
		}

		// Increment citizen's position on path
		s.Citizens[i].incrementLocation()
		if s.Citizens[i].Status == CitizenEscaped {
			s.EscapedCitizensCount++
		}
	}

	s.TickCount++

	if len(s.Citizens) > 0 && s.DeadCitizensCount+s.EscapedCitizensCount == len(s.Citizens) {
		s.State = SimulationCompleted
	}
}
