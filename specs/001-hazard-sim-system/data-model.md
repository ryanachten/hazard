# Data Model: Hazard Simulation System

## Entity Overview

```
Simulation ──1:N──→ Citizen
Simulation ──1:N──→ Hazard
Simulation ──1:N──→ SafeZone
Simulation ──1:N──→ StaticObstacle
Simulation ──1:N──→ SimulationEvent
Simulation ──1:1──→ Environment (configuration only)
```

## Simulation

```go
type Simulation struct {
    ID          string          // UUID
    Config      SimulationConfig
    State       SimulationState // created → running → paused → completed
    Tick        uint64          // current tick number
    StartedAt   time.Time
    CompletedAt *time.Time
    Citizens    []Citizen
    Hazards     []Hazard
    SafeZones   []SafeZone
    Obstacles   []StaticObstacle
    Grid        *Grid           // 2D grid representation for pathfinding
}
```

**State transitions**: `created → running ⇄ paused → completed`

- `created`: Environment configured, citizens seeded, waiting for start
- `running`: Tick loop active; citizens idle (random movement) until the first hazard emerges, then navigate to safe zones as hazards expand. Additional safe zones may emerge at a configurable interval.
- `paused`: Tick loop suspended, state preserved
- `completed`: All citizens escaped or dead

**Validation rules**:
- Must have at least 1 safe zone (otherwise citizens have no goal)
- Dimensions must be positive (width > 0, height > 0)

## SimulationConfig

```go
type SimulationConfig struct {
    Width                   int              // Grid width in cells
    Height                  int              // Grid height in cells
    TickInterval            time.Duration    // e.g., 100ms (10 Hz default)
    Seed                    int64            // RNG seed for reproducibility
    CitizenCountRange       [2]int           // [min, max] citizens to spawn
    CitizenSpeedRange       [2]int           // [min, max] cells per tick
    CitizenRiskTolerance    [2]float64       // [min, max] per-citizen risk tolerance
    CitizenSpeedVariation   [2]float64       // [min, max] per-citizen speed variation
    CitizenPathPreference   string           // "shortest" | "safest" | "balanced"
    HazardIntervalRange     [2]int           // [min, max] ticks between emergence
    MaxHazards              int              // Maximum concurrent hazards
    HazardSpreadRateRange   [2]float64       // [min, max] cells per tick
    HazardDurationRange     [2]int           // [min, max] ticks
    HazardTypeNames         []string         // Names from built-in registry for weighted random selection
    SafeZoneRadiusRange            [2]int    // [min, max] radius in cells
    SafeZonePlacement               string    // "far_from_start" | "random" | "corners" (initial + dynamic zones)
    SafeZoneEmergenceIntervalRange  [2]int    // [min, max] ticks between dynamic safe zone appearances
    ObstacleCountRange      [2]int           // [min, max] obstacles to generate
    ObstacleSizeRange       SizeRange        // Width/height ranges
    ObstacleTypes           []string         // Names from built-in registry for random assignment
}

type SizeRange struct {
    Width  [2]int // [min, max]
    Height [2]int // [min, max]
}
```

## Citizen

```go
type Citizen struct {
    ID             string         // UUID
    SimulationID   string
    StartPos       Position
    CurrentPos     Position       // Optional cached projection; if present must equal CurrentPath[PathIndex] when path exists
    Status         CitizenStatus  // navigating, escaped, dead
    Speed          int            // Cells per tick (from config or per-citizen)
    Autonomy       AutonomyProfile
    CurrentPath    []Position     // Planned path to nearest safe zone
    PathIndex      int            // Current step along path (canonical for derived position)
    KilledBy       *string        // Hazard ID if status == dead
    EscapedAt      *time.Time
    EscapedToZone  *string        // SafeZone ID
}

type CitizenStatus string

const (
    CitizenIdle       CitizenStatus = "idle"
    CitizenNavigating CitizenStatus = "navigating"
    CitizenEscaped    CitizenStatus = "escaped"
    CitizenDead       CitizenStatus = "dead"
)
```

**State transitions**: `idle → navigating → escaped | dead`

- `idle`: Pre-hazard phase; moves randomly each tick using open adjacent cells, does not pathfind to safe zones yet
- `navigating`: Active, following path toward safe zone, may recalculate
- `escaped`: Reached a safe zone
- `dead`: Overtaken by a hazard

**Validation**:
- StartPos must be within grid bounds
- StartPos must not overlap with a static obstacle
- Speed must be >= 1

## AutonomyProfile

```go
type AutonomyProfile struct {
    RiskTolerance float64 // 0.0 = avoid all hazards (prefer long safe path), 1.0 = ignore hazards (prefer shortest)
    SpeedVariation float64 // 0.0 = constant speed, >0 = random variation ±%
    PathPreference string  // "shortest" | "safest" | "balanced"
}
```

## Hazard

```go
type Hazard struct {
    ID            string        // UUID
    SimulationID  string
    Type          HazardType
    Origin        Position
    CurrentRadius float64       // Expands over time
    InitialRadius float64
    MaxRadius     float64
    SpreadRate    float64       // Cells per tick
    Severity      float64       // 0.0–1.0 (affects citizen death threshold)
    Duration      int           // Ticks before expiry
    CreatedTick   uint64        // Tick when hazard emerged
}

type HazardKind string

const (
    HazardKindExpanding HazardKind = "expanding" // fire, flood, lava
    HazardKindStrike    HazardKind = "strike"    // lightning (future)
    HazardKindGlobal    HazardKind = "global"    // earthquake (future)
)

type HazardType struct {
    Name string     // e.g., "fire", "flood", "lava"
    Kind HazardKind // governs tick behavior
}

Built-in hazard types are defined in app code as a registry (map[string]HazardType).
Config references them by name via HazardTypeNames in SimulationConfig.
Example built-in registry:

var HazardTypes = map[string]HazardType{
    "fire":  {Name: "fire",  Kind: HazardKindExpanding},
    "flood": {Name: "flood", Kind: HazardKindExpanding},
    "lava":  {Name: "lava",  Kind: HazardKindExpanding},
}
```

**Lifecycle**: Hazards are created active at the tick they emerge and are removed from the simulation when their duration expires. No intermediate state machine.

**Validation**:
- Origin must be within grid bounds
- Spread rate must be >= 0
- Duration must be > 0

## SafeZone

```go
type SafeZone struct {
    ID          string   // UUID
    Position    Position
    Radius      int      // In grid cells
}
```

**Lifecycle**: The simulation always starts with exactly one safe zone placed before tick 0. Additional safe zones may emerge dynamically at a configurable interval during the simulation (similar to hazard emergence, but zones persist indefinitely once placed). When a new safe zone appears, all `navigating` citizens recalculate their path toward the nearest safe zone.

**Validation**:
- Must be within grid bounds
- Must not overlap with static obstacles

## StaticObstacle

```go
type StaticObstacle struct {
    ID       string   // UUID
    Position Position
    Width    int      // In grid cells
    Height   int      // In grid cells
    Type     string   // References built-in obstacle type registry, e.g., "building", "terrain"
}
```

**Validation**:
- Must be within grid bounds
- Must not overlap with safe zones

## Environment (value object / config component)

```go
type Environment struct {
    Width     int
    Height    int
    Obstacles []StaticObstacle
    SafeZones []SafeZone
}
```

## Grid (runtime construct)

```go
type CellType int

const (
    CellOpen     CellType = 0  // Passable
    CellObstacle CellType = 1  // Blocked by static obstacle
    CellHazard   CellType = 2  // Blocked by hazard zone
    CellSafeZone CellType = 3  // Goal area
)

type Grid struct {
    Width  int
    Height int
    Cells  [][]CellType // [y][x] indexed
}
```

## Position

```go
type Position struct {
    X int
    Y int
}
```

## SimulationEvent

```go
type SimulationEvent struct {
    ID           string    // UUID
    SimulationID string
    Timestamp    time.Time
    Tick         uint64
    EventType    EventType
    EntityID     string    // Citizen, Hazard, or Simulation ID
    Payload      []byte    // JSON-encoded event-specific data
}

type EventType string

const (
    EventSimulationCreated   EventType = "simulation.created"
    EventSimulationPaused    EventType = "simulation.paused"
    EventSimulationStopped   EventType = "simulation.stopped"
    EventSimulationCompleted EventType = "simulation.completed"
    EventCitizenMoved        EventType = "citizen.moved"
    EventCitizenEscaped      EventType = "citizen.escaped"
    EventCitizenDied         EventType = "citizen.died"
    EventCitizenRecalculated EventType = "citizen.recalculated"
    EventCitizenAlerted      EventType = "citizen.alerted"
    EventSafeZoneEmerged     EventType = "safezone.emerged"
    EventHazardEmerged       EventType = "hazard.emerged"
    EventHazardExpanded      EventType = "hazard.expanded"
    EventHazardDissipated    EventType = "hazard.dissipated"
)
```

## Future Considerations

These are not implemented in v1 but are documented for potential later enhancement.

### Safe Zone Capacity

The v1 model treats safe zones as infinite-capacity areas (complementary to the dynamic emergence mechanic — citizens have more options over time, but no slot contention). A future iteration could add:

- `MaxOccupants int` to `SafeZone` — maximum number of citizens that can occupy the zone
- `OccupantIDs []string` — tracking which citizens are currently inside
- Behavior: when a safe zone reaches capacity, citizens targeting it must recalculate toward other safe zones (or wait for a spot if a citizen inside dies or another escapes later)

This would add strategic depth — citizens compete for finite safety — but was deferred to keep Slice 4 (Safe Zones + Death) focused on the basic entry/detection mechanic. Best added as a discrete enhancement after Slice 4 is complete. Dynamic emergence and capacity are fully complementary: the former creates more options over time, the latter incentivizes spreading out across them.

---

## Edge Case Handling

| Scenario | Behavior |
|---|---|
| All paths blocked to all safe zones | Citizens recalculate each tick; when no path exists, they stay in place until either a path opens or hazard overtakes them |
| Hazard completely envelops a safe zone | Safe zone becomes unreachable; citizens still targeting it will recalculate toward other safe zones |
| All citizens die before any escape | Simulation ends when last citizen dies (FR-009) |
| Citizen cannot outpace hazard | If hazard radius reaches citizen position before citizen reaches safe zone → citizen marked dead (FR-008) |
| Idle citizen when first hazard emerges | All idle citizens immediately switch to navigating and pathfind to safe zones |
| Zero hazard interval | No hazards emerge; citizens idle until `HazardInterval` ticks elapse, then auto-switch to navigating |
| Citizen seeded inside hazard | Citizen dies immediately (FR-008) — idle or not |
| Citizen with no valid path at start | Idle movement still occurs within valid cells; when first hazard emerges, tries to pathfind and becomes trapped |
| Zero citizens | Simulation starts and immediately completes (FR-009: all citizens either escaped or dead — vacuously true) |
| Hazard on top of citizen at emergence | Citizen dies on that tick (FR-008) |
| Safe zone enveloped before citizens reach | Citizens recalculate (FR-007); if all safe zones blocked, citizens are trapped |
| New safe zone appears mid-simulation | Citizens recalculate toward the nearest safe zone (including the new one); may cause path shifts mid-route |
| All safe zones are unreachable | Citizens recalculate each tick; when no path exists, they stay in place until either a path opens or hazard overtakes them |
| Idle citizen with no unblocked adjacent cell | Citizen stays in place for that tick (all directions blocked by obstacles or grid boundary) |
