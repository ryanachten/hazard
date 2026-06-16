---
description: "Task list for Hazard Simulation System - 9 progressive slices for learning Go, Kafka, and event-driven architecture"
---

# Tasks: Hazard Simulation System

**Input**: Design documents from `specs/001-hazard-sim-system/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Learning Path**: 9 progressive slices, each introducing new Go concepts. Build in order — no skipping ahead.

> **Note**: Old Phases 1 (Setup) and 2 (Foundational types in isolation) were collapsed into Phase 1. Empty directory structures and Docker Compose for Kafka add no value early. Types are defined alongside the code that uses them.

**Organization**: Tasks are grouped by user story. Each user story is independently testable.

**Format**: `[ID] [P?] [Story] Description with file path`

---

## Phase 1: User Story 1 — Grid + A* Pathfinding (P1) 🎯 MVP SLICE 1/6

> **Note**: Old Phase 1 (empty setup) and Phase 2 (isolated type definitions) are folded into this phase. T004 (Position/CellType/Grid types) is merged into T008 below since they live in the same package. T005–T007 (entity structs, config, events) are deferred to their respective phases where they're first used.

**Goal**: Build a 2D grid and A* pathfinder. Citizens will later use this to navigate. This is the mathematical core of the system.

**Independent Test**: `FindPath(grid, {0,0}, {5,5})` returns a valid path avoiding obstacles, or nil when no path exists. Verify with `go test -v ./internal/pathfinding/`.

**Learning Outcome**: Go interfaces, method receivers, slices, `make`, table-driven tests

- [x] T008 [P] [US1] Define `Position`, `CellType`, `Grid` types and `Pathfinder` interface with `FindPath()` and `Name()` methods in `internal/pathfinding/interface.go`
- [x] T009 [P] [US1] Implement A* algorithm with Manhattan distance heuristic in `internal/pathfinding/astar.go` — cardinal-only movement, returns `[]Position` or nil
- [x] T010 [US1] Write table-driven tests in `internal/pathfinding/astar_test.go` testing: open grid path, obstacle avoidance, no-path case, start==goal edge case

**Checkpoint**: `go test -v ./internal/pathfinding/` passes. You understand interfaces, structs, and slices.

---

## Phase 2: User Story 1 — Citizen Movement (P1) 🎯 MVP SLICE 2/6

**Goal**: Citizens exist on the grid and follow paths toward destinations on each tick. This introduces state mutation over time.

**Independent Test**: Create a citizen with a pre-computed path, run N ticks, verify position matches expected point on the path.

**Learning Outcome**: Go methods on structs, state machines, iteration, tick-based simulation

- [x] T011 [US1] Implement `Citizen` struct with `Status` state machine (`idle`, `navigating`, `escaped`, `dead`) and validation in `internal/engine/citizen.go`
- [x] T012 [US1] Implement `Simulation` struct with `NewSimulation()`, `Tick()`, and state management in `internal/engine/simulation.go`
- [x] T013 [US1] Implement citizen path following: on each tick, advance `PathIndex` toward goal and ensure current position is consistent (either derive from `CurrentPath` + `PathIndex`, or update cached `CurrentPos`) in `internal/engine/citizen.go`
- [x] T014 [US1] Write tests in `internal/engine/simulation_test.go` verifying: citizen advances along path each tick, citizen stops at goal, multiple citizens move independently

**Checkpoint**: `go test -v ./internal/engine/` passes. You understand state mutation and tick loops.

---

## Phase 3: User Story 1 — Hazards + Envelopment (P1) 🎯 MVP SLICE 3/6

**Goal**: Hazards emerge at scheduled ticks and expand their radius over time, blocking grid cells and forcing citizens to recalculate.

**Independent Test**: Create a hazard at a known position with known spread rate, run N ticks, verify radius matches expected value and hazard cells block pathfinding.

**Learning Outcome**: Concurrent state (via sequential tick), config-driven behavior, floating-point radius → grid-cell mapping

- [ ] T015 [P] [US1] Implement `Hazard` struct with lifecycle state machine (`scheduled`, `active`, `dissipated`) and `HazardType` in `internal/engine/hazard.go`
- [ ] T016 [US1] Implement hazard emergence scheduling and radius expansion logic in `Simulation.Tick()` in `internal/engine/simulation.go`
- [ ] T017 [US1] Implement hazard-cell-to-Grid integration: mark cells within hazard radius as `CellHazard` so pathfinding avoids them, in `internal/engine/simulation.go`
- [ ] T018 [US1] Write tests in `internal/engine/hazard_test.go` verifying: radius grows each tick, hazard cells block pathfinding, hazard dissipates after duration

**Checkpoint**: `go test -v ./internal/engine/` passes. You understand float-to-grid mapping and lifecycle management.

---

## Phase 4: User Story 1 — Safe Zones + Death (P1) 🎯 MVP SLICE 4/6

**Goal**: Citizens can reach safe zones (escape) or be overtaken by hazards (die). Simulation ends when all citizens are resolved.

**Independent Test**: Create a simulation with one citizen, one safe zone, one hazard. Verify the citizen either escapes or dies based on distance/speed. Verify simulation auto-completes.

**Learning Outcome**: Multiple termination conditions, collision detection, simulation completion detection

- [ ] T019 [P] [US1] Implement `SafeZone` struct and environment generation (safe zone placement, obstacle placement, citizen seeding) in `internal/engine/environment.go`
- [ ] T020 [US1] Implement escape detection (citizen position within safe zone radius) and death detection (hazard radius reaches citizen position) in `Simulation.Tick()`
- [ ] T021 [US1] Implement simulation completion: detect when all citizens are `escaped` or `dead`, set simulation state to `completed` in `internal/engine/simulation.go`
- [ ] T022 [US1] Write tests in `internal/engine/simulation_test.go` verifying: citizen reaches safe zone and escapes, citizen overtaken by hazard dies, simulation completes when all resolved

**Checkpoint**: `go test -v ./internal/engine/` passes. You understand state machines and multiple termination flows.

---

## Phase 5: User Story 1 — CLI Controls (P1) 🎯 MVP SLICE 5/6

**Goal**: An operator can start, pause, resume, stop, and check status of a simulation via a CLI tool.

**Independent Test**: Run `go build ./cmd/simctl && ./simctl start --config examples/simple-sim.json` and verify simulation runs, then `./simctl status` shows progress.

**Learning Outcome**: Go `flag` package, JSON config loading, signal handling, CLI patterns

- [ ] T023 [US1] Create `cmd/simctl/main.go` with command structure and `flag`-based subcommand parsing for `start`, `pause`, `resume`, `stop`, `status`
- [ ] T024 [US1] Implement config loading from JSON file path in `internal/config/config.go` — read file, unmarshal into `SimulationConfig`, validate required fields
- [ ] T025 [US1] Implement simulation lifecycle commands: `start` kicks off tick loop in goroutine, `pause`/`resume` toggle state, `stop` terminates, `status` prints summary
- [ ] T026 [US1] Add `examples/simple-sim.json` with a minimal working config (1 citizen, 1 safe zone, 1 hazard, small grid) and verify end-to-end: build, run, status, stop

**Checkpoint**: `go build ./cmd/simctl` succeeds. `./simctl start --config examples/simple-sim.json` runs a simulation. You understand CLI construction and JSON config.

---

## Phase 6: User Story 1 — Event Emission (P1) 🎯 MVP SLICE 6/6

**Goal**: The simulation emits structured events for every state change, stored in-memory for now. This is the bridge to Kafka and WebSocket in later phases.

**Independent Test**: Run a simulation and verify that events are produced for: simulation start, citizen moves, citizen escapes/deaths, hazard emergence/expansion, simulation completion.

**Learning Outcome**: Go event emission patterns, `time.Time`, UUID generation, structured logging

- [ ] T027 [P] [US1] Implement event emission helpers in `internal/events/events.go` — constructor functions for each event type that generate IDs and timestamps
- [ ] T028 [US1] Integrate event emission into `Simulation.Tick()`: emit events for citizen moves, hazard expansions, state changes in `internal/engine/simulation.go`
- [ ] T029 [US1] Store emitted events in-memory on the `Simulation` struct and add a `Events()` accessor for later retrieval
- [ ] T030 [US1] Write tests in `internal/engine/events_test.go` verifying: each tick produces expected events, event ordering is correct, completed simulation has complete event log

**Checkpoint**: `go test -v ./internal/engine/` passes with event tests. US1 is now complete — you have a fully functional simulation with CLI control and event output. **You've built a complete Go application from scratch.**

---

## Phase 7: User Story 2 — WebSocket Broadcast + Canvas Viz (P2) 🌐

**Goal**: Simulation events stream to browser clients in real-time via WebSocket, rendered on an HTML Canvas.

**Independent Test**: Start `simviz`, connect a browser to `http://localhost:8080`, start a simulation via `simctl`, verify citizens and hazards render and update in real-time.

**Learning Outcome**: Go WebSocket (`coder/websocket`), hub-and-spoke pattern, Canvas 2D API, frontend WebSocket client

- [ ] T031 [P] [US2] Implement WebSocket hub in `internal/vis/hub.go` — register/unregister clients, broadcast JSON events to all connected clients
- [ ] T032 [P] [US2] Implement WebSocket client handler in `internal/vis/client.go` — accept upgrade, manage read/write lifecycle, handle disconnect
- [ ] T033 [US2] Create `cmd/simviz/main.go` — HTTP server that serves `web/` static files and upgrades WebSocket connections at `/ws`, subscribes to simulation events
- [ ] T034 [P] [US2] Create `web/index.html` with canvas element and script/style includes
- [ ] T035 [US2] Create `web/canvas.js` — WebSocket client that connects to `ws://localhost:8080/ws`, parses JSON events, renders grid/citizens/hazards/safe zones on Canvas 2D using `requestAnimationFrame`
- [ ] T036 [US2] Create `web/style.css` — basic layout, dark background, centered canvas
- [ ] T037 [US2] Integration test: run `simviz`, connect mock WebSocket client, emit events, verify client receives and renders them

**Checkpoint**: You can SEE your simulation running in a browser. You understand WebSocket protocol, hub-spoke pattern, and Canvas rendering.

---

## Phase 8: User Story 3 — Kafka Integration + Event History (P3) 📡

**Goal**: Simulation events are published to Kafka and can be replayed after the simulation ends.

**Independent Test**: Start Kafka via docker-compose, run a simulation, consume events from the `simulation-events` topic, verify complete ordered sequence.

**Learning Outcome**: `franz-go` Kafka client, event streaming, consumer groups, offset management, `context.Context`

- [ ] T038 [P] [US3] Implement Kafka producer in `internal/messaging/producer.go` — connect to broker, produce JSON-serialized `SimulationEvent` to `simulation-events` topic
- [ ] T039 [P] [US3] Implement Kafka consumer in `internal/messaging/consumer.go` — subscribe to topic, consume events in order with configurable consumer group
- [ ] T040 [US3] Integrate Kafka producer into simulation tick loop — emit events to Kafka instead of (or in addition to) in-memory storage in `internal/engine/simulation.go`
- [ ] T041 [US3] Implement event replay in `cmd/simctl/main.go` — add `replay` command that consumes events from Kafka and prints or saves them in order
- [ ] T042 [US3] Write tests in `internal/messaging/producer_test.go` and `consumer_test.go` — use a mock Kafka or verify serialization/deserialization round-trip

**Checkpoint**: Events flow through Kafka. You can replay a simulation from the event stream. You understand producers, consumers, and event sourcing.

---

## Phase 9: User Story 4 — Autonomy + Performance (P3) ⚡

**Goal**: Citizens have varied behaviors (risk tolerance, path preference) and the system handles 100+ citizens at <2x real-time.

**Independent Test**: Run a simulation with 100 citizens, 5 hazards, obstacles — verify it completes within 2x real-time and citizens show varied path choices.

**Learning Outcome**: Weighted A*, benchmarking (`go test -bench`), `pprof`, algorithm optimization

- [ ] T043 [P] [US4] Implement weighted A* variant in `internal/pathfinding/astar.go` — accept a weight function to penalize cells near hazards for "safest" path preference
- [ ] T044 [US4] Implement autonomy profile integration — use citizen's `RiskTolerance`, `SpeedVariation`, and `PathPreference` to influence movement and path selection in `internal/engine/citizen.go`
- [ ] T045 [US4] Add benchmarks in `internal/pathfinding/astar_bench_test.go` and `internal/engine/simulation_bench_test.go` — measure pathfinding time for 100 citizens, measure simulation tick time
- [ ] T046 [US4] Optimize: profile with `pprof`, identify bottlenecks, optimize hot paths (grid cell allocation, path copying, hazard cell marking) to meet <2x real-time target

**Checkpoint**: `go test -bench=. -benchtime=10x ./...` shows acceptable performance. 100 citizens navigate with varied behaviors.

---

## Phase 10: Polish & Cross-Cutting Concerns ✨

**Purpose**: Final quality pass, documentation, and end-to-end validation

- [ ] T047 Run full validation: `gofmt -s -w . && go vet ./... && staticcheck ./... && go test -race ./...` — fix all issues
- [ ] T048 Validate all quickstart.md steps work end-to-end: Kafka up → simctl start → events in Kafka → simviz → browser shows simulation

---

## Dependencies & Execution Order

### Phase Dependencies

```
Phase 1: US1 Grid+A* — no dependencies (types defined here)
    ↓
Phase 2: US1 Movement — depends on Phase 1 (uses Grid + A*)
    ↓
Phase 3: US1 Hazards — depends on Phase 2 (uses Citizen movement)
    ↓
Phase 4: US1 Safe Zones — depends on Phase 3 (uses Hazards)
    ↓
Phase 5: US1 CLI — depends on Phase 4 (uses full simulation)
    ↓
Phase 6: US1 Events — depends on Phase 5 (CLI needs events)
    ↓
Phase 7: US2 WebSocket+Viz — depends on Phase 6 (needs events)
    ↓
Phase 8: US3 Kafka — depends on Phase 6 (needs events)
    ↓
Phase 9: US4 Autonomy — depends on Phase 4 (needs full simulation)
    ↓
Phase 10: Polish — depends on all phases
```

### User Story Dependencies

- **US1 (P1)**: Core simulation — no dependencies on other stories. Must be 100% complete first.
- **US2 (P2)**: WebSocket/Viz — depends on US1 event emission (Phase 6). Events flow to WebSocket.
- **US3 (P3)**: Kafka/History — depends on US1 event emission (Phase 6). Events flow to Kafka.
- **US4 (P3)**: Autonomy/Scale — depends on US1 core engine (Phase 4). Extends citizen behavior.

### Within Each User Story Phase

- Tasks marked `[P]` can run in parallel (different files, no dependencies)
- Tasks without `[P]` are sequential and build on each other
- Core implementation before integration
- Story complete before moving to next priority

### Parallel Opportunities

- Phase 1 (US1/Grid): T008, T009 can run in parallel (interface + implementation)
- Phase 3 (US1/Hazards): T015 is independent
- Phase 5 (US1/CLI): All tasks are sequential
- Phase 7 (US2/Viz): T031, T032, T034, T035, T036 can all run in parallel
- Phase 8 (US3/Kafka): T038, T039 can run in parallel
- Phase 9 (US4/Autonomy): T043 is independent

---

## Parallel Example: User Story 2 (WebSocket + Viz)

```bash
# Launch all parallel tasks for US2 together:
# Task T031 + T032: vis package (hub + client)
# Task T034 + T035 + T036: web/ assets (HTML + JS + CSS)

# Then assemble:
# Task T033: simviz server (depends on all of the above)
# Task T037: integration test
```

---

## Implementation Strategy

### MVP Scope (User Story 1 Only)

The MVP covers Phases 1-6 (T008-T030). This delivers:
- A working simulation with citizens, hazards, safe zones
- A* pathfinding with obstacle avoidance
- Event emission in memory
- CLI controls (start, pause, stop, status)
- All testable via `go test ./...`

**Do NOT skip ahead to US2/US3/US4 until all 6 slices of US1 are complete.**

### Learning Milestones

| Slice | Go Concepts Mastered | Milestone |
|-------|---------------------|-----------|
| 1 | structs, slices, interfaces, table-driven tests | `go test -v ./internal/pathfinding/` passes |
| 2 | methods, state mutation, tick loops | Citizens move step-by-step each tick |
| 3 | config-driven behavior, lifecycle management | Hazards emerge and expand on schedule |
| 4 | termination conditions, edge cases | Simulation auto-completes |
| 5 | `flag` package, JSON unmarshal, signal handling | `./simctl start --config x.json` works |
| 6 | event emission, UUID, time | Complete event log for any run |
| 7 | WebSocket, hub pattern, Canvas 2D | Live browser visualization |
| 8 | Kafka producer/consumer, context | Events streaming through Kafka |
| 9 | weighted algorithms, benchmarks, pprof | 100 citizens at <2x real-time |

### Incremental Delivery

1. Complete Phase 1 → Grid + A* pathfinding ready
2. Complete Phases 1-6 → MVP: fully functional simulation with CLI
3. Add Phase 7 → Real-time visualization (huge win!)
4. Add Phase 8 → Kafka event streaming (core learning goal)
5. Add Phase 9 → Scale and autonomy (polish)

### Validation

After each phase, run:
```bash
gofmt -s -w .
go vet ./...
staticcheck ./...
go test -race ./...
```

Before committing, verify nothing is broken.
