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
	}

	simulation, err := NewSimulation(config)

	require.Nil(t, err)
	require.Equal(t, SimulationCreated, simulation.State)
	require.Equal(t, uint64(0), simulation.TickCount)
	require.NotNil(t, simulation.Grid)
	require.Equal(t, config.Width, simulation.Grid.Width)
	require.Equal(t, config.Height, simulation.Grid.Height)
	require.Len(t, simulation.Citizens, 3)

	require.GreaterOrEqual(t, simulation.SafeZone.X, 0)
	require.Less(t, simulation.SafeZone.X, simulation.Grid.Width)
	require.GreaterOrEqual(t, simulation.SafeZone.Y, 0)
	require.Less(t, simulation.SafeZone.Y, simulation.Grid.Height)

	for _, citizen := range simulation.Citizens {
		require.Equal(t, CitizenIdle, citizen.Status)
		require.Equal(t, 0, citizen.CurrentPathIndex)
		require.NotEmpty(t, citizen.Path)
		require.Equal(t, simulation.SafeZone, citizen.Path[len(citizen.Path)-1])
	}
}

func TestTick_AdvancesCitizenOneStepPerTick(t *testing.T) {
	simulation := Simulation{
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
	simulation := Simulation{
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
	simulation := Simulation{
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
