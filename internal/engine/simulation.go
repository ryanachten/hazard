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

// SimulationConfig configuration for a simulation
type SimulationConfig struct {
	TickIntervalMs    int
	Width             int
	Height            int
	CitizenCountRange [2]int
	Hazard            HazardConfig
}

// NewSimulation creates a simulation based on configuration
func NewSimulation(config SimulationConfig) Simulation {
	if config.Width <= 0 || config.Height <= 0 {
		panic("simulation width and height must be greater than zero")
	}
	if config.CitizenCountRange[0] > config.CitizenCountRange[1] {
		panic("CitizenCountRange[0] must be less than or equal to CitizenCountRange[1]")
	}

	grid := pf.NewGrid(config.Width, config.Height, pf.CellOpen)

	safeZone := grid.GetRandomPosition()

	return Simulation{
		Config:    config,
		State:     SimulationCreated,
		TickCount: 0,
		Grid:      &grid,
		SafeZone:  safeZone,
		Citizens:  createCitizens(config.CitizenCountRange, grid, safeZone),
	}
}

// Tick increments a simulation by one tick
func (s *Simulation) Tick() {
	if s.State == SimulationPaused || s.State == SimulationCompleted {
		return
	}
	s.State = SimulationRunning

	hazardConfig := s.Config.Hazard
	for i, hazard := range s.Hazards {
		// If hazard has expired, remove it from the simulation
		if s.TickCount > hazard.CreatedAt+uint64(hazard.Duration) {
			hazard.removeHazard(*s.Grid)
			s.Hazards = slices.Delete(s.Hazards, i, i+1)
			continue
		}

		hazard.expandHazard(*s.Grid)
	}

	// Randomly generate hazards within hazard limits
	if len(s.Hazards) < hazardConfig.MaxHazards && rand.Float32() <= hazardConfig.HazardProbability {
		hazard := createHazard(hazardConfig, *s.Grid)
		hazard.CreatedAt = s.TickCount
		s.Hazards = append(s.Hazards, hazard)
	}

	// Update citizen pathfinding state
	for i, citizen := range s.Citizens {

		// Check if any path intersects with hazards and needs recalculating
		for _, pos := range citizen.CurrentPath {
			if s.Grid.Cells[pos.Y][pos.X] == pf.CellHazard {
				citizen.updatePath(*s.Grid, citizen.CurrentPath[citizen.CurrentPathIndex], s.SafeZone, 0)
				break
			}
		}

		// Increment citizen's position on path
		if citizen.CurrentPathIndex < len(citizen.CurrentPath)-1 {
			s.Citizens[i].Status = CitizenNavigating
			s.Citizens[i].CurrentPathIndex++

			log.Printf("Citizen %v now at %v of %v steps", i, s.Citizens[i].CurrentPathIndex, len(citizen.CurrentPath))
		}
	}

	s.TickCount++
}
