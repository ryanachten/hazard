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

	citizens := CreateCitizens([2]int{3, 3}, &grid)

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

	citizens := CreateCitizens([2]int{5, 5}, &grid)

	require.Empty(t, citizens)
}

func TestFindNearestSafeZone_FindsPathToSafeZone(t *testing.T) {
	grid := pf.NewGrid(5, 5, pf.CellOpen)
	grid.UpdateCell(pf.Position{X: 4, Y: 4}, pf.CellSafeZone)

	citizen := Citizen{
		ID:              uuid.New(),
		Status:          CitizenIdle,
		CurrentPosition: pf.Position{X: 0, Y: 0},
	}

	err := citizen.FindNearestSafeZone(&grid)
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

	err := citizen.FindNearestSafeZone(&grid)
	require.Error(t, err)
	require.Empty(t, citizen.Path)
}

func TestUpdatePath_RecalculatesAroundObstacle(t *testing.T) {
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

	err := citizen.UpdatePath(&grid)
	require.NoError(t, err)
	require.NotEmpty(t, citizen.Path)
	require.Equal(t, pf.Position{X: 0, Y: 0}, citizen.Path[0])
	require.Zero(t, citizen.CurrentPathIndex)
	for _, pos := range citizen.Path {
		require.NotEqual(t, pf.CellObstacle, grid.GetCell(pos),
			"path must not enter obstacle cell at %v", pos)
	}
}

func TestUpdatePath_ReturnsErrorForUnreachableDestination(t *testing.T) {
	grid := pf.NewGrid(3, 3, pf.CellObstacle)

	citizen := Citizen{
		ID:                 uuid.New(),
		Status:             CitizenIdle,
		CurrentPosition:    pf.Position{X: 0, Y: 0},
		CurrentDestination: pf.Position{X: 2, Y: 2},
	}

	err := citizen.UpdatePath(&grid)
	require.Error(t, err)
}

func TestIncrementLocation_MovesCitizenOneStep(t *testing.T) {
	citizen := Citizen{
		Status:           CitizenIdle,
		CurrentPosition:  pf.Position{X: 0, Y: 0},
		Path:             []pf.Position{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}},
		CurrentPathIndex: 0,
	}

	moved := citizen.IncrementLocation()

	require.True(t, moved)
	require.Equal(t, 1, citizen.CurrentPathIndex)
	require.Equal(t, pf.Position{X: 1, Y: 0}, citizen.CurrentPosition)
	require.Equal(t, CitizenNavigating, citizen.Status)
}

func TestIncrementLocation_ReachingEndMarksEscaped(t *testing.T) {
	citizen := Citizen{
		Status:           CitizenIdle,
		CurrentPosition:  pf.Position{X: 0, Y: 0},
		Path:             []pf.Position{{X: 0, Y: 0}, {X: 1, Y: 0}},
		CurrentPathIndex: 0,
	}

	moved := citizen.IncrementLocation()
	require.True(t, moved)

	moved = citizen.IncrementLocation()

	require.False(t, moved, "citizen at path end should not report movement")
	require.Equal(t, 1, citizen.CurrentPathIndex)
	require.Equal(t, pf.Position{X: 1, Y: 0}, citizen.CurrentPosition)
	require.Equal(t, CitizenEscaped, citizen.Status)
}

func TestIncrementLocation_DoesNotMoveEscapedCitizen(t *testing.T) {
	citizen := Citizen{
		Status:           CitizenEscaped,
		CurrentPosition:  pf.Position{X: 2, Y: 0},
		Path:             []pf.Position{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}},
		CurrentPathIndex: 2,
	}

	moved := citizen.IncrementLocation()

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

	moved := citizen.IncrementLocation()

	require.False(t, moved)
	require.Equal(t, 1, citizen.CurrentPathIndex)
	require.Equal(t, CitizenDead, citizen.Status)
}
