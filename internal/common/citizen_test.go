package common

import (
	pf "hazard/internal/pathfinding"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCreateCitizens_PlacedOnOpenGridCells(t *testing.T) {
	grid := pf.NewGrid(10, 10, pf.CellOpen)
	// Place a safe zone so pathfinding works
	grid.UpdateCell(pf.Position{X: 9, Y: 9}, pf.CellSafeZone)

	safeZone := SafeZone{
		Position:    pf.Position{X: 9, Y: 9},
		Radius:      0,
		Cells:       []pf.Position{{X: 9, Y: 9}},
		HasCapacity: true,
	}
	safeZoneLocations := map[pf.Position]*SafeZone{
		{X: 9, Y: 9}: &safeZone,
	}

	citizens := CreateCitizens(PositiveRange{Min: 3, Max: 3}, &grid, safeZoneLocations)

	require.Len(t, citizens, 3)
	for _, c := range citizens {
		require.NotEqual(t, uuid.Nil, c.ID)
		require.Equal(t, CitizenIdle, c.Status)
		require.True(t, grid.InBounds(c.CurrentPosition))
		require.NotEmpty(t, c.Path)
		require.Equal(t, pf.CellSafeZone, grid.GetCell(c.Path[len(c.Path)-1]),
			"path must end at a safe zone cell")
		require.Zero(t, c.CurrentPathIndex)
	}
}

func TestCreateCitizens_ReturnsEmptyWhenNoOpenCells(t *testing.T) {
	grid := pf.NewGrid(2, 2, pf.CellObstacle)

	citizens := CreateCitizens(PositiveRange{Min: 5, Max: 5}, &grid, make(map[pf.Position]*SafeZone))

	require.Empty(t, citizens)
}

func TestFindNearestSafeZone_FindsPathToSafeZone(t *testing.T) {
	grid := pf.NewGrid(5, 5, pf.CellOpen)
	grid.UpdateCell(pf.Position{X: 4, Y: 4}, pf.CellSafeZone)

	safeZone := SafeZone{
		Position:    pf.Position{X: 4, Y: 4},
		Radius:      0,
		Cells:       []pf.Position{{X: 4, Y: 4}},
		HasCapacity: true,
	}
	safeZoneLocations := map[pf.Position]*SafeZone{
		{X: 4, Y: 4}: &safeZone,
	}

	citizen := Citizen{
		ID:              uuid.New(),
		Status:          CitizenIdle,
		CurrentPosition: pf.Position{X: 0, Y: 0},
	}

	err := citizen.FindNearestSafeZone(&grid, safeZoneLocations)
	require.NoError(t, err)
	require.NotEmpty(t, citizen.Path)
	require.Equal(t, pf.Position{X: 4, Y: 4}, citizen.CurrentDestination)
	require.Equal(t, pf.CellSafeZone, grid.GetCell(citizen.Path[len(citizen.Path)-1]))
}

func TestFindNearestSafeZone_ReturnsErrorWhenNoSafeZone(t *testing.T) {
	grid := pf.NewGrid(3, 3, pf.CellOpen)

	citizen := Citizen{
		ID:              uuid.New(),
		Status:          CitizenIdle,
		CurrentPosition: pf.Position{X: 0, Y: 0},
	}

	err := citizen.FindNearestSafeZone(&grid, make(map[pf.Position]*SafeZone))
	require.Error(t, err)
	require.Empty(t, citizen.Path)
}

func TestRecalculatePath_RecalculatesAroundObstacle(t *testing.T) {
	grid := pf.NewGrid(5, 5, pf.CellOpen)

	// Build a wall that blocks the direct corridor from (0,0) to (4,4),
	// forcing a detour
	grid.UpdateCell(pf.Position{X: 1, Y: 0}, pf.CellObstacle)
	grid.UpdateCell(pf.Position{X: 1, Y: 1}, pf.CellObstacle)
	grid.UpdateCell(pf.Position{X: 1, Y: 2}, pf.CellObstacle)
	grid.UpdateCell(pf.Position{X: 1, Y: 3}, pf.CellObstacle)

	citizen := Citizen{
		ID:                 uuid.New(),
		Status:             CitizenIdle,
		CurrentPosition:    pf.Position{X: 0, Y: 0},
		CurrentDestination: pf.Position{X: 4, Y: 4},
	}

	err := citizen.RecalculatePath(&grid)
	require.NoError(t, err)
	require.NotEmpty(t, citizen.Path)
	require.Equal(t, pf.Position{X: 0, Y: 0}, citizen.Path[0])
	require.Zero(t, citizen.CurrentPathIndex)
	for _, pos := range citizen.Path {
		require.NotEqual(t, pf.CellObstacle, grid.GetCell(pos),
			"path must not enter obstacle cell at %v", pos)
	}
}

func TestRecalculatePath_ReturnsErrorForUnreachableDestination(t *testing.T) {
	grid := pf.NewGrid(3, 3, pf.CellObstacle)

	citizen := Citizen{
		ID:                 uuid.New(),
		Status:             CitizenIdle,
		CurrentPosition:    pf.Position{X: 0, Y: 0},
		CurrentDestination: pf.Position{X: 2, Y: 2},
	}

	err := citizen.RecalculatePath(&grid)
	require.Error(t, err)
}

func TestIncrementLocation_MovesCitizenOneStep(t *testing.T) {
	grid := pf.NewGrid(3, 1, pf.CellOpen)

	citizen := Citizen{
		Status:           CitizenIdle,
		CurrentPosition:  pf.Position{X: 0, Y: 0},
		Path:             []pf.Position{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}},
		CurrentPathIndex: 0,
	}

	moved, _ := citizen.IncrementLocation(&grid)

	require.True(t, moved)
	require.Equal(t, 1, citizen.CurrentPathIndex)
	require.Equal(t, pf.Position{X: 1, Y: 0}, citizen.CurrentPosition)
	require.Equal(t, CitizenNavigating, citizen.Status)
}

func TestIncrementLocation_ReachingEndReportsEscaped(t *testing.T) {
	grid := pf.NewGrid(2, 1, pf.CellOpen)

	citizen := Citizen{
		Status:           CitizenIdle,
		CurrentPosition:  pf.Position{X: 0, Y: 0},
		Path:             []pf.Position{{X: 0, Y: 0}, {X: 1, Y: 0}},
		CurrentPathIndex: 0,
	}

	moved, escaped := citizen.IncrementLocation(&grid)
	require.True(t, moved)
	require.True(t, escaped, "last step should report escaped")
	require.Equal(t, CitizenNavigating, citizen.Status)

	moved, escaped = citizen.IncrementLocation(&grid)

	require.False(t, moved, "citizen at path end should not report movement")
	require.True(t, escaped, "citizen at path end should report escaped")
	require.Equal(t, 1, citizen.CurrentPathIndex)
	require.Equal(t, pf.Position{X: 1, Y: 0}, citizen.CurrentPosition)
	require.Equal(t, CitizenNavigating, citizen.Status,
		"IncrementLocation no longer sets CitizenEscaped; caller must handle it")
}

func TestIncrementLocation_DoesNotMoveEscapedCitizen(t *testing.T) {
	citizen := Citizen{
		Status:           CitizenEscaped,
		CurrentPosition:  pf.Position{X: 2, Y: 0},
		Path:             []pf.Position{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}},
		CurrentPathIndex: 2,
	}

	moved, _ := citizen.IncrementLocation(nil)

	require.False(t, moved)
	require.Equal(t, 2, citizen.CurrentPathIndex)
	require.Equal(t, CitizenEscaped, citizen.Status)
}

func TestIncrementLocation_DoesNotMoveDeadCitizen(t *testing.T) {
	citizen := Citizen{
		Status:           CitizenDead,
		CurrentPosition:  pf.Position{X: 1, Y: 0},
		Path:             []pf.Position{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}},
		CurrentPathIndex: 1,
	}

	moved, _ := citizen.IncrementLocation(nil)

	require.False(t, moved)
	require.Equal(t, 1, citizen.CurrentPathIndex)
	require.Equal(t, CitizenDead, citizen.Status)
}

func TestCreateCitizens_MarksCellCitizenOnGrid(t *testing.T) {
	grid := pf.NewGrid(10, 10, pf.CellOpen)
	grid.UpdateCell(pf.Position{X: 9, Y: 9}, pf.CellSafeZone)

	safeZone := SafeZone{
		Position:    pf.Position{X: 9, Y: 9},
		Radius:      0,
		Cells:       []pf.Position{{X: 9, Y: 9}},
		HasCapacity: true,
	}
	safeZoneLocations := map[pf.Position]*SafeZone{
		{X: 9, Y: 9}: &safeZone,
	}

	citizens := CreateCitizens(PositiveRange{Min: 3, Max: 3}, &grid, safeZoneLocations)

	require.Len(t, citizens, 3)
	for _, c := range citizens {
		require.Equal(t, pf.CellCitizen, grid.GetCell(c.CurrentPosition),
			"citizen at %v must be marked as CellCitizen", c.CurrentPosition)
	}
}

func TestIncrementLocation_UnmarksPreviousCellAndMarksNewCell(t *testing.T) {
	grid := pf.NewGrid(3, 1, pf.CellOpen)

	citizen := Citizen{
		Status:           CitizenIdle,
		CurrentPosition:  pf.Position{X: 0, Y: 0},
		Path:             []pf.Position{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}},
		CurrentPathIndex: 0,
		PreviousCellType: pf.CellOpen,
	}

	// Mark start cell as citizen before movement (as CreateCitizens does)
	grid.UpdateCell(pf.Position{X: 0, Y: 0}, pf.CellCitizen)

	moved, _ := citizen.IncrementLocation(&grid)

	require.True(t, moved)
	// Previous cell should be restored
	require.Equal(t, pf.CellOpen, grid.GetCell(pf.Position{X: 0, Y: 0}),
		"previous cell must be restored to open after departure")
	// New cell should be marked as citizen
	require.Equal(t, pf.CellCitizen, grid.GetCell(pf.Position{X: 1, Y: 0}),
		"new cell must be marked as CellCitizen after arrival")
	// PreviousCellType should reflect the type of the new cell before occupation
	require.Equal(t, pf.CellOpen, citizen.PreviousCellType)
}

func TestIncrementLocation_RestoresPreviousCellTypeOnMove(t *testing.T) {
	grid := pf.NewGrid(3, 1, pf.CellOpen)
	grid.UpdateCell(pf.Position{X: 1, Y: 0}, pf.CellSafeZone)

	citizen := Citizen{
		Status:           CitizenIdle,
		CurrentPosition:  pf.Position{X: 0, Y: 0},
		Path:             []pf.Position{{X: 0, Y: 0}, {X: 1, Y: 0}},
		CurrentPathIndex: 0,
		PreviousCellType: pf.CellCitizen,
	}

	// Current cell is marked as citizen, simulating that citizen was already here before
	grid.UpdateCell(pf.Position{X: 0, Y: 0}, pf.CellCitizen)

	moved, _ := citizen.IncrementLocation(&grid)

	require.True(t, moved)
	// Previous (departure) cell should revert to the citizen-marking
	require.Equal(t, pf.CellCitizen, grid.GetCell(pf.Position{X: 0, Y: 0}),
		"previous cell must be restored to its PreviousCellType (CellCitizen)")
	// New cell was SafeZone before occupation, so PreviousCellType should be CellSafeZone
	require.Equal(t, pf.CellSafeZone, citizen.PreviousCellType)
}
