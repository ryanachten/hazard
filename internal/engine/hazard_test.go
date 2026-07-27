package engine

import (
	c "hazard/internal/citizen"
	config "hazard/internal/configuration"
	"hazard/internal/events"
	h "hazard/internal/hazard"
	pf "hazard/internal/pathfinding"
	r "hazard/internal/ranging"
	sz "hazard/internal/safe_zone"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func newTestSim() (Simulation, error) {
	return NewSimulation(10, 10, config.SimulationConfig{
		CitizenCount: 0,
		Hazard: h.Config{
			DurationRange: r.PositiveRange{Min: 100, Max: 100},
			Probability:   0,
			Count:         0,
		},
		SafeZone: sz.Config{
			Probability: 0,
			Count:       1,
			RadiusRange: r.Range{Min: 1, Max: 1},
		},
	}, events.New())
}

func TestHazard_RadiusGrowsEachTick(t *testing.T) {
	sim, err := newTestSim()
	require.NoError(t, err)

	sim.Hazards = append(sim.Hazards, h.Hazard{
		ID:            uuid.New(),
		Type:          h.FireHazard,
		CreatedAt:     0,
		Duration:      100,
		Origin:        pf.Position{X: 5, Y: 5},
		CurrentRadius: 0,
	})

	require.Equal(t, 0, sim.Hazards[0].CurrentRadius)

	sim.Tick()
	require.Equal(t, 1, sim.Hazards[0].CurrentRadius)

	sim.Tick()
	require.Equal(t, 2, sim.Hazards[0].CurrentRadius)

	sim.Tick()
	require.Equal(t, 3, sim.Hazards[0].CurrentRadius)
}

func TestHazard_RemovedAfterDuration(t *testing.T) {
	grid := pf.NewGrid(10, 10, pf.CellOpen)
	// Place safe zone far from the hazard area so expansion/removal is deterministic
	grid.UpdateCell(pf.Position{X: 9, Y: 9}, pf.CellSafeZone)

	sim := Simulation{
		Grid:     &grid,
		eventBus: events.New(),
		SafeZones: []sz.SafeZone{
			{Position: pf.Position{X: 9, Y: 9}, Radius: 1},
		},
	}

	hazard := h.Hazard{
		ID:            uuid.New(),
		Type:          h.FloodHazard,
		CreatedAt:     0,
		Duration:      1,
		Origin:        pf.Position{X: 5, Y: 5},
		CurrentRadius: 0,
	}
	sim.Grid.UpdateCell(pf.Position{X: 5, Y: 5}, pf.CellHazard)
	sim.Hazards = append(sim.Hazards, hazard)

	require.Len(t, sim.Hazards, 1)

	sim.Tick()
	require.Len(t, sim.Hazards, 1)

	sim.Tick()
	require.Len(t, sim.Hazards, 1)

	sim.Tick()
	require.Len(t, sim.Hazards, 0)

	require.Equal(t, pf.CellOpen, sim.Grid.GetCell(pf.Position{X: 5, Y: 5}),
		"origin cell must be restored to open after hazard removal")
	require.Equal(t, pf.CellOpen, sim.Grid.GetCell(pf.Position{X: 5, Y: 4}),
		"expanded cell at radius 1 must be restored to open")
	require.Equal(t, pf.CellOpen, sim.Grid.GetCell(pf.Position{X: 5, Y: 3}),
		"expanded cell at radius 2 must be restored to open")
}

func TestHazard_CreationViaTick(t *testing.T) {
	sim, err := NewSimulation(10, 10, config.SimulationConfig{
		CitizenCount: 0,
		Hazard: h.Config{
			DurationRange: r.PositiveRange{Min: 10, Max: 10},
			Probability:   1.0,
			Count:         5,
		},
		SafeZone: sz.Config{
			Probability: 0,
			Count:       0,
			RadiusRange: r.Range{Min: 1, Max: 1},
		},
	}, events.New())
	require.NoError(t, err)
	require.Empty(t, sim.Hazards)

	sim.Tick()
	require.Len(t, sim.Hazards, 1)

	h1 := sim.Hazards[0]
	require.NotEqual(t, uuid.Nil, h1.ID)
	require.Contains(t, []h.Type{h.FireHazard, h.FloodHazard, h.LavaHazard}, h1.Type)
	require.Equal(t, 0, h1.CurrentRadius)
	require.GreaterOrEqual(t, h1.Duration, 10)
	require.LessOrEqual(t, h1.Duration, 10)
	require.Equal(t, uint64(0), h1.CreatedAt)
	require.True(t, sim.Grid.InBounds(h1.Origin))
	require.Equal(t, pf.CellHazard, sim.Grid.GetCell(h1.Origin))

	sim.Tick()
	require.Len(t, sim.Hazards, 2)

	sim.Tick()
	require.Len(t, sim.Hazards, 3)
}

func TestHazard_BlocksCitizenPath(t *testing.T) {
	grid := pf.NewGrid(5, 5, pf.CellOpen)
	destination := pf.Position{X: 4, Y: 4}

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

	sz1 := sz.SafeZone{
		ID:          uuid.New(),
		Position:    destination,
		Radius:      1,
		HasCapacity: true,
	}
	// Mark safe zone cells on the grid and populate location map
	safeZoneCells := []pf.Position{
		{X: 3, Y: 3}, {X: 3, Y: 4},
		{X: 4, Y: 3}, {X: 4, Y: 4},
	}
	safeZoneLocations := make(map[pf.Position]*sz.SafeZone, len(safeZoneCells))
	for _, cell := range safeZoneCells {
		grid.UpdateCell(cell, pf.CellSafeZone)
		safeZoneLocations[cell] = &sz1
	}
	sz1.Cells = safeZoneCells

	sim := Simulation{
		Config: config.SimulationConfig{
			SafeZone: sz.Config{
				Count:       1,
				RadiusRange: r.Range{Min: 1, Max: 1},
			},
			Hazard: h.Config{
				DurationRange: r.PositiveRange{Min: 100, Max: 100},
				Probability:   0,
				Count:         0,
			},
		},
		State:             SimulationCreated,
		Grid:              &grid,
		eventBus:          events.New(),
		safeZoneLocations: safeZoneLocations,
		SafeZones:         []sz.SafeZone{sz1},
		Citizens: []c.Citizen{
			{
				ID:                 uuid.New(),
				Status:             c.StatusIdle,
				CurrentPosition:    pf.Position{X: 0, Y: 0},
				CurrentDestination: destination,
				Path:               path,
				CurrentPathIndex:   1,
				TargetSafeZone:     &sz1,
			},
		},
		Hazards: []h.Hazard{},
	}

	blockedCell := pf.Position{X: 0, Y: 2}
	grid.UpdateCell(blockedCell, pf.CellHazard)
	require.Equal(t, pf.CellHazard, grid.GetCell(blockedCell))

	origPath := make([]pf.Position, len(sim.Citizens[0].Path))
	copy(origPath, sim.Citizens[0].Path)

	sim.Tick()

	require.NotEqual(t, origPath, sim.Citizens[0].Path,
		"path should have been recalculated to avoid hazard")
	require.NotContains(t, sim.Citizens[0].Path, blockedCell,
		"recalculated path must avoid hazard cell")
	require.Equal(t, 1, sim.Citizens[0].CurrentPathIndex,
		"citizen advanced one step after recalculation")
	lastCell := sim.Citizens[0].Path[len(sim.Citizens[0].Path)-1]
	require.Equal(t, pf.CellSafeZone, sim.Grid.GetCell(lastCell),
		"recalculated path must lead to a safe zone cell")
}
