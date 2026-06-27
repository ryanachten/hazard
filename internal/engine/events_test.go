package engine

import (
	"hazard/internal/events"
	pf "hazard/internal/pathfinding"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestNewSimulation_EventLogIsEmpty(t *testing.T) {
	config := SimulationConfig{
		Width:             5,
		Height:            5,
		CitizenCountRange: [2]int{0, 0},
		Hazard: HazardConfig{
			Probability:   0,
			CountRange:    [2]int{0, 0},
			DurationRange: [2]int{1, 1},
		},
		SafeZone: SafeZoneConfig{
			Probability: 0,
			CountRange:  [2]int{1, 1},
			RadiusRange: [2]int{1, 1},
		},
	}

	sim, err := NewSimulation(config)
	require.NoError(t, err)

	evts := sim.Events()
	require.Empty(t, evts, "new simulation should start with empty event log")
}

func TestTick_EmitsCitizenMovedEvent(t *testing.T) {
	grid := pf.NewGrid(3, 1, pf.CellOpen)
	sim := Simulation{
		Grid:         &grid,
		EventEmitter: &events.InMemoryEventLog{},
		Citizens: []Citizen{
			{
				Status:           CitizenIdle,
				CurrentPosition:  pf.Position{X: 0, Y: 0},
				Path:             []pf.Position{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}},
				CurrentPathIndex: 0,
			},
		},
	}

	sim.Tick()

	evts := sim.Events()
	foundMove := false
	for _, e := range evts {
		if string(e.EventType) == "citizen.moved" {
			foundMove = true
			require.Equal(t, sim.ID, e.Metadata.SimulationID)
			require.Equal(t, uint64(0), e.Metadata.Tick)
			break
		}
	}
	require.True(t, foundMove, "tick should produce a citizen.moved event")
}

func TestTick_EmitsCitizenEscapedEvent(t *testing.T) {
	grid := pf.NewGrid(3, 1, pf.CellOpen)
	grid.UpdateCell(pf.Position{X: 2, Y: 0}, pf.CellSafeZone)

	sim := Simulation{
		Grid:         &grid,
		EventEmitter: &events.InMemoryEventLog{},
		Citizens: []Citizen{
			{
				Status:           CitizenIdle,
				CurrentPosition:  pf.Position{X: 0, Y: 0},
				Path:             []pf.Position{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}},
				CurrentPathIndex: 0,
			},
		},
	}

	sim.Tick()
	sim.Tick()

	evts := sim.Events()
	foundEscaped := false
	for _, e := range evts {
		if string(e.EventType) == "citizen.escaped" {
			foundEscaped = true
			break
		}
	}
	require.True(t, foundEscaped, "citizen reaching safe zone should produce a citizen.escaped event")

	foundCompleted := false
	for _, e := range evts {
		if string(e.EventType) == "simulation.completed" {
			foundCompleted = true
			break
		}
	}
	require.True(t, foundCompleted, "all citizens resolved should produce a simulation.completed event")
}

func TestTick_EmitsCitizenDiedEvent(t *testing.T) {
	grid := pf.NewGrid(3, 1, pf.CellOpen)
	grid.UpdateCell(pf.Position{X: 0, Y: 0}, pf.CellHazard)

	sim := Simulation{
		Grid:         &grid,
		EventEmitter: &events.InMemoryEventLog{},
		Citizens: []Citizen{
			{
				Status:           CitizenIdle,
				CurrentPosition:  pf.Position{X: 0, Y: 0},
				Path:             []pf.Position{{X: 0, Y: 0}, {X: 1, Y: 0}},
				CurrentPathIndex: 0,
			},
		},
	}

	sim.Tick()

	evts := sim.Events()
	foundDied := false
	for _, e := range evts {
		if string(e.EventType) == "citizen.died" {
			foundDied = true
			break
		}
	}
	require.True(t, foundDied, "citizen on hazard cell should produce a citizen.died event")
}

func TestTick_EmitsHazardExpandedEvents(t *testing.T) {
	grid := pf.NewGrid(10, 10, pf.CellOpen)

	sim := Simulation{
		Grid:         &grid,
		EventEmitter: &events.InMemoryEventLog{},
		Hazards: []Hazard{
			{
				ID:            uuid.New(),
				CreatedAt:     0,
				Duration:      100,
				Origin:        pf.Position{X: 5, Y: 5},
				CurrentRadius: 0,
			},
		},
	}

	sim.Tick()

	evts := sim.Events()
	foundExpanded := false
	for _, e := range evts {
		if string(e.EventType) == "hazard.expanded" {
			foundExpanded = true
			require.Equal(t, uint64(0), e.Metadata.Tick, "hazard expansion event tick should match current tick")
			break
		}
	}
	require.True(t, foundExpanded, "existing hazard should produce hazard.expanded event each tick")
}

func TestTick_EmitsSimulationCompletedEvent(t *testing.T) {
	grid := pf.NewGrid(3, 1, pf.CellOpen)
	grid.UpdateCell(pf.Position{X: 2, Y: 0}, pf.CellSafeZone)

	sim := Simulation{
		Grid:         &grid,
		EventEmitter: &events.InMemoryEventLog{},
		Citizens: []Citizen{
			{
				Status:           CitizenIdle,
				CurrentPosition:  pf.Position{X: 0, Y: 0},
				Path:             []pf.Position{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}},
				CurrentPathIndex: 0,
			},
		},
	}

	sim.Tick()
	sim.Tick()

	evts := sim.Events()
	require.NotEmpty(t, evts)
	lastEvent := evts[len(evts)-1]
	require.Equal(t, "simulation.completed", string(lastEvent.EventType),
		"last event in a completed simulation should be simulation.completed")
	require.Equal(t, SimulationCompleted, sim.State)
}

func TestEventTicks_InAscendingOrder(t *testing.T) {
	grid := pf.NewGrid(5, 1, pf.CellOpen)
	grid.UpdateCell(pf.Position{X: 4, Y: 0}, pf.CellSafeZone)

	sim := Simulation{
		Grid:         &grid,
		EventEmitter: &events.InMemoryEventLog{},
		Citizens: []Citizen{
			{
				Status:           CitizenIdle,
				CurrentPosition:  pf.Position{X: 0, Y: 0},
				Path:             []pf.Position{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}, {X: 3, Y: 0}, {X: 4, Y: 0}},
				CurrentPathIndex: 0,
			},
		},
	}

	sim.Tick()
	sim.Tick()
	sim.Tick()
	sim.Tick()

	evts := sim.Events()
	require.NotEmpty(t, evts)

	var prevTick uint64
	for i, e := range evts {
		require.GreaterOrEqual(t, e.Metadata.Tick, prevTick,
			"event %d (type=%s) at tick %d is out of order with previous tick %d",
			i, string(e.EventType), e.Metadata.Tick, prevTick)
		prevTick = e.Metadata.Tick
	}
}

func TestSimulationEventsAccessor_AccumulatesEvents(t *testing.T) {
	grid := pf.NewGrid(3, 1, pf.CellOpen)

	sim := Simulation{
		Grid:         &grid,
		EventEmitter: &events.InMemoryEventLog{},
		Citizens: []Citizen{
			{
				Status:           CitizenIdle,
				CurrentPosition:  pf.Position{X: 0, Y: 0},
				Path:             []pf.Position{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}},
				CurrentPathIndex: 0,
			},
		},
	}

	require.Empty(t, sim.Events(), "event log should be empty before first tick")

	sim.Tick()
	tick1Count := len(sim.Events())
	require.NotZero(t, tick1Count, "first tick should produce events")

	sim.Tick()
	tick2Events := sim.Events()
	require.Greater(t, len(tick2Events), tick1Count,
		"events should accumulate across ticks - got %d after tick 1, %d after tick 2", tick1Count, len(tick2Events))
}

func TestTick_EmitsCitizenPathUpdatedOnRecalculation(t *testing.T) {
	grid := pf.NewGrid(5, 5, pf.CellOpen)
	destination := pf.Position{X: 4, Y: 4}

	grid.UpdateCell(pf.Position{X: 4, Y: 4}, pf.CellSafeZone)

	path := []pf.Position{
		{X: 0, Y: 0},
		{X: 0, Y: 1},
		{X: 0, Y: 2},
		{X: 0, Y: 3},
		{X: 0, Y: 4},
		{X: 1, Y: 4},
		{X: 2, Y: 4},
		{X: 3, Y: 4},
		{X: 4, Y: 4},
	}

	sim := Simulation{
		EventEmitter: &events.InMemoryEventLog{},
		Grid:         &grid,
		Citizens: []Citizen{
			{
				ID:                 uuid.New(),
				Status:             CitizenIdle,
				CurrentPosition:    pf.Position{X: 0, Y: 0},
				CurrentDestination: destination,
				Path:               path,
				CurrentPathIndex:   0,
			},
		},
	}

	blockedCell := pf.Position{X: 0, Y: 2}
	grid.UpdateCell(blockedCell, pf.CellHazard)

	sim.Tick()

	evts := sim.Events()
	foundPathUpdated := false
	for _, e := range evts {
		if string(e.EventType) == "citizen.pathUpdated" {
			foundPathUpdated = true
			break
		}
	}
	require.True(t, foundPathUpdated, "path blocked by hazard should produce citizen.pathUpdated event")
}

func TestCompletedSimulation_HasCompleteEventChain(t *testing.T) {
	grid := pf.NewGrid(5, 1, pf.CellOpen)
	grid.UpdateCell(pf.Position{X: 4, Y: 0}, pf.CellSafeZone)

	sim := Simulation{
		ID:           uuid.New(),
		Grid:         &grid,
		EventEmitter: &events.InMemoryEventLog{},
		Citizens: []Citizen{
			{
				ID:               uuid.New(),
				Status:           CitizenIdle,
				CurrentPosition:  pf.Position{X: 0, Y: 0},
				Path:             []pf.Position{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}, {X: 3, Y: 0}, {X: 4, Y: 0}},
				CurrentPathIndex: 0,
			},
		},
	}

	sim.Tick()
	sim.Tick()
	sim.Tick()
	sim.Tick()

	evts := sim.Events()
	require.NotEmpty(t, evts)

	hasMoved := false
	hasEscaped := false
	for _, e := range evts {
		switch string(e.EventType) {
		case "citizen.moved":
			hasMoved = true
		case "citizen.escaped":
			hasEscaped = true
		}
	}
	require.True(t, hasMoved, "completed simulation should have citizen.moved events")
	require.True(t, hasEscaped, "completed simulation should have citizen.escaped event")

	lastEvent := evts[len(evts)-1]
	require.Equal(t, "simulation.completed", string(lastEvent.EventType),
		"the final event of a completed simulation must be simulation.completed")
}

func TestMultipleCitizens_ProduceIndependentEvents(t *testing.T) {
	grid := pf.NewGrid(5, 2, pf.CellOpen)
	grid.UpdateCell(pf.Position{X: 4, Y: 0}, pf.CellSafeZone)
	grid.UpdateCell(pf.Position{X: 4, Y: 1}, pf.CellSafeZone)

	citizen1ID := uuid.New()
	citizen2ID := uuid.New()

	sim := Simulation{
		Grid:         &grid,
		EventEmitter: &events.InMemoryEventLog{},
		Citizens: []Citizen{
			{
				ID:               citizen1ID,
				Status:           CitizenIdle,
				CurrentPosition:  pf.Position{X: 0, Y: 0},
				Path:             []pf.Position{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}, {X: 3, Y: 0}, {X: 4, Y: 0}},
				CurrentPathIndex: 0,
			},
			{
				ID:               citizen2ID,
				Status:           CitizenIdle,
				CurrentPosition:  pf.Position{X: 0, Y: 1},
				Path:             []pf.Position{{X: 0, Y: 1}, {X: 1, Y: 1}, {X: 2, Y: 1}, {X: 3, Y: 1}, {X: 4, Y: 1}},
				CurrentPathIndex: 0,
			},
		},
	}

	sim.Tick()

	evts := sim.Events()
	citizen1Moves := 0
	citizen2Moves := 0
	for _, e := range evts {
		if string(e.EventType) == "citizen.moved" {
			if e.EntityID == citizen1ID {
				citizen1Moves++
			}
			if e.EntityID == citizen2ID {
				citizen2Moves++
			}
		}
	}
	require.Equal(t, 1, citizen1Moves, "citizen 1 should have moved once")
	require.Equal(t, 1, citizen2Moves, "citizen 2 should have moved once")
}

func TestCitizenDiedEvent_IncludesMetadata(t *testing.T) {
	simID := uuid.New()
	grid := pf.NewGrid(3, 1, pf.CellOpen)
	grid.UpdateCell(pf.Position{X: 0, Y: 0}, pf.CellHazard)

	citizenID := uuid.New()

	sim := Simulation{
		ID:           simID,
		Grid:         &grid,
		EventEmitter: &events.InMemoryEventLog{},
		Citizens: []Citizen{
			{
				ID:               citizenID,
				Status:           CitizenIdle,
				CurrentPosition:  pf.Position{X: 0, Y: 0},
				Path:             []pf.Position{{X: 0, Y: 0}, {X: 1, Y: 0}},
				CurrentPathIndex: 0,
			},
		},
	}

	sim.Tick()

	evts := sim.Events()
	for _, e := range evts {
		if string(e.EventType) == "citizen.died" {
			require.Equal(t, citizenID, e.EntityID, "citizen.died event entity ID should match citizen")
			require.Equal(t, simID, e.Metadata.SimulationID, "citizen.died metadata should contain simulation ID")
			require.Equal(t, uint64(0), e.Metadata.Tick, "citizen.died should be associated with the correct tick")
			return
		}
	}
	t.Fatal("expected citizen.died event but none found")
}

func TestPausedSimulation_EmitsNoEvents(t *testing.T) {
	grid := pf.NewGrid(3, 1, pf.CellOpen)

	sim := Simulation{
		State:        SimulationPaused,
		Grid:         &grid,
		EventEmitter: &events.InMemoryEventLog{},
		Citizens: []Citizen{
			{
				Status:           CitizenIdle,
				Path:             []pf.Position{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}},
				CurrentPathIndex: 0,
			},
		},
	}

	sim.Tick()
	require.Empty(t, sim.Events(), "paused simulation should not produce events")
}

func TestCompletedSimulation_EmitsNoNewEvents(t *testing.T) {
	grid := pf.NewGrid(3, 1, pf.CellOpen)
	grid.UpdateCell(pf.Position{X: 2, Y: 0}, pf.CellSafeZone)

	sim := Simulation{
		Grid:         &grid,
		EventEmitter: &events.InMemoryEventLog{},
		Citizens: []Citizen{
			{
				Status:           CitizenIdle,
				CurrentPosition:  pf.Position{X: 0, Y: 0},
				Path:             []pf.Position{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}},
				CurrentPathIndex: 0,
			},
		},
	}

	sim.Tick()
	sim.Tick()

	tick2Count := len(sim.Events())

	sim.Tick()

	require.Equal(t, tick2Count, len(sim.Events()),
		"completed simulation should not produce new events on subsequent ticks")
}
