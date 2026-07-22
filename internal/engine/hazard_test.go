package engine

import (
	c "hazard/internal/common"
	"hazard/internal/events"
	pf "hazard/internal/pathfinding"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func newTestSim() (Simulation, error) {
	return NewSimulation(10, 10, c.SimulationConfig{
		CitizenCount: 0,
		Hazard: c.HazardConfig{
			DurationRange: c.PositiveRange{Min: 100, Max: 100},
			Probability:   0,
			Count:         0,
		},
		SafeZone: c.SafeZoneConfig{
			Probability: 0,
			Count:       1,
			RadiusRange: c.Range{Min: 1, Max: 1},
		},
	}, events.CreateEventBus())
}

func TestHazard_RadiusGrowsEachTick(t *testing.T) {
	sim, err := newTestSim()
	require.NoError(t, err)

	sim.Hazards = append(sim.Hazards, c.Hazard{
		ID:            uuid.New(),
		Type:          c.FireHazard,
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
		eventBus: events.CreateEventBus(),
		SafeZones: []c.SafeZone{
			{Position: pf.Position{X: 9, Y: 9}, Radius: 1},
		},
	}

	hazard := c.Hazard{
		ID:            uuid.New(),
		Type:          c.FloodHazard,
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
	sim, err := NewSimulation(10, 10, c.SimulationConfig{
		CitizenCount: 0,
		Hazard: c.HazardConfig{
			DurationRange: c.PositiveRange{Min: 10, Max: 10},
			Probability:   1.0,
			Count:         5,
		},
		SafeZone: c.SafeZoneConfig{
			Probability: 0,
			Count:       0,
			RadiusRange: c.Range{Min: 1, Max: 1},
		},
	}, events.CreateEventBus())
	require.NoError(t, err)
	require.Empty(t, sim.Hazards)

	sim.Tick()
	require.Len(t, sim.Hazards, 1)

	h := sim.Hazards[0]
	require.NotEqual(t, uuid.Nil, h.ID)
	require.Contains(t, []c.HazardType{c.FireHazard, c.FloodHazard, c.LavaHazard}, h.Type)
	require.Equal(t, 0, h.CurrentRadius)
	require.GreaterOrEqual(t, h.Duration, 10)
	require.LessOrEqual(t, h.Duration, 10)
	require.Equal(t, uint64(0), h.CreatedAt)
	require.True(t, sim.Grid.InBounds(h.Origin))
	require.Equal(t, pf.CellHazard, sim.Grid.GetCell(h.Origin))

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

	sz := c.SafeZone{
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
	safeZoneLocations := make(map[pf.Position]*c.SafeZone, len(safeZoneCells))
	for _, cell := range safeZoneCells {
		grid.UpdateCell(cell, pf.CellSafeZone)
		safeZoneLocations[cell] = &sz
	}
	sz.Cells = safeZoneCells

	sim := Simulation{
		Config: c.SimulationConfig{
			SafeZone: c.SafeZoneConfig{
				Count:       1,
				RadiusRange: c.Range{Min: 1, Max: 1},
			},
			Hazard: c.HazardConfig{
				DurationRange: c.PositiveRange{Min: 100, Max: 100},
				Probability:   0,
				Count:         0,
			},
		},
		State:             SimulationCreated,
		Grid:              &grid,
		eventBus:          events.CreateEventBus(),
		safeZoneLocations: safeZoneLocations,
		SafeZones:         []c.SafeZone{sz},
		Citizens: []c.Citizen{
			{
				ID:                 uuid.New(),
				Status:             c.CitizenIdle,
				CurrentPosition:    pf.Position{X: 0, Y: 0},
				CurrentDestination: destination,
				Path:               path,
				CurrentPathIndex:   1,
				TargetSafeZone:     &sz,
			},
		},
		Hazards: []c.Hazard{},
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
