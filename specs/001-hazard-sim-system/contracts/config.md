# Simulation Configuration Contract

## File Format (JSON)

Simulations are configured via a JSON file passed to the CLI.

The config defines **generative rules** — ranges, counts, and constraints from which the simulation engine procedurally generates a concrete run at startup. This ensures each run (even with the same config) is varied and interesting. Pass the same `seed` to reproduce a specific run.

```json
{
  "environment": {
    "width": 100,
    "height": 100,
    "seed": 42
  },
  "citizens": {
    "count_range": [5, 15],
    "speed_range": [1, 3],
    "autonomy": {
      "risk_tolerance_range": [0.1, 0.6],
      "speed_variation_range": [0.05, 0.2],
      "path_preference": "balanced"
    }
  },
  "hazards": {
    "emergence_interval_range": [30, 70],
    "max_concurrent": 5,
    "spread_rate_range": [0.3, 0.7],
    "duration_range": [200, 400],
    "type_names": ["generic"]
  },
  "safe_zones": {
    "count_range": [1, 3],
    "radius_range": [2, 4],
    "placement": "far_from_start"
  },
  "obstacles": {
    "count_range": [1, 4],
    "size_range": {"width": [3, 6], "height": [3, 8]},
    "types": ["building"]
  },
  "simulation": {
    "tick_interval_ms": 100
  }
}
```

### Field Descriptions

| Field | Type | Description |
|---|---|---|
| `environment.width` | int | Grid width in cells |
| `environment.height` | int | Grid height in cells |
| `environment.seed` | int64 | RNG seed for reproducibility |
| `citizens.count_range` | [int, int] | Number of citizens to spawn, sampled uniformly |
| `citizens.speed_range` | [int, int] | Per-citizen speed sampled uniformly from this range (cells/tick) |
| `citizens.autonomy.risk_tolerance_range` | [float, float] | Per-citizen risk tolerance sampled uniformly from this range |
| `citizens.autonomy.speed_variation_range` | [float, float] | Per-citizen speed variation sampled uniformly from this range |
| `citizens.autonomy.path_preference` | string | "shortest", "safest", or "balanced" — may be overridden per citizen in future |
| `hazards.emergence_interval_range` | [int, int] | Ticks between hazard emergence, sampled uniformly per interval |
| `hazards.max_concurrent` | int | Maximum hazards active simultaneously |
| `hazards.spread_rate_range` | [float, float] | Per-hazard spread rate sampled uniformly (cells/tick) |
| `hazards.duration_range` | [int, int] | Per-hazard active duration sampled uniformly (ticks, 0 = indefinite) |
| `hazards.type_names` | [string] | Names of built-in hazard types for weighted random selection; visual and behavioral properties are defined in app code, not config |
| `safe_zones.count_range` | [int, int] | Number of safe zones to place, sampled uniformly |
| `safe_zones.radius_range` | [int, int] | Per-zone radius sampled uniformly |
| `safe_zones.placement` | string | Strategy: "far_from_start", "random", "corners" |
| `obstacles.count_range` | [int, int] | Number of obstacles to generate, sampled uniformly |
| `obstacles.size_range` | object | Width and height ranges sampled uniformly |
| `obstacles.types` | [string] | Pool of obstacle types for random assignment |
| `simulation.tick_interval_ms` | int | Milliseconds per simulation tick |

### Generation Phase

Before the first tick, the engine runs a generation pass:

1. **Seed RNG** with `environment.seed`
2. **Place safe zones** per `safe_zones.placement` strategy (validated: no overlap with each other, within bounds)
3. **Place obstacles** per `obstacles.count` and `size_range` (validated: no overlap with safe zones or each other)
4. **Spawn citizens** at random open positions (validated: not on obstacles or safe zones)
5. **Assign per-citizen attributes**: speed sampled from `speed_range`, autonomy values sampled from their respective ranges
6. **Schedule first hazard** at tick sampled from `emergence_interval_range`

All generation is deterministic given the same seed.

## CLI Command

```text
simctl start --config simulation.json
simctl pause
simctl resume
simctl stop
simctl status
```

## Simulation Lifecycle (CLI)

```text
simctl start --config config.json
  → Loads config, generates run, starts tick loop
  → Returns simulation ID

simctl pause
  → Suspends tick loop (state preserved)

simctl resume
  → Continues tick loop

simctl stop
  → Terminates simulation
  → All citizens marked dead on abrupt stop

simctl status
  → Prints current state summary (ticks elapsed, citizens alive/dead/escaped, hazards active)
```
