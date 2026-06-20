package engine

import (
	pf "hazard/internal/pathfinding"
	"log"
	"math/rand"
	"slices"
)

// Simulation engine for hazards
type Simulation struct {
	Config    SimulationConfig
	State     SimulationState
	TickCount uint64
	Citizens  []Citizen
	SafeZone  pf.Position
	Grid      *pf.Grid
	Hazards   []Hazard
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

	safeZone, err := grid.GetRandomOpenPosition()
	if err != nil {
		return Simulation{}, err
	}

	return Simulation{
		Config:    config,
		State:     SimulationCreated,
		TickCount: 0,
		Grid:      &grid,
		SafeZone:  safeZone,
		Citizens:  createCitizens(config.CitizenCountRange, &grid, safeZone),
	}, nil
}

// Tick increments a simulation by one tick
func (s *Simulation) Tick() {
	if s.State == SimulationPaused || s.State == SimulationCompleted {
		return
	}
	s.State = SimulationRunning

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
	if len(s.Hazards) < hazardConfig.MaxHazards && rand.Float32() <= hazardConfig.Probability {
		hazard, err := createHazard(hazardConfig, s.Grid)
		if err != nil {
			log.Printf("error creating hazard: %v", err)
		} else {
			hazard.CreatedAt = s.TickCount
			s.Hazards = append(s.Hazards, hazard)
		}
	}

	// Update citizen pathfinding state
	for i := range s.Citizens {

		// Check if any path intersects with hazards and needs recalculating
		for _, pos := range s.Citizens[i].Path {
			if s.Grid.Cells[pos.Y][pos.X] == pf.CellHazard {
				err := s.Citizens[i].updatePath(s.Grid, s.Citizens[i].Path[s.Citizens[i].CurrentPathIndex], s.SafeZone)
				if err != nil {
					log.Printf("error updating citizen %v path: %v", i, err)
				}
				break
			}
		}

		// Increment citizen's position on path
		if s.Citizens[i].CurrentPathIndex < len(s.Citizens[i].Path)-1 {
			s.Citizens[i].Status = CitizenNavigating
			s.Citizens[i].CurrentPathIndex++

			log.Printf("Citizen %v now at %v of %v steps", i, s.Citizens[i].CurrentPathIndex, len(s.Citizens[i].Path))
		}
	}

	s.TickCount++
}
