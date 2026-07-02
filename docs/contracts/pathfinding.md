# Pathfinding Interface Contract

## Go Interface

```go
// Pathfinder defines the contract for pathfinding algorithms.
// All implementations must satisfy this interface.
type Pathfinder interface {
    // FindPath returns a sequence of positions from start to goal
    // that avoids blocked cells. Returns nil if no path exists.
    FindPath(grid *Grid, start, goal Position) []Position

    // Name returns the algorithm name for logging/metrics.
    Name() string
}
```

## Grid Contract

```go
// Grid provides the environment representation for pathfinding.
type Grid struct {
    Width  int
    Height int
    Cells  [][]CellType
}

// CellType indicates passability.
type CellType int

const (
    CellOpen     CellType = 0  // Traversable
    CellObstacle CellType = 1  // Impassable (static)
    CellHazard   CellType = 2  // Impassable (dynamic)
    CellSafeZone CellType = 3  // Goal
)
```

## Expectations

1. **Deterministic**: Same grid + same start/goal → same path (seed-independent)
2. **A* (initial)**: Manhattan distance heuristic for grid movement
3. **Shortest path**: Minimize number of cells traversed
4. **No diagonal movement** (v1): Only cardinal directions (N, S, E, W)
5. **Performance**: Must compute for 100 citizens within one tick (100ms target)

## Future Extensions

- Weighted A* (penalize cells near hazards for "safest" path preference)
- D* Lite for incremental replanning when hazards expand
- Hierarchical pathfinding for large grids
