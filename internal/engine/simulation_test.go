package engine

import (
	pf "hazard/internal/pathfinding"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewSimulation_InitializesCoreState(t *testing.T) {
	config := SimulationConfig{
		TickIntervalMs:    100,
		Width:             8,
		Height:            6,
		CitizenCountRange: [2]int{3, 3},
		SafeZone: SafeZoneConfig{
			Probability: 0,
			CountRange:  [2]int{1, 1},
			RadiusRange: [2]int{0, 0},
		},
		Hazard: HazardConfig{
			Probability:   0,
			CountRange:    [2]int{0, 0},
			DurationRange: [2]int{1, 1},
		},
	}

	simulation, err := NewSimulation(config)

	require.Nil(t, err)
	require.Equal(t, SimulationCreated, simulation.State)
	require.Equal(t, uint64(0), simulation.TickCount)
	require.NotNil(t, simulation.Grid)
	require.Equal(t, config.Width, simulation.Grid.Width)
	require.Equal(t, config.Height, simulation.Grid.Height)
	require.Len(t, simulation.Citizens, 3)
	require.Len(t, simulation.SafeZones, 1)

	require.GreaterOrEqual(t, simulation.SafeZones[0].Position.X, 0)
	require.Less(t, simulation.SafeZones[0].Position.X, simulation.Grid.Width)
	require.GreaterOrEqual(t, simulation.SafeZones[0].Position.Y, 0)
	require.Less(t, simulation.SafeZones[0].Position.Y, simulation.Grid.Height)

	for _, citizen := range simulation.Citizens {
		require.Equal(t, CitizenIdle, citizen.Status)
		require.Equal(t, 0, citizen.CurrentPathIndex)
		require.NotEmpty(t, citizen.Path)
		require.Equal(t, simulation.SafeZones[0].Position, citizen.Path[len(citizen.Path)-1])
	}
}

func TestTick_AdvancesCitizenOneStepPerTick(t *testing.T) {
	grid := pf.NewGrid(3, 1, pf.CellOpen)
	simulation := Simulation{
		Grid: &grid,
		Citizens: []Citizen{
			{
				Status: CitizenIdle,
				Path: []pf.Position{
					{X: 0, Y: 0},
					{X: 1, Y: 0},
					{X: 2, Y: 0},
				},
				CurrentPathIndex: 0,
			},
		},
	}

	simulation.Tick()
	require.Equal(t, uint64(1), simulation.TickCount)
	require.Equal(t, SimulationRunning, simulation.State)
	require.Equal(t, 1, simulation.Citizens[0].CurrentPathIndex)
	require.Equal(t, CitizenNavigating, simulation.Citizens[0].Status)

	simulation.Tick()
	require.Equal(t, uint64(2), simulation.TickCount)
	require.Equal(t, 2, simulation.Citizens[0].CurrentPathIndex)
}

func TestTick_CitizenStopsAtGoal(t *testing.T) {
	grid := pf.NewGrid(2, 1, pf.CellOpen)
	simulation := Simulation{
		Grid: &grid,
		Citizens: []Citizen{
			{
				Status: CitizenIdle,
				Path: []pf.Position{
					{X: 0, Y: 0},
					{X: 1, Y: 0},
				},
				CurrentPathIndex: 1,
			},
		},
	}

	simulation.Tick()

	require.Equal(t, 1, simulation.Citizens[0].CurrentPathIndex, "citizen should remain at goal index")
}

func TestTick_MultipleCitizensMoveIndependently(t *testing.T) {
	grid := pf.NewGrid(4, 2, pf.CellOpen)
	simulation := Simulation{
		Grid: &grid,
		Citizens: []Citizen{
			{
				Status: CitizenIdle,
				Path: []pf.Position{
					{X: 0, Y: 0},
					{X: 1, Y: 0},
				},
				CurrentPathIndex: 0,
			},
			{
				Status: CitizenIdle,
				Path: []pf.Position{
					{X: 0, Y: 1},
					{X: 1, Y: 1},
					{X: 2, Y: 1},
					{X: 3, Y: 1},
				},
				CurrentPathIndex: 0,
			},
		},
	}

	simulation.Tick()
	require.Equal(t, 1, simulation.Citizens[0].CurrentPathIndex)
	require.Equal(t, 1, simulation.Citizens[1].CurrentPathIndex)

	simulation.Tick()
	require.Equal(t, 1, simulation.Citizens[0].CurrentPathIndex, "first citizen should be at goal and stop")
	require.Equal(t, 2, simulation.Citizens[1].CurrentPathIndex, "second citizen should continue moving")
}

func TestTick_TransitionsCreatedToRunning(t *testing.T) {
	simulation := Simulation{
		State:     SimulationCreated,
		TickCount: 0,
		Citizens:  []Citizen{},
	}

	simulation.Tick()

	require.Equal(t, SimulationRunning, simulation.State)
	require.Equal(t, uint64(1), simulation.TickCount)
}

func TestTick_DoesNothingWhenPaused(t *testing.T) {
	simulation := Simulation{
		State: SimulationPaused,
		Citizens: []Citizen{
			{
				Status: CitizenIdle,
				Path: []pf.Position{
					{X: 0, Y: 0},
					{X: 1, Y: 0},
				},
				CurrentPathIndex: 0,
			},
		},
	}

	simulation.Tick()

	require.Equal(t, SimulationPaused, simulation.State)
	require.Equal(t, uint64(0), simulation.TickCount)
	require.Equal(t, 0, simulation.Citizens[0].CurrentPathIndex)
	require.Equal(t, CitizenIdle, simulation.Citizens[0].Status)
}

func TestTick_DoesNothingWhenCompleted(t *testing.T) {
	simulation := Simulation{
		State: SimulationCompleted,
		Citizens: []Citizen{
			{
				Status: CitizenIdle,
				Path: []pf.Position{
					{X: 0, Y: 0},
					{X: 1, Y: 0},
				},
				CurrentPathIndex: 0,
			},
		},
	}

	simulation.Tick()

	require.Equal(t, SimulationCompleted, simulation.State)
	require.Equal(t, uint64(0), simulation.TickCount)
	require.Equal(t, 0, simulation.Citizens[0].CurrentPathIndex)
	require.Equal(t, CitizenIdle, simulation.Citizens[0].Status)
}

func TestTick_CitizenReachesSafeZoneAndEscapes(t *testing.T) {
	grid := pf.NewGrid(3, 1, pf.CellOpen)
	grid.UpdateCell(pf.Position{X: 2, Y: 0}, pf.CellSafeZone)

	sim := Simulation{
		Grid: &grid,
		Citizens: []Citizen{
			{
				Status:          CitizenIdle,
				CurrentPosition: pf.Position{X: 0, Y: 0},
				Path: []pf.Position{
					{X: 0, Y: 0},
					{X: 1, Y: 0},
					{X: 2, Y: 0},
				},
				CurrentPathIndex: 0,
			},
		},
	}

	sim.Tick()
	require.Equal(t, CitizenNavigating, sim.Citizens[0].Status)
	require.Equal(t, pf.Position{X: 1, Y: 0}, sim.Citizens[0].CurrentPosition)
	require.Equal(t, 0, sim.EscapedCitizensCount)

	sim.Tick()
	require.Equal(t, CitizenEscaped, sim.Citizens[0].Status)
	require.Equal(t, pf.Position{X: 2, Y: 0}, sim.Citizens[0].CurrentPosition)
	require.Equal(t, 1, sim.EscapedCitizensCount)
	require.Equal(t, pf.CellSafeZone, sim.Grid.GetCell(sim.Citizens[0].CurrentPosition))
}

func TestTick_CitizenOvertakenByHazardDies(t *testing.T) {
	grid := pf.NewGrid(3, 1, pf.CellOpen)
	grid.UpdateCell(pf.Position{X: 0, Y: 0}, pf.CellHazard)

	sim := Simulation{
		Grid: &grid,
		Citizens: []Citizen{
			{
				Status:          CitizenIdle,
				CurrentPosition: pf.Position{X: 0, Y: 0},
				Path: []pf.Position{
					{X: 0, Y: 0},
					{X: 1, Y: 0},
				},
				CurrentPathIndex: 0,
			},
		},
	}

	sim.Tick()
	require.Equal(t, CitizenDead, sim.Citizens[0].Status)
	require.Equal(t, pf.Position{X: 0, Y: 0}, sim.Citizens[0].CurrentPosition)
	require.Equal(t, 1, sim.DeadCitizensCount)
}

func TestTick_NewSafeZoneAppearsOnSchedule(t *testing.T) {
	config := SimulationConfig{
		Width:             10,
		Height:            10,
		CitizenCountRange: [2]int{1, 1},
		SafeZone: SafeZoneConfig{
			Probability: 1.0,
			CountRange:  [2]int{2, 2},
			RadiusRange: [2]int{1, 1},
		},
		Hazard: HazardConfig{
			Probability:   0,
			CountRange:    [2]int{0, 0},
			DurationRange: [2]int{1, 1},
		},
	}

	sim, err := NewSimulation(config)
	require.NoError(t, err)
	require.Len(t, sim.SafeZones, 1)

	sim.Tick()

	require.Len(t, sim.SafeZones, 2)
	for _, sz := range sim.SafeZones {
		require.True(t, sim.Grid.InBounds(sz.Position))
	}
}

func TestTick_CitizensRecalculateTowardNearestZoneAfterEmergence(t *testing.T) {
	grid := pf.NewGrid(10, 10, pf.CellOpen)

	// Place initial safe zone at far corner
	grid.UpdateCell(pf.Position{X: 9, Y: 9}, pf.CellSafeZone)

	sim := Simulation{
		Config: SimulationConfig{
			SafeZone: SafeZoneConfig{
				Probability: 1.0,
				CountRange:  [2]int{2, 2},
				RadiusRange: [2]int{1, 1},
			},
			Hazard: HazardConfig{
				Probability:   0,
				CountRange:    [2]int{0, 0},
				DurationRange: [2]int{1, 1},
			},
		},
		Grid:         &grid,
		MaxSafeZones: 2,
		SafeZones: []SafeZone{
			{Position: pf.Position{X: 9, Y: 9}, Radius: 1},
		},
		Citizens: []Citizen{
			{
				ID:               0,
				Status:           CitizenIdle,
				CurrentPosition:  pf.Position{X: 0, Y: 0},
				Path:             []pf.Position{{X: 0, Y: 0}},
				CurrentPathIndex: 0,
			},
		},
	}

	sim.Tick()

	require.Len(t, sim.SafeZones, 2)

	dest := sim.Citizens[0].Path[len(sim.Citizens[0].Path)-1]
	require.Equal(t, pf.CellSafeZone, sim.Grid.GetCell(dest),
		"citizen must recalculate path toward nearest safe zone after emergence")
}

func TestTick_SimulationCompletesWhenAllResolved(t *testing.T) {
	grid := pf.NewGrid(3, 1, pf.CellOpen)

	sim := Simulation{
		Grid:              &grid,
		DeadCitizensCount: 1,
		Citizens: []Citizen{
			{
				Status: CitizenDead,
			},
			{
				Status:          CitizenIdle,
				CurrentPosition: pf.Position{X: 0, Y: 0},
				Path: []pf.Position{
					{X: 0, Y: 0},
					{X: 1, Y: 0},
				},
				CurrentPathIndex: 0,
			},
		},
	}

	sim.Tick()

	require.Equal(t, CitizenEscaped, sim.Citizens[1].Status)
	require.Equal(t, 1, sim.EscapedCitizensCount)
	require.Equal(t, 1, sim.DeadCitizensCount)
	require.Equal(t, SimulationCompleted, sim.State)
}
