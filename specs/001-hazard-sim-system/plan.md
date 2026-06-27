# Implementation Plan: Hazard Simulation System

**Branch**: `001-hazard-sim-system` | **Date**: 2026-06-13 | **Spec**: `specs/001-hazard-sim-system/spec.md`
**Input**: Feature specification from `specs/001-hazard-sim-system/spec.md`

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
specs/001-hazard-sim-system/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
└── tasks.md             # Phase 2 output (NOT created by this command)
```

### Source Code (repository root)

```text
internal/
├── engine/              # Core simulation engine (tick loop, state management)
│   ├── simulation.go
│   ├── citizen.go
│   ├── hazard.go
│   └── environment.go
├── pathfinding/         # A* and alternative pathfinding implementations
│   ├── astar.go
│   └── interface.go
├── messaging/           # NATS JetStream producer/consumer wrappers
│   ├── producer.go
│   └── consumer.go
├── events/              # Event type definitions and serialization
│   └── events.go
├── tui/                 # Bubbletea TUI components
│   ├── model.go         # Main model, update, view
│   ├── grid_view.go     # Grid rendering with lipgloss
│   ├── controls.go      # Keybinding helpers
│   └── styles.go        # Lipgloss style definitions
└── config/              # Simulation configuration loading
    └── config.go

cmd/
├── simctl/              # CLI for operator (start, pause, stop, configure)
│   └── main.go
└── simviz/              # Terminal TUI for visualization
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
| 6 | Terminal TUI (Bubbletea) | Bubbletea model/view/update, lipgloss grid rendering, keyboard controls (start/pause/quit) | External Go libraries, Elm architecture, `lipgloss` styling, event-driven UI | TUI shows live simulation, responds to keyboard input |
| 7 | CLI Controls | `simctl` start, pause, stop, status | `flag` package, JSON config, signal handling | `simctl start --config x.json` runs simulation |
| 8 | Event Fan-Out + Optional Remote Observer | In-process event hub broadcasting to TUI + optional subscribers; optional HTTP/WebSocket for remote observation | Go channels for fan-out, hub-and-spoke pattern, optional `coder/websocket` | Events fanned out to TUI and optional remote client |
| 9 | NATS/JetStream Integration | Produce/consume simulation events | `nats.go` JetStream producer/consumer, `docker-compose`, event serialization | Events appear in JetStream stream |
| 10 | Autonomy + Performance | Risk tolerance, path preference, 100+ citizens | A* variants (weighted), benchmarking, `pprof` | Citizens take different paths, 100 citizens at <2x real-time |

## Future Enhancements

The following are deferred from v1 but documented for potential later iteration:

- **Safe Zone Capacity**: Add `MaxOccupants` and `OccupantIDs` to `SafeZone`. Citizens targeting a full zone recalculate. Best implemented as a discrete enhancement after Slice 4 (Safe Zones + Death). Fully complementary with dynamic emergence (more zones + capacity limits = citizens spread out).

## Complexity Tracking

> Not applicable — all gates pass without violations.
