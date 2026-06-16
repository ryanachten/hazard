package engine

import (
	pf "hazard/internal/pathfinding"
	"log"
	"math/rand"
)

// Simulation engine for hazards
type Simulation struct {
	State     SimulationState
	TickCount uint64
	Citizens  []Citizen
	SafeZone  pf.Position
	Grid      *pf.Grid
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

	citizenMin := config.CitizenCountRange[0]
	citizenMax := config.CitizenCountRange[1]
	citizenCount := citizenMin + rand.Intn(citizenMax-citizenMin+1)
	citizens := make([]Citizen, citizenCount)

	pathfinder := pf.AStar{}

	for i := range citizens {
		startPosition := grid.GetRandomPosition()
		path, err := pathfinder.FindPath(grid, startPosition, safeZone)
		if err != nil {
			log.Printf("Error creating path for citizen %v: %v", i, err)
			citizens[i] = Citizen{
				Status:           CitizenIdle,
				CurrentPath:      []pf.Position{startPosition},
				CurrentPathIndex: 0,
			}
			continue
		}

		citizens[i] = Citizen{
			Status:           CitizenIdle,
			CurrentPath:      path,
			CurrentPathIndex: 0,
		}
	}

	return Simulation{
		State:     SimulationCreated,
		TickCount: 0,
		Citizens:  citizens,
		Grid:      &grid,
		SafeZone:  safeZone,
	}
}

// Tick increments a simulation by one tick
func (s *Simulation) Tick() {
	if s.State == SimulationPaused || s.State == SimulationCompleted {
		return
	}

	s.State = SimulationRunning

	for i, citizen := range s.Citizens {
		if citizen.CurrentPathIndex < len(citizen.CurrentPath)-1 {
			s.Citizens[i].Status = CitizenNavigating
			s.Citizens[i].CurrentPathIndex++

			log.Printf("Citizen %v now at %v of %v steps", i, s.Citizens[i].CurrentPathIndex, len(citizen.CurrentPath))
		}
	}

	s.TickCount++
}
