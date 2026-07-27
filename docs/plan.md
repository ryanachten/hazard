# Implementation Plan: Hazard Simulation System

## Summary

A 2D grid-based hazard simulation system where citizens navigate around static obstacles and expanding hazard zones to reach safe zones. Built with Go + NATS/JetStream for event-driven architecture, featuring a real-time terminal TUI visualization. Developed in 11 progressive slices, each independently testable, to build Go proficiency while exploring event streaming and pathfinding algorithms.

## Technical Context

**Language/Version**: Go 1.26+ (per constitution)  
**Primary Dependencies**: NATS/JetStream (event backbone), Go NATS client — `github.com/nats-io/nats.go` (per research), Bubbletea TUI — `github.com/charmbracelet/bubbletea` + `github.com/charmbracelet/lipgloss` + `github.com/charmbracelet/bubbles`, no external pathfinding/graph library (per constitution)  
**Storage**: JetStream streams for event persistence (built-in with NATS); no RDBMS for v1  
**Testing**: Go standard `testing` package, table-driven tests, `go test -race` (per constitution)  
**Target Platform**: macOS/Linux development machine, terminal emulator (TUI visualization)
**NATS Dev Setup**: Single `nats-server` binary or `docker compose up -d` — no CGo, zero-config JetStream enabled by default in `nats-server v2.10+`  
**Project Type**: CLI (simulation operator) + TUI (terminal visualization)  
**Performance Goals**: 100 citizens + 10 hazards at <2x real-time (SC-001); state changes reflected within 1s (SC-003)  
**Constraints**: Single machine for v1 (per spec assumption); 2D grid-based; no external graph deps; `gofmt -s`, `go vet`, `staticcheck` must pass (per constitution)  
**Scale/Scope**: Single simulation at a time; 100+ citizens, 10+ hazards per simulation

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle I — Small Slice Delivery ✅
The spec already defines progressive phases: (1) basic movement + static obstacles + expanding hazards, (2) citizen death + safe zones, (3) hazard types + progressive envelopment, (4) citizen autonomy, (5) additional hazard kinds (strike, global). The implementation slices must match or subdivide these further.

### Principle II — Agent-Assisted Code Review ✅
User writes all production code; agents review. Code review feedback addressed before merge. Agents may generate tests, config, boilerplate but not feature logic.

### Principle III — Idiomatic Go ✅
All code passes `gofmt -s`, `go vet`, `staticcheck`. Proper error handling, interface usage, idiomatic naming. No external pathfinding libraries.

### Principle IV — Event-Driven Architecture ✅
NATS/JetStream as event backbone. Components communicate through events. Event schemas versioned and backward-compatible. Consumers independently testable with simulated event streams.

### Principle V — Pathfinding & Autonomy ✅
Self-contained pathfinding module with well-defined interface. Layered autonomy: reactive → deliberative → event-driven coordination. Algorithm selection documented with trade-off rationale.

**Result**: All gates pass. Proceeding to Phase 0.

## Project Structure

### Documentation (this feature)

```text
docs/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── contracts/           # Phase 1 output
└── tasks.md             # Phase 2 output (NOT created by this command)
```

### Source Code (repository root)

```text
internal/
├── common/              # Shared domain types (Citizen, Hazard, SafeZone, Configuration, utilities)
│   ├── citizen.go
│   ├── configuration.go
│   ├── hazard.go
│   ├── safezone.go
│   └── utilities.go
├── engine/              # Core simulation engine (tick loop, state management)
│   └── simulation.go
├── pathfinding/         # A* and alternative pathfinding implementations
│   ├── astar.go
│   ├── dijkstra.go
│   └── grid.go
├── events/              # Event types, event bus, and commands
│   ├── commands.go
│   ├── eventbus.go
│   └── events.go
├── tui/                 # Bubbletea TUI components
│   ├── model.go         # Main model, update, view
│   ├── events.go        # Event handlers for grid rendering
│   └── styles.go        # Lipgloss style definitions
```

Entry points (currently `main.go` at repository root; will migrate to `cmd/` in future phases):

```text
cmd/
├── simctl/              # CLI for operator (start, pause, stop, configure) — planned Phase 10
│   └── main.go
└── simviz/              # Terminal TUI for visualization — planned Phase 6+ migration
    └── main.go
```

**Structure Decision**: Standard Go monorepo layout with `cmd/` for entry points and `internal/` for private packages. Terminal visualization uses Bubbletea TUI framework. No frontend language or build pipeline needed — everything is pure Go.

## Implementation Slices

Each slice is independently testable and introduces one new concept. Developed in order — no skipping ahead until the current slice is working.

| # | Slice | What You Build | Go Concepts | Verified By |
|---|---|---|---|---|---|
| 1 | Grid + A* Pathfinding | 2D grid, A* algorithm, path reconstruction | Structs, slices, interfaces, unit tests, `go test` | Path from (0,0) to (5,5) avoiding obstacles |
| 2 | Citizen Movement | Citizens follow paths, tick loop | Methods, tick timers, state mutation, `go vet` | Citizens move toward goals each tick |
| 3 | Hazards + Envelopment | Hazard emergence, radius expansion, grid blocking | Concurrent state, config-driven behavior, edge cases | Hazard cells block pathfinding, radius grows |
| 4 | Safe Zones + Death | Citizens escape or die, simulation completion; dynamic safe zone emergence | State machine, multiple termination conditions, scheduled emergence | Simulation ends when all citizens resolved; new safe zones appear mid-run |
| 5 | Event Emission | Event type constructors, tick integration, in-memory storage | `time.Time`, UUID, event patterns | Complete event log for any run |
| 6 | Terminal TUI — Core | Bubbletea model/view/update, lipgloss grid rendering, keyboard controls (start/pause/quit) | External Go libraries, Elm architecture, `lipgloss` styling, event-driven UI | TUI shows live simulation, responds to keyboard input |
| 7 | Entity-Grid Occupancy | `CellCitizen` cell type, citizen grid marking, obstacle placement (`StaticObstacle` + `ObstacleConfig`), safe zone capacity (`MaxOccupants`/`OccupantIDs`), pathfinding respects citizen-occupied, obstacle, and full safe-zone cells | Grid cell state management, pathfinding integration with dynamic blocking, capacity-aware navigation, environment generation | Citizen cells and obstacles block pathfinding; safe zone refuses entry when full; citizens recalculate toward available zone |
| 8 | Config Sidebar | Split-pane layout (grid + config panel), live-adjustable `SimulationConfig` fields with focus cycling | `lipgloss.JoinHorizontal`, focus management, config-as-UI mapping | Config sidebar renders alongside grid, Tab/↑↓ adjust values |
| 9 | CLI Controls (P2) | `simctl` start, pause, stop, status | `flag` package, JSON config, signal handling | `simctl start --config x.json` runs simulation |
| 10 | Event Fan-Out + Optional Remote Observer (P2) | In-process event hub broadcasting to TUI + optional subscribers; optional HTTP/WebSocket for remote observation | Go channels for fan-out, hub-and-spoke pattern, optional `coder/websocket` | Events fanned out to TUI and optional remote client |
| 11 | NATS/JetStream Integration (P3) | Produce/consume simulation events | `nats.go` JetStream producer/consumer, `docker-compose`, event serialization | Events appear in JetStream stream |
| 12 | Autonomy + Performance (P3) | Risk tolerance, path preference, 100+ citizens | A* variants (weighted), benchmarking, `pprof` | Citizens take different paths, 100 citizens at <2x real-time |

## Future Enhancements

> Safe zone capacity was promoted from Future Enhancements to Phase 7 (Entity-Grid Occupancy).

## Complexity Tracking

> Not applicable — all gates pass without violations.
