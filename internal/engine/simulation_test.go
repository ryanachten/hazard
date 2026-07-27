package engine

import (
	c "hazard/internal/citizen"
	config "hazard/internal/configuration"
	"hazard/internal/events"
	h "hazard/internal/hazard"
	o "hazard/internal/obstacle"
	pf "hazard/internal/pathfinding"
	r "hazard/internal/ranging"
	sz "hazard/internal/safe_zone"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestProcessCommand_PauseRunningSimulation(t *testing.T) {
	sim := Simulation{
		State: SimulationRunning,
	}

	sim.ProcessCommand(events.SimulationCommand{
		CommandType: events.PauseSimulation,
	})

	require.Equal(t, SimulationPaused, sim.State)
}

func TestProcessCommand_ResumePausedSimulation(t *testing.T) {
	sim := Simulation{
		State: SimulationPaused,
	}

	sim.ProcessCommand(events.SimulationCommand{
		CommandType: events.PauseSimulation,
	})

	require.Equal(t, SimulationRunning, sim.State)
}

func TestProcessCommand_ToggleFromCreated(t *testing.T) {
	sim := Simulation{
		State: SimulationCreated,
	}

	sim.ProcessCommand(events.SimulationCommand{
		CommandType: events.PauseSimulation,
	})

	require.Equal(t, SimulationRunning, sim.State,
		"non-running states should transition to running")
}

func TestProcessCommand_ToggleFromCompleted(t *testing.T) {
	sim := Simulation{
		State: SimulationCompleted,
	}

	sim.ProcessCommand(events.SimulationCommand{
		CommandType: events.PauseSimulation,
	})

	require.Equal(t, SimulationRunning, sim.State,
		"completed simulation should toggle to running")
}

func TestProcessCommand_UnknownCommandTypeDoesNothing(t *testing.T) {
	sim := Simulation{
		State: SimulationRunning,
	}

	sim.ProcessCommand(events.SimulationCommand{
		CommandType: "unknown.command",
	})

	require.Equal(t, SimulationRunning, sim.State,
		"unknown command type must not change state")
}

func TestTick_ObstaclesBlockPathfinding(t *testing.T) {
	// Obstacles placed at simulation start must block pathfinding,
	// forcing citizens to route around them.
	cfg := config.SimulationConfig{
		CitizenCount: 1,
		SafeZone: sz.Config{
			Probability: 0,
			Count:       1,
			RadiusRange: r.Range{Min: 1, Max: 1},
		},
		Hazard: h.Config{
			Probability:   0,
			Count:         0,
			DurationRange: r.PositiveRange{Min: 1, Max: 1},
		},
		Obstacle: o.Config{
			CountRange: r.Range{Min: 0, Max: 0},
			SizeRange:  r.PositiveRange{Min: 1, Max: 1},
		},
	}

	sim, err := NewSimulation(10, 10, cfg, events.New())
	require.NoError(t, err)

	// Grid should have safe zone cells marked
	safeZoneCells := sim.SafeZones[0].Cells
	for _, cell := range safeZoneCells {
		require.Equal(t, pf.CellSafeZone, sim.Grid.GetCell(cell))
	}

	// Citizen is spawned with a path that ends in a safe zone cell
	require.NotEmpty(t, sim.Citizens[0].Path)
	lastCell := sim.Citizens[0].Path[len(sim.Citizens[0].Path)-1]
	require.Equal(t, pf.CellSafeZone, sim.Grid.GetCell(lastCell))
	// All open path cells must not be obstacle or hazard
	for _, pos := range sim.Citizens[0].Path {
		cellType := sim.Grid.GetCell(pos)
		require.NotEqual(t, pf.CellObstacle, cellType,
			"citizen path must not pass through obstacle cell at %v", pos)
	}
}

func TestNewSimulation_InitializesCoreState(t *testing.T) {
	cfg := config.SimulationConfig{
		TickIntervalMs: 100,
		CitizenCount:   3,
		SafeZone: sz.Config{
			Probability: 0,
			Count:       1,
			RadiusRange: r.Range{Min: 0, Max: 0},
		},
		Hazard: h.Config{
			Probability:   0,
			Count:         0,
			DurationRange: r.PositiveRange{Min: 1, Max: 1},
		},
	}

	simulation, err := NewSimulation(8, 6, cfg, events.New())

	require.Nil(t, err)
	require.Equal(t, SimulationRunning, simulation.State)
	require.Equal(t, uint64(0), simulation.TickCount)
	require.NotNil(t, simulation.Grid)
	require.Equal(t, 8, simulation.Grid.Width)
	require.Equal(t, 6, simulation.Grid.Height)
	require.Len(t, simulation.Citizens, 3)
	require.Len(t, simulation.SafeZones, 1)

	require.GreaterOrEqual(t, simulation.SafeZones[0].Position.X, 0)
	require.Less(t, simulation.SafeZones[0].Position.X, simulation.Grid.Width)
	require.GreaterOrEqual(t, simulation.SafeZones[0].Position.Y, 0)
	require.Less(t, simulation.SafeZones[0].Position.Y, simulation.Grid.Height)

	for _, citizen := range simulation.Citizens {
		require.Equal(t, c.StatusIdle, citizen.Status)
		require.Equal(t, 0, citizen.CurrentPathIndex)
		require.NotEmpty(t, citizen.Path)
		require.Equal(t, simulation.SafeZones[0].Position, citizen.Path[len(citizen.Path)-1])
	}
}

func TestTick_AdvancesCitizenOneStepPerTick(t *testing.T) {
	grid := pf.NewGrid(3, 1, pf.CellOpen)

	safeZone := sz.SafeZone{
		ID:          uuid.New(),
		Cells:       []pf.Position{{X: 2, Y: 0}},
		HasCapacity: true,
	}
	safeZoneLocations := map[pf.Position]*sz.SafeZone{
		{X: 2, Y: 0}: &safeZone,
	}

	simulation := Simulation{
		Grid:              &grid,
		eventBus:          events.New(),
		safeZoneLocations: safeZoneLocations,
		Citizens: []c.Citizen{
			{
				Status: c.StatusIdle,
				Path: []pf.Position{
					{X: 0, Y: 0},
					{X: 1, Y: 0},
					{X: 2, Y: 0},
				},
				CurrentPathIndex: 0,
				TargetSafeZone:   &safeZone,
			},
		},
	}

	simulation.Tick()
	require.Equal(t, uint64(1), simulation.TickCount)
	require.Equal(t, SimulationRunning, simulation.State)
	require.Equal(t, 1, simulation.Citizens[0].CurrentPathIndex)
	require.Equal(t, c.StatusNavigating, simulation.Citizens[0].Status)

	simulation.Tick()
	require.Equal(t, uint64(2), simulation.TickCount)
	require.Equal(t, 2, simulation.Citizens[0].CurrentPathIndex)
}

func TestTick_CitizenStopsAtGoal(t *testing.T) {
	grid := pf.NewGrid(2, 1, pf.CellOpen)
	simulation := Simulation{
		Grid:     &grid,
		eventBus: events.New(),
		Citizens: []c.Citizen{
			{
				Status: c.StatusIdle,
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

	safeZone := sz.SafeZone{
		ID:          uuid.New(),
		Cells:       []pf.Position{{X: 1, Y: 0}, {X: 1, Y: 1}},
		HasCapacity: true,
	}
	safeZoneLocations := map[pf.Position]*sz.SafeZone{
		{X: 1, Y: 0}: &safeZone,
		{X: 1, Y: 1}: &safeZone,
	}

	simulation := Simulation{
		Grid:              &grid,
		eventBus:          events.New(),
		safeZoneLocations: safeZoneLocations,
		Citizens: []c.Citizen{
			{
				Status: c.StatusIdle,
				Path: []pf.Position{
					{X: 0, Y: 0},
					{X: 1, Y: 0},
				},
				CurrentPathIndex: 0,
				TargetSafeZone:   &safeZone,
			},
			{
				Status: c.StatusIdle,
				Path: []pf.Position{
					{X: 0, Y: 1},
					{X: 1, Y: 1},
					{X: 2, Y: 1},
					{X: 3, Y: 1},
				},
				CurrentPathIndex: 0,
				TargetSafeZone:   &safeZone,
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
		Citizens:  []c.Citizen{},
	}

	simulation.Tick()

	require.Equal(t, SimulationRunning, simulation.State)
	require.Equal(t, uint64(1), simulation.TickCount)
}

func TestTick_DoesNothingWhenPaused(t *testing.T) {
	simulation := Simulation{
		State: SimulationPaused,
		Citizens: []c.Citizen{
			{
				Status: c.StatusIdle,
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
	require.Equal(t, c.StatusIdle, simulation.Citizens[0].Status)
}

func TestTick_DoesNothingWhenCompleted(t *testing.T) {
	simulation := Simulation{
		State: SimulationCompleted,
		Citizens: []c.Citizen{
			{
				Status: c.StatusIdle,
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
	require.Equal(t, c.StatusIdle, simulation.Citizens[0].Status)
}

func TestTick_CitizenReachesSafeZoneAndEscapes(t *testing.T) {
	grid := pf.NewGrid(3, 1, pf.CellOpen)
	grid.UpdateCell(pf.Position{X: 2, Y: 0}, pf.CellSafeZone)

	safeZone := sz.SafeZone{
		ID:          uuid.New(),
		Position:    pf.Position{X: 2, Y: 0},
		Radius:      0,
		Cells:       []pf.Position{{X: 2, Y: 0}},
		HasCapacity: true,
	}
	safeZoneLocations := map[pf.Position]*sz.SafeZone{
		{X: 2, Y: 0}: &safeZone,
	}

	sim := Simulation{
		Grid:              &grid,
		eventBus:          events.New(),
		safeZoneLocations: safeZoneLocations,
		Citizens: []c.Citizen{
			{
				Status:          c.StatusIdle,
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
	require.Equal(t, c.StatusNavigating, sim.Citizens[0].Status)
	require.Equal(t, pf.Position{X: 1, Y: 0}, sim.Citizens[0].CurrentPosition)
	require.Equal(t, 0, sim.EscapedCitizensCount)

	sim.Tick()
	require.Equal(t, c.StatusEscaped, sim.Citizens[0].Status)
	require.Equal(t, pf.Position{X: 2, Y: 0}, sim.Citizens[0].CurrentPosition)
	require.Equal(t, 1, sim.EscapedCitizensCount)
	require.Equal(t, pf.CellEscapedCitizen, sim.Grid.GetCell(sim.Citizens[0].CurrentPosition))
}

func TestTick_CitizenOvertakenByHazardDies(t *testing.T) {
	grid := pf.NewGrid(3, 1, pf.CellOpen)
	grid.UpdateCell(pf.Position{X: 0, Y: 0}, pf.CellHazard)

	sim := Simulation{
		Grid:     &grid,
		eventBus: events.New(),
		Citizens: []c.Citizen{
			{
				Status:          c.StatusIdle,
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
	require.Equal(t, c.StatusDead, sim.Citizens[0].Status)
	require.Equal(t, pf.Position{X: 0, Y: 0}, sim.Citizens[0].CurrentPosition)
	require.Equal(t, 1, sim.DeadCitizensCount)
}

func TestTick_NewSafeZoneAppearsOnSchedule(t *testing.T) {
	cfg := config.SimulationConfig{
		CitizenCount: 1,
		SafeZone: sz.Config{
			Probability: 1.0,
			Count:       2,
			RadiusRange: r.Range{Min: 1, Max: 1},
		},
		Hazard: h.Config{
			Probability:   0,
			Count:         0,
			DurationRange: r.PositiveRange{Min: 1, Max: 1},
		},
	}

	sim, err := NewSimulation(10, 10, cfg, events.New())
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

	initialSZ := sz.SafeZone{
		Position:    pf.Position{X: 9, Y: 9},
		Radius:      1,
		HasCapacity: true,
	}
	safeZoneLocations := map[pf.Position]*sz.SafeZone{
		{X: 9, Y: 9}: &initialSZ,
	}

	sim := Simulation{
		Config: config.SimulationConfig{
			SafeZone: sz.Config{
				Probability: 1.0,
				Count:       2,
				RadiusRange: r.Range{Min: 1, Max: 1},
			},
			Hazard: h.Config{
				Probability:   0,
				Count:         0,
				DurationRange: r.PositiveRange{Min: 1, Max: 1},
			},
		},
		Grid:              &grid,
		eventBus:          events.New(),
		safeZoneLocations: safeZoneLocations,
		SafeZones: []sz.SafeZone{
			{Position: pf.Position{X: 9, Y: 9}, Radius: 1},
		},
		Citizens: []c.Citizen{
			{
				ID:               uuid.New(),
				Status:           c.StatusIdle,
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

func TestTick_SafeZoneCapacityPreventsOverfilling(t *testing.T) {
	// Two safe zones: sz1 has capacity 2 (two cells), sz2 has capacity 1 (one cell).
	// Two citizens target sz1, filling it. A third citizen arriving later
	// and targeting sz1 should recalculate toward sz2.
	sz1Cells := []pf.Position{{X: 4, Y: 0}, {X: 4, Y: 1}}
	sz2Cells := []pf.Position{{X: 3, Y: 0}}

	grid := pf.NewGrid(5, 2, pf.CellOpen)
	for _, c := range sz1Cells {
		grid.UpdateCell(c, pf.CellSafeZone)
	}
	for _, c := range sz2Cells {
		grid.UpdateCell(c, pf.CellSafeZone)
	}

	sz1 := sz.SafeZone{
		ID:          uuid.New(),
		Position:    pf.Position{X: 4, Y: 0},
		Radius:      0,
		Cells:       sz1Cells,
		HasCapacity: true,
		Occupants:   []uuid.UUID{},
	}
	sz2 := sz.SafeZone{
		ID:          uuid.New(),
		Position:    pf.Position{X: 3, Y: 0},
		Radius:      0,
		Cells:       sz2Cells,
		HasCapacity: true,
		Occupants:   []uuid.UUID{},
	}

	safeZoneLocations := map[pf.Position]*sz.SafeZone{
		{X: 4, Y: 0}: &sz1,
		{X: 4, Y: 1}: &sz1,
		{X: 3, Y: 0}: &sz2,
	}

	sim := Simulation{
		Grid:              &grid,
		eventBus:          events.New(),
		safeZoneLocations: safeZoneLocations,
		SafeZones:         []sz.SafeZone{sz1, sz2},
		Citizens: []c.Citizen{
			{
				ID:               uuid.New(),
				Status:           c.StatusIdle,
				CurrentPosition:  pf.Position{X: 0, Y: 0},
				Path:             []pf.Position{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}, {X: 3, Y: 0}, {X: 4, Y: 0}},
				CurrentPathIndex: 0,
				TargetSafeZone:   &sz1,
			},
			{
				ID:               uuid.New(),
				Status:           c.StatusIdle,
				CurrentPosition:  pf.Position{X: 0, Y: 1},
				Path:             []pf.Position{{X: 0, Y: 1}, {X: 1, Y: 1}, {X: 2, Y: 1}, {X: 3, Y: 1}, {X: 4, Y: 1}},
				CurrentPathIndex: 0,
				TargetSafeZone:   &sz1,
			},
		},
	}

	// Both citizens move toward sz1
	sim.Tick()
	sim.Tick()
	sim.Tick()
	sim.Tick()

	require.Equal(t, c.StatusEscaped, sim.Citizens[0].Status,
		"first citizen should have escaped into sz1")
	require.Equal(t, c.StatusEscaped, sim.Citizens[1].Status,
		"second citizen should have escaped into sz1")
	require.Equal(t, 2, sim.EscapedCitizensCount)
	require.False(t, safeZoneLocations[pf.Position{X: 4, Y: 0}].HasCapacity,
		"sz1 should be at capacity after two occupants")

	// Add a third citizen whose path goes through sz1's cells.
	// Since sz1 is full, they should recalculate toward sz2.
	grid.UpdateCell(pf.Position{X: 0, Y: 0}, pf.CellOpen)

	sim.Citizens = append(sim.Citizens, c.Citizen{
		ID:               uuid.New(),
		Status:           c.StatusIdle,
		CurrentPosition:  pf.Position{X: 0, Y: 0},
		Path:             []pf.Position{{X: 0, Y: 0}},
		CurrentPathIndex: 0,
		TargetSafeZone:   safeZoneLocations[pf.Position{X: 4, Y: 0}],
	})
	sim.State = SimulationRunning

	sim.Tick()

	lastCell := sim.Citizens[2].Path[len(sim.Citizens[2].Path)-1]
	require.NotEqual(t, safeZoneLocations[pf.Position{X: 4, Y: 0}], sim.Citizens[2].TargetSafeZone,
		"citizen with target at capacity must recalculate to a different safe zone")
	require.Equal(t, pf.Position{X: 3, Y: 0}, lastCell,
		"recalculated path must lead to the available safe zone (sz2)")
}

func TestTick_SimulationCompletesWhenAllResolved(t *testing.T) {
	grid := pf.NewGrid(3, 1, pf.CellOpen)

	safeZone := sz.SafeZone{
		ID:          uuid.New(),
		Cells:       []pf.Position{{X: 1, Y: 0}},
		HasCapacity: true,
	}
	safeZoneLocations := map[pf.Position]*sz.SafeZone{
		{X: 1, Y: 0}: &safeZone,
	}

	sim := Simulation{
		Grid:              &grid,
		eventBus:          events.New(),
		safeZoneLocations: safeZoneLocations,
		DeadCitizensCount: 1,
		Citizens: []c.Citizen{
			{
				Status: c.StatusDead,
			},
			{
				Status:          c.StatusIdle,
				CurrentPosition: pf.Position{X: 0, Y: 0},
				Path: []pf.Position{
					{X: 0, Y: 0},
					{X: 1, Y: 0},
				},
				CurrentPathIndex: 0,
				TargetSafeZone:   &safeZone,
			},
		},
	}

	sim.Tick()

	require.Equal(t, c.StatusEscaped, sim.Citizens[1].Status)
	require.Equal(t, 1, sim.EscapedCitizensCount)
	require.Equal(t, 1, sim.DeadCitizensCount)
	require.Equal(t, SimulationCompleted, sim.State)
}
