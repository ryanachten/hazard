# Implementation Plan: Hazard Simulation System

**Branch**: `001-hazard-sim-system` | **Date**: 2026-06-13 | **Spec**: `specs/001-hazard-sim-system/spec.md`
**Input**: Feature specification from `specs/001-hazard-sim-system/spec.md`

## Summary

A 2D grid-based hazard simulation system where citizens navigate around static obstacles and expanding hazard zones to reach safe zones. Built with Go + Kafka for event-driven architecture, featuring real-time WebSocket visualization. Developed in 10 progressive slices, each independently testable, to build Go proficiency while exploring event streaming and pathfinding algorithms.

## Technical Context

**Language/Version**: Go 1.26+ (per constitution)  
**Primary Dependencies**: Apache Kafka (event backbone), Go Kafka client — `github.com/twmb/franz-go` (per research), Go WebSocket library — `github.com/coder/websocket` (per research), no external pathfinding/graph library (per constitution)  
**Storage**: Kafka topics for event stream (persisted for replay); no RDBMS for v1  
**Testing**: Go standard `testing` package, table-driven tests, `go test -race` (per constitution)  
**Target Platform**: macOS/Linux development machine, web browser (visualization)  
**Project Type**: CLI (simulation operator) + Web service (WebSocket visualization server)  
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
Kafka as event backbone. Components communicate through events. Event schemas versioned and backward-compatible. Consumers independently testable with simulated event streams.

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
# Single project structure for the Go monorepo
internal/
├── engine/              # Core simulation engine (tick loop, state management)
│   ├── simulation.go
│   ├── citizen.go
│   ├── hazard.go
│   └── environment.go
├── pathfinding/         # A* and alternative pathfinding implementations
│   ├── astar.go
│   └── interface.go
├── messaging/           # Kafka producer/consumer wrappers
│   ├── producer.go
│   └── consumer.go
├── events/              # Event type definitions and serialization
│   └── events.go
├── vis/                 # WebSocket server, client state broadcast
│   ├── hub.go
│   └── client.go
└── config/              # Simulation configuration loading
    └── config.go

cmd/
├── simctl/              # CLI for operator (start, pause, stop, configure)
│   └── main.go
└── simviz/              # Web server for visualization
    └── main.go

web/                     # Frontend visualization (HTML/CSS/JS)
├── index.html
├── canvas.js
└── style.css
```

**Structure Decision**: Standard Go monorepo layout with `cmd/` for entry points and `internal/` for private packages. Frontend visualization lives in `web/` as static assets served by the `simviz` binary.

## Implementation Slices

Each slice is independently testable and introduces one new concept. Developed in order — no skipping ahead until the current slice is working.

| # | Slice | What You Build | Go Concepts | Verified By |
|---|---|---|---|---|
| 1 | Grid + A* Pathfinding | 2D grid, A* algorithm, path reconstruction | Structs, slices, interfaces, unit tests, `go test` | Path from (0,0) to (5,5) avoiding obstacles |
| 2 | Citizen Movement | Citizens follow paths, tick loop | Methods, tick timers, state mutation, `go vet` | Citizens move toward goals each tick |
| 3 | Hazards + Envelopment | Emergency emergence, radius expansion, grid blocking | Concurrent state, config-driven behavior, edge cases | Hazard cells block pathfinding, radius grows |
| 4 | Safe Zones + Death | Citizens escape or die, simulation completion | State machine, event emission, multiple termination conditions | Simulation ends when all citizens resolved |
| 5 | Kafka Integration | Produce/consume simulation events | `franz-go` producer/consumer, `docker-compose`, event serialization | Events appear in Kafka topic |
| 6 | WebSocket Broadcast | Stream events to browser clients | `coder/websocket`, hub-and-spoke pattern, HTTP integration | Multiple browsers see same events |
| 7 | HTML Canvas Viz | Real-time browser visualization | Canvas 2D API, JSON parsing, requestAnimationFrame | Citizens and hazards rendered and moving |
| 8 | CLI Controls | `simctl` start, pause, stop, status | `flag` package, signal handling, command pattern | `simctl start --config x.json` runs simulation |
| 9 | Event History | Full event replay from Kafka topic | Consumer groups, offset management, event sourcing order | Replayed events match original sequence |
| 10 | Autonomy + Performance | Risk tolerance, path preference, 100+ citizens | A* variants (weighted), benchmarking, `pprof` | Citizens take different paths, 100 citizens at <2x real-time |

## Complexity Tracking

> Not applicable — all gates pass without violations.
