package safezone

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"hazard/internal/bounds"
	"hazard/internal/pathfinding"
)

func TestCreateSafeZone_PlacedAtOpenPosition(t *testing.T) {
	grid := pathfinding.NewGrid(10, 10, pathfinding.CellOpen)
	config := Config{
		Count:       1,
		RadiusRange: bounds.Range{Min: 1, Max: 1},
	}

	safeZone, err := Create(config, &grid)

	require.NoError(t, err)
	require.True(t, grid.InBounds(safeZone.Position))
	require.NotEmpty(t, safeZone.Cells)
	for _, cell := range safeZone.Cells {
		require.Equal(t, pathfinding.CellSafeZone, grid.GetCell(cell),
			"cell %v must be marked as safe zone", cell)
	}
}

func TestCreateSafeZone_RadiusZeroMarksOnlyOrigin(t *testing.T) {
	grid := pathfinding.NewGrid(5, 5, pathfinding.CellOpen)
	config := Config{
		Count:       1,
		RadiusRange: bounds.Range{Min: 0, Max: 0},
	}

	safeZone, err := Create(config, &grid)

	require.NoError(t, err)
	require.Len(t, safeZone.Cells, 1)
	require.Equal(t, safeZone.Position, safeZone.Cells[0])
	require.Equal(t, pathfinding.CellSafeZone, grid.GetCell(safeZone.Position))
}

func TestCreateSafeZone_DoesNotOverwriteNonOpenCells(t *testing.T) {
	grid := pathfinding.NewGrid(5, 5, pathfinding.CellObstacle)
	// Only one open cell, forcing the origin to (2,1) with radius 1
	grid.UpdateCell(pathfinding.Position{X: 2, Y: 1}, pathfinding.CellOpen)
	grid.UpdateCell(pathfinding.Position{X: 2, Y: 2}, pathfinding.CellObstacle)
	config := Config{
		Count:       1,
		RadiusRange: bounds.Range{Min: 1, Max: 1},
	}

	safeZone, err := Create(config, &grid)

	require.NoError(t, err)
	require.Equal(t, pathfinding.CellObstacle, grid.GetCell(pathfinding.Position{X: 2, Y: 2}),
		"obstacle cell must not be overwritten by safe zone")
	for _, cell := range safeZone.Cells {
		require.NotEqual(t, pathfinding.Position{X: 2, Y: 2}, cell,
			"safe zone should not claim obstacle cell")
	}
}

func TestCreateSafeZone_ReturnsErrorWhenNoOpenCells(t *testing.T) {
	grid := pathfinding.NewGrid(4, 4, pathfinding.CellObstacle)
	config := Config{
		Count:       1,
		RadiusRange: bounds.Range{Min: 1, Max: 1},
	}

	_, err := Create(config, &grid)

	require.Error(t, err)
	require.Contains(t, err.Error(), "no open cells")
}

func TestCreateSafeZone_RadiusMarksMultipleCells(t *testing.T) {
	grid := pathfinding.NewGrid(5, 5, pathfinding.CellOpen)
	config := Config{
		Count:       1,
		RadiusRange: bounds.Range{Min: 2, Max: 2},
	}

	safeZone, err := Create(config, &grid)

	require.NoError(t, err)
	require.Equal(t, 2, safeZone.Radius)
	// A radius-2 square from origin would cover up to (2r+1)^2 = 25 cells,
	// but the origin position is random, so we just check at least some cells are marked
	require.Greater(t, len(safeZone.Cells), 1)
}

func TestAddOccupant_AdmitsCitizenAndMarksCell(t *testing.T) {
	grid := pathfinding.NewGrid(3, 3, pathfinding.CellOpen)
	sz := SafeZone{
		Cells:         []pathfinding.Position{{X: 1, Y: 1}},
		HasCapacity:   true,
		Occupants:     []uuid.UUID{},
		occupiedCells: map[pathfinding.Position]bool{},
	}

	pos, ok := sz.AddOccupant(uuid.New(), pathfinding.Position{X: 1, Y: 1}, &grid)

	require.True(t, ok)
	require.Equal(t, pathfinding.Position{X: 1, Y: 1}, pos)
	require.Len(t, sz.Occupants, 1)
	require.Equal(t, pathfinding.CellEscapedCitizen, grid.GetCell(pos))
}

func TestAddOccupant_DeniesWhenAtCapacity(t *testing.T) {
	grid := pathfinding.NewGrid(3, 3, pathfinding.CellOpen)
	// Single-cell safe zone, one occupant fills it
	grid.UpdateCell(pathfinding.Position{X: 1, Y: 1}, pathfinding.CellSafeZone)
	sz := SafeZone{
		Cells:         []pathfinding.Position{{X: 1, Y: 1}},
		HasCapacity:   true,
		Occupants:     []uuid.UUID{},
		occupiedCells: map[pathfinding.Position]bool{},
	}

	// First occupant admitted — single-cell zone is now at capacity
	_, ok := sz.AddOccupant(uuid.New(), pathfinding.Position{X: 1, Y: 1}, &grid)
	require.True(t, ok)
	require.False(t, sz.HasCapacity, "single-cell zone should be at capacity after first occupant")

	// Second occupant must be denied — zone is full
	_, ok = sz.AddOccupant(uuid.New(), pathfinding.Position{X: 1, Y: 1}, &grid)
	require.False(t, ok, "single-cell zone with one occupant must deny second entrant")
	require.False(t, sz.HasCapacity)
}

func TestAddOccupant_ReassignsCellIfArrivalCellTaken(t *testing.T) {
	grid := pathfinding.NewGrid(3, 3, pathfinding.CellOpen)
	sz := SafeZone{
		Cells:         []pathfinding.Position{{X: 1, Y: 1}, {X: 2, Y: 1}},
		HasCapacity:   true,
		Occupants:     []uuid.UUID{},
		occupiedCells: map[pathfinding.Position]bool{},
	}

	// First occupant takes (1,1)
	pos1, ok := sz.AddOccupant(uuid.New(), pathfinding.Position{X: 1, Y: 1}, &grid)
	require.True(t, ok)
	require.Equal(t, pathfinding.Position{X: 1, Y: 1}, pos1)

	// Second occupant arrives at (1,1) but gets reassigned to (2,1)
	pos2, ok := sz.AddOccupant(uuid.New(), pathfinding.Position{X: 1, Y: 1}, &grid)
	require.True(t, ok)
	require.Equal(t, pathfinding.Position{X: 2, Y: 1}, pos2)
	require.Len(t, sz.Occupants, 2)
	require.Equal(t, pathfinding.CellEscapedCitizen, grid.GetCell(pos2))
}

func TestSafeZone_CopyCreatesDeepCopy(t *testing.T) {
	original := SafeZone{
		ID:            uuid.New(),
		Position:      pathfinding.Position{X: 5, Y: 5},
		Radius:        2,
		Cells:         []pathfinding.Position{{X: 4, Y: 4}, {X: 5, Y: 5}},
		HasCapacity:   true,
		Occupants:     []uuid.UUID{uuid.New()},
		occupiedCells: map[pathfinding.Position]bool{{X: 4, Y: 4}: true},
	}
	copied := original.Copy()

	require.Equal(t, original.ID, copied.ID)
	require.Equal(t, original.Position, copied.Position)
	require.Equal(t, original.Radius, copied.Radius)
	require.Equal(t, original.HasCapacity, copied.HasCapacity)
	require.Equal(t, original.Occupants, copied.Occupants)

	// Mutate originals to verify deep copy
	original.Cells[0] = pathfinding.Position{X: 9, Y: 9}
	original.Occupants[0] = uuid.Nil
	require.NotEqual(t, original.Cells[0], copied.Cells[0],
		"mutating original cells must not affect copy")
	require.NotEqual(t, original.Occupants[0], copied.Occupants[0],
		"mutating original occupants must not affect copy")
}
