package engine

import (
	"hazard/internal/events"
	pf "hazard/internal/pathfinding"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func newTestSim() (Simulation, error) {
	return NewSimulation(SimulationConfig{
		Width:             10,
		Height:            10,
		CitizenCountRange: [2]int{0, 0},
		Hazard: HazardConfig{
			DurationRange: [2]int{100, 100},
			Probability:   0,
			CountRange:    [2]int{0, 0},
		},
		SafeZone: SafeZoneConfig{
			Probability: 0,
			CountRange:  [2]int{1, 1},
			RadiusRange: [2]int{1, 1},
		},
	})
}

func TestHazard_RadiusGrowsEachTick(t *testing.T) {
	sim, err := newTestSim()
	require.NoError(t, err)

	sim.Hazards = append(sim.Hazards, Hazard{
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

func TestHazard_CellsBlockPathfinding(t *testing.T) {
	grid := pf.NewGrid(5, 5, pf.CellOpen)
	hazard := Hazard{
		Origin:        pf.Position{X: 2, Y: 2},
		CurrentRadius: 0,
	}

	grid.UpdateCell(pf.Position{X: 2, Y: 2}, pf.CellHazard)
	hazard.expandHazard(&grid)

	require.Equal(t, pf.CellHazard, grid.GetCell(pf.Position{X: 2, Y: 2}))
	require.Equal(t, pf.CellHazard, grid.GetCell(pf.Position{X: 2, Y: 1}))
	require.Equal(t, pf.CellHazard, grid.GetCell(pf.Position{X: 2, Y: 3}))
	require.Equal(t, pf.CellHazard, grid.GetCell(pf.Position{X: 1, Y: 2}))
	require.Equal(t, pf.CellHazard, grid.GetCell(pf.Position{X: 3, Y: 2}))

	path, err := pf.FindPath(&grid, pf.Position{X: 0, Y: 0}, pf.Position{X: 4, Y: 4})
	require.NoError(t, err)
	for _, pos := range path {
		require.NotEqual(t, pf.CellHazard, grid.GetCell(pos), "path must not enter hazard cell at %v", pos)
	}

	for y := range grid.Height {
		for x := range grid.Width {
			grid.Cells[y][x] = pf.CellHazard
		}
	}
	_, err = pf.FindPath(&grid, pf.Position{X: 0, Y: 0}, pf.Position{X: 4, Y: 4})
	require.ErrorIs(t, err, pf.ErrDestinationUnreachable)
}

func TestHazard_RemovedAfterDuration(t *testing.T) {
	sim, err := newTestSim()
	require.NoError(t, err)

	hazard := Hazard{
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

func TestCreateHazard(t *testing.T) {
	grid := pf.NewGrid(10, 10, pf.CellOpen)
	config := HazardConfig{
		DurationRange: [2]int{5, 10},
		Probability:   0.5,
		CountRange:    [2]int{0, 0},
	}

	hazard, err := createHazard(config, &grid)
	require.NoError(t, err)
	require.GreaterOrEqual(t, hazard.Duration, 5)
	require.LessOrEqual(t, hazard.Duration, 10)
	require.Contains(t, []string{"fire", "flood", "lava"}, hazard.Type.Name)
	require.True(t, grid.InBounds(hazard.Origin))
	require.Equal(t, pf.CellHazard, grid.GetCell(hazard.Origin))
}

func TestCreateHazard_NoOpenCells(t *testing.T) {
	grid := pf.NewGrid(2, 2, pf.CellObstacle)
	config := HazardConfig{
		DurationRange: [2]int{5, 5},
	}

	_, err := createHazard(config, &grid)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no open cells available")
}

func TestRandomHazardType(t *testing.T) {
	seen := make(map[string]bool)
	for range 100 {
		ht := randomHazardType()
		require.Contains(t, HazardTypes, ht.Name)
		seen[ht.Name] = true
	}
	require.GreaterOrEqual(t, len(seen), 2)
}

func TestHazard_CreationViaTick(t *testing.T) {
	sim, err := NewSimulation(SimulationConfig{
		Width:             10,
		Height:            10,
		CitizenCountRange: [2]int{0, 0},
		Hazard: HazardConfig{
			DurationRange: [2]int{10, 10},
			Probability:   1.0,
			CountRange:    [2]int{5, 5},
		},
		SafeZone: SafeZoneConfig{
			Probability: 0,
			CountRange:  [2]int{0, 0},
			RadiusRange: [2]int{1, 1},
		},
	})
	require.NoError(t, err)
	require.Empty(t, sim.Hazards)

	sim.Tick()
	require.Len(t, sim.Hazards, 1)

	h := sim.Hazards[0]
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

	sim := Simulation{
		Config: SimulationConfig{
			SafeZone: SafeZoneConfig{
				CountRange:  [2]int{1, 1},
				RadiusRange: [2]int{1, 1},
			},
			Hazard: HazardConfig{
				DurationRange: [2]int{100, 100},
				Probability:   0,
				CountRange:    [2]int{0, 0},
			},
		},
		State:        SimulationCreated,
		Grid:         &grid,
		EventEmitter: &events.InMemoryEventLog{},
		MaxHazards:   0,
		MaxSafeZones: 1,
		SafeZones: []SafeZone{
			{Position: destination, Radius: 1},
		},
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
		Hazards: []Hazard{},
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
	require.Equal(t, destination, sim.Citizens[0].Path[len(sim.Citizens[0].Path)-1],
		"recalculated path must still lead to destination")
}
