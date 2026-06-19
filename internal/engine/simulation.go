package engine

import (
	pf "hazard/internal/pathfinding"
	"log"
	"math/rand"
)

// Simulation engine for hazards
type Simulation struct {
	State        SimulationState
	TickCount    uint64
	Citizens     []Citizen
	SafeZone     pf.Position
	Grid         *pf.Grid
	Hazards      []Hazard
	HazardConfig hazardConfig
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
	Hazard            hazardConfig
}

type hazardConfig struct {
	HazardDurationRange [2]int
	HazardProbability   float32
	MaxHazards          int
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

	// Randomly generate hazards within hazard limits
	if len(s.Hazards) < s.HazardConfig.MaxHazards && rand.Float32() <= s.HazardConfig.HazardProbability {
		s.Hazards = append(s.Hazards, createHazard(s.HazardConfig, *s.Grid))
		// TODO: update citizen paths now that the path contains more obstacles?
	}

	// Update citizen pathfinding state
	for i, citizen := range s.Citizens {
		if citizen.CurrentPathIndex < len(citizen.CurrentPath)-1 {
			s.Citizens[i].Status = CitizenNavigating
			s.Citizens[i].CurrentPathIndex++

			// TODO: when citizen steps into hazard cell, they should die
			log.Printf("Citizen %v now at %v of %v steps", i, s.Citizens[i].CurrentPathIndex, len(citizen.CurrentPath))
		}
	}

	s.TickCount++
}
