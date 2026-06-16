package engine

import (
	pf "hazard/internal/pathfinding"
	"log"
	"math/rand"
)

// Simulation engine for hazards
type Simulation struct {
	TickCount uint64
	Citizens  []Citizen
	SafeZone  pf.Position
	Grid      *pf.Grid
}

type SimulationConfig struct {
	TickIntervalMs    int
	Width             int
	Height            int
	CitizenCountRange [2]int
}

func NewSimulation(config SimulationConfig) Simulation {

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
			continue
		}

		citizens[i] = Citizen{
			Status:           CitizenIdle,
			CurrentPath:      path,
			CurrentPathIndex: 0,
		}
	}

	return Simulation{
		TickCount: 0,
		Citizens:  citizens,
		Grid:      &grid,
		SafeZone:  safeZone,
	}
}

func (s *Simulation) Tick() {
	for i, citizen := range s.Citizens {
		if citizen.CurrentPathIndex < len(citizen.CurrentPath) {
			s.Citizens[i].Status = CitizenNavigating
			s.Citizens[i].CurrentPathIndex++

			log.Printf("Citizen %v now at %v of %v steps", i, s.Citizens[i].CurrentPathIndex, len(citizen.CurrentPath))
		}
	}

	s.TickCount++
}
