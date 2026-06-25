---
description: "Task list for Hazard Simulation System - 11 progressive slices for learning Go, NATS/JetStream, and event-driven architecture"
---

# Tasks: Hazard Simulation System

**Input**: Design documents from `specs/001-hazard-sim-system/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Learning Path**: 11 progressive slices, each introducing new Go concepts. Build in order — no skipping ahead.

> **Note**: Old Phases 1 (Setup) and 2 (Foundational types in isolation) were collapsed into Phase 1. Empty directory structures and Docker Compose for NATS add no value early. Types are defined alongside the code that uses them.
>
> **Reorder**: Event emission (was Phase 6) and browser visualization (was Phase 7) are brought forward to Phases 5 and 6 to provide visual feedback earlier. CLI controls are deferred to Phase 7. The WebSocket upgrade (part of old Phase 7) becomes Phase 8.

**Organization**: Tasks are grouped by user story. Each user story is independently testable.

**Format**: `[ID] [P?] [Story] Description with file path`

---

## Phase 1: User Story 1 — Grid + A* Pathfinding (P1) 🎯 MVP SLICE 1/7

> **Note**: Old Phase 1 (empty setup) and Phase 2 (isolated type definitions) are folded into this phase. T004 (Position/CellType/Grid types) is merged into T008 below since they live in the same package. T005–T007 (entity structs, config, events) are deferred to their respective phases where they're first used.

**Goal**: Build a 2D grid and A* pathfinder. Citizens will later use this to navigate. This is the mathematical core of the system.

**Independent Test**: `FindPath(grid, {0,0}, {5,5})` returns a valid path avoiding obstacles, or nil when no path exists. Verify with `go test -v ./internal/pathfinding/`.

**Learning Outcome**: Go interfaces, method receivers, slices, `make`, table-driven tests

- [x] T008 [P] [US1] Define `Position`, `CellType`, `Grid` types and `Pathfinder` interface with `FindPath()` and `Name()` methods in `internal/pathfinding/interface.go`
- [x] T009 [P] [US1] Implement A* algorithm with Manhattan distance heuristic in `internal/pathfinding/astar.go` — cardinal-only movement, returns `[]Position` or nil
- [x] T010 [US1] Write table-driven tests in `internal/pathfinding/astar_test.go` testing: open grid path, obstacle avoidance, no-path case, start==goal edge case

**Checkpoint**: `go test -v ./internal/pathfinding/` passes. You understand interfaces, structs, and slices.

---

## Phase 2: User Story 1 — Citizen Movement (P1) 🎯 MVP SLICE 2/7

**Goal**: Citizens exist on the grid and follow paths toward destinations on each tick. This introduces state mutation over time.

**Independent Test**: Create a citizen with a pre-computed path, run N ticks, verify position matches expected point on the path.

**Learning Outcome**: Go methods on structs, state machines, iteration, tick-based simulation

- [x] T011 [US1] Implement `Citizen` struct with `Status` state machine (`idle`, `navigating`, `escaped`, `dead`) and validation in `internal/engine/citizen.go`
- [x] T012 [US1] Implement `Simulation` struct with `NewSimulation()`, `Tick()`, and state management in `internal/engine/simulation.go`
- [x] T013 [US1] Implement citizen path following: on each tick, advance `PathIndex` toward goal and ensure current position is consistent (either derive from `CurrentPath` + `PathIndex`, or update cached `CurrentPos`) in `internal/engine/citizen.go`
- [x] T014 [US1] Write tests in `internal/engine/simulation_test.go` verifying: citizen advances along path each tick, citizen stops at goal, multiple citizens move independently

**Checkpoint**: `go test -v ./internal/engine/` passes. You understand state mutation and tick loops.

---

## Phase 3: User Story 1 — Hazards + Envelopment (P1) 🎯 MVP SLICE 3/7

**Goal**: Hazards emerge at regular intervals and expand their radius over time, blocking grid cells and forcing citizens to recalculate.

**Independent Test**: Create a hazard at a known position with known spread rate, run N ticks, verify radius matches expected value and hazard cells block pathfinding.

**Learning Outcome**: Concurrent state (via sequential tick), config-driven behavior, floating-point radius → grid-cell mapping

- [x] T015 [P] [US1] Implement `Hazard` struct (active when created, removed on expiry) and `HazardType` in `internal/engine/hazard.go`
- [x] T016 [US1] Implement hazard emergence scheduling and radius expansion logic in `Simulation.Tick()` in `internal/engine/simulation.go`
- [x] T017 [US1] Implement hazard-cell-to-Grid integration: mark cells within hazard radius as `CellHazard` so pathfinding avoids them, in `internal/engine/simulation.go`
- [x] T018 [US1] Write tests in `internal/engine/hazard_test.go` verifying: radius grows each tick, hazard cells block pathfinding, hazard is removed after duration expires

**Checkpoint**: `go test -v ./internal/engine/` passes. You understand float-to-grid mapping and lifecycle management.

---

## Phase 4: User Story 1 — Safe Zones + Death (P1) 🎯 MVP SLICE 4/7

**Goal**: Citizens can reach safe zones (escape) or be overtaken by hazards (die). Additional safe zones emerge dynamically mid-simulation, triggering citizen path recalculation. Simulation ends when all citizens are resolved.

**Independent Test**: Create a simulation with one citizen, one safe zone, one hazard. Verify the citizen either escapes or dies based on distance/speed. Verify a new safe zone appears at the expected interval and triggers path recalculation. Verify simulation auto-completes.

**Learning Outcome**: Multiple termination conditions, collision detection, scheduled emergence, path recalculation triggers

- [x] T019 [P] [US1] Implement `SafeZone` struct and environment generation (initial safe zone placement, obstacle placement, citizen seeding) in `internal/engine/safe_zone.go`
- [x] T020 [US1] Implement escape detection (citizen position within safe zone radius) and death detection (hazard radius reaches citizen position) in `Simulation.Tick()`
- [x] T021 [US1] Implement dynamic safe zone emergence scheduling: at a configurable interval, place a new safe zone at a random valid position, mark its cells on the grid, and trigger citizen path recalculation toward the nearest zone
- [x] T022 [US1] Implement simulation completion: detect when all citizens are `escaped` or `dead`, set simulation state to `completed` in `internal/engine/simulation.go`
- [x] T023 [US1] Write tests in `internal/engine/simulation_test.go` verifying: citizen reaches safe zone and escapes, citizen overtaken by hazard dies, new safe zone appears on schedule, citizens recalculate toward nearest zone after emergence, simulation completes when all resolved

**Checkpoint**: `go test -v ./internal/engine/` passes. You understand state machines, scheduled emergence, and multiple termination flows.

**Future enhancements — Safe Zone Capacity + Dynamic Emergence**: After this phase is complete, consider adding `MaxOccupants`/`OccupantIDs` to `SafeZone`. Dynamic emergence creates more options over time; capacity limits incentivize citizens to spread out. See `data-model.md` Future Considerations for details.

---

## Phase 5: User Story 1 — Event Emission (P1) 🎯 MVP SLICE 5/7

**Goal**: The simulation emits structured events for every state change, stored in-memory. This is the bridge to visualization, CLI output, and NATS JetStream.

**Independent Test**: Run a simulation and verify that events are produced for: simulation start, citizen moves, citizen escapes/deaths, hazard emergence/expansion, simulation completion.

**Learning Outcome**: Go event emission patterns, `time.Time`, UUID generation, structured logging

- [ ] T024 [P] [US1] Implement event emission helpers in `internal/events/events.go` — constructor functions for each event type that generate IDs and timestamps
- [ ] T025 [US1] Integrate event emission into `Simulation.Tick()`: emit events for citizen moves, hazard expansions, safe zone emergence, state changes in `internal/engine/simulation.go`
- [ ] T026 [US1] Store emitted events in-memory on the `Simulation` struct and add an `Events()` accessor for later retrieval
- [ ] T027 [US1] Write tests in `internal/engine/events_test.go` verifying: each tick produces expected events, event ordering is correct, completed simulation has complete event log

**Checkpoint**: `go test -v ./internal/engine/` passes with event tests. Events capture every state change in the simulation.

---

## Phase 6: User Story 1 — Browser Visualization (P1) 🎯 MVP SLICE 6/7

**Goal**: Simulation state renders in a browser via Canvas 2D. An HTTP server serves events as JSON; the frontend polls and renders.

**Independent Test**: Start `simviz`, open `http://localhost:8080`, run a simulation, verify the browser shows citizens, hazards, and safe zones updating.

**Learning Outcome**: `net/http` JSON serving, `bun build` for TS→JS, Canvas 2D API, `requestAnimationFrame` render loop, polling pattern

- [ ] T028 [P] [US1] Create `web/index.html` with canvas element and script/style includes
- [ ] T029 [P] [US1] Create `web/src/canvas.ts` — polls `http://localhost:8080/state` each tick, parses JSON, renders grid/citizens/hazards/safe zones on Canvas 2D using `requestAnimationFrame`. Build with `bun build`
- [ ] T030 [P] [US1] Create `web/style.css` — basic layout, dark background, centered canvas
- [ ] T031 [US1] Implement `cmd/simviz/main.go` — HTTP server that serves `web/` static files and exposes a `/state` JSON endpoint returning simulation events and state
- [ ] T032 [US1] Wire end-to-end: verify simviz starts, HTTP endpoints respond, Canvas renders and updates

**Checkpoint**: You can SEE your simulation running in a browser. You understand `net/http`, JSON serialization, and Canvas rendering.

---

## Phase 7: User Story 1 — CLI Controls (P1) 🎯 MVP SLICE 7/7

**Goal**: An operator can start, pause, resume, stop, and check status of a simulation via a CLI tool.

**Independent Test**: Run `go build ./cmd/simctl && ./simctl start --config examples/simple-sim.json` and verify simulation runs, then `./simctl status` shows progress.

**Learning Outcome**: Go `flag` package, JSON config loading, signal handling, CLI patterns

- [ ] T033 [US1] Create `cmd/simctl/main.go` with command structure and `flag`-based subcommand parsing for `start`, `pause`, `resume`, `stop`, `status`
- [ ] T034 [US1] Implement config loading from JSON file path in `internal/config/config.go` — read file, unmarshal into `SimulationConfig`, validate required fields
- [ ] T035 [US1] Implement simulation lifecycle commands: `start` kicks off tick loop in goroutine, `pause`/`resume` toggle state, `stop` terminates, `status` prints summary
- [ ] T036 [US1] Add `examples/simple-sim.json` with a minimal working config (1 citizen, 1 safe zone, 1 hazard, small grid) and verify end-to-end: build, run, status, stop

**Checkpoint**: `go build ./cmd/simctl` succeeds. `./simctl start --config examples/simple-sim.json` runs a simulation. You understand CLI construction and JSON config.

---

## Phase 8: User Story 2 — WebSocket Upgrade (P2)

**Goal**: Replace HTTP polling in the visualization with WebSocket push for real-time updates.

**Independent Test**: Start `simviz`, connect a browser to `http://localhost:8080`, start a simulation, verify citizens and hazards render and update in real-time without polling.

**Learning Outcome**: Go WebSocket (`coder/websocket`), hub-and-spoke pattern

- [ ] T037 [P] [US2] Implement WebSocket hub in `internal/vis/hub.go` — register/unregister clients, broadcast JSON events to all connected clients
- [ ] T038 [P] [US2] Implement WebSocket client handler in `internal/vis/client.go` — accept upgrade, manage read/write lifecycle, handle disconnect
- [ ] T039 [US2] Update `cmd/simviz/main.go` — add WebSocket endpoint `/ws`, subscribe to simulation events, broadcast to connected clients
- [ ] T040 [US2] Update `web/src/canvas.ts` — connect to `ws://localhost:8080/ws`, switch from polling to event-driven rendering. Rebuild with `bun build`
- [ ] T041 [US2] Integration test: run `simviz`, connect mock WebSocket client, emit events, verify client receives them

**Checkpoint**: Visualization updates in real-time without polling. You understand WebSocket protocol and hub-spoke pattern.

---

## Phase 9: User Story 3 — NATS/JetStream Integration + Event History (P3) 📡

**Goal**: Simulation events are published to NATS JetStream and can be replayed after the simulation ends.

**Independent Test**: Start NATS via docker-compose, run a simulation, consume events from the `simulation-events` stream, verify complete ordered sequence.

**Learning Outcome**: `nats.go` JetStream client, event streaming, consumer management, ordered delivery, `context.Context`

- [ ] T042 [P] [US3] Implement JetStream producer in `internal/messaging/producer.go` — connect to NATS, create JetStream context, publish JSON-serialized `SimulationEvent` to `simulation-events.stream` subject
- [ ] T043 [P] [US3] Implement JetStream consumer in `internal/messaging/consumer.go` — subscribe to stream, consume events in order with configurable durable consumer name
- [ ] T044 [US3] Integrate JetStream producer into simulation tick loop — publish events to JetStream instead of (or in addition to) in-memory storage in `internal/engine/simulation.go`
- [ ] T045 [US3] Implement event replay in `cmd/simctl/main.go` — add `replay` command that creates a JetStream consumer and prints or saves events in order
- [ ] T046 [US3] Write tests in `internal/messaging/producer_test.go` and `consumer_test.go` — use an embedded NATS server (`github.com/nats-io/nats-server/v2/server`) for integration tests or verify serialization/deserialization round-trip

**Checkpoint**: Events flow through NATS JetStream. You can replay a simulation from the event stream. You understand JetStream producers, consumers, and event sourcing.

---

## Phase 10: User Story 4 — Autonomy + Performance (P3) ⚡

**Goal**: Citizens have varied behaviors (risk tolerance, path preference) and the system handles 100+ citizens at <2x real-time.

**Independent Test**: Run a simulation with 100 citizens, 5 hazards, obstacles — verify it completes within 2x real-time and citizens show varied path choices.

**Learning Outcome**: Weighted A*, benchmarking (`go test -bench`), `pprof`, algorithm optimization

- [ ] T047 [P] [US4] Implement weighted A* variant in `internal/pathfinding/astar.go` — accept a weight function to penalize cells near hazards for "safest" path preference
- [ ] T048 [US4] Implement autonomy profile integration — use citizen's `RiskTolerance`, `SpeedVariation`, and `PathPreference` to influence movement and path selection in `internal/engine/citizen.go`
- [ ] T049 [US4] Add benchmarks in `internal/pathfinding/astar_bench_test.go` and `internal/engine/simulation_bench_test.go` — measure pathfinding time for 100 citizens, measure simulation tick time
- [ ] T050 [US4] Optimize: profile with `pprof`, identify bottlenecks, optimize hot paths (grid cell allocation, path copying, hazard cell marking) to meet <2x real-time target

**Checkpoint**: `go test -bench=. -benchtime=10x ./...` shows acceptable performance. 100 citizens navigate with varied behaviors.

---

## Phase 11: Polish & Cross-Cutting Concerns ✨

**Purpose**: Final quality pass, documentation, and end-to-end validation

- [ ] T051 Run full validation: `gofmt -s -w . && go vet ./... && staticcheck ./... && go test -race ./...` — fix all issues
- [ ] T052 Validate all quickstart.md steps work end-to-end: NATS up → simctl start → events in JetStream → simviz → browser shows simulation

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
Phase 5: US1 Events — depends on Phase 4 (needs full simulation state)
    ↓
Phase 6: US1 Viz — depends on Phase 5 (needs events to display)
    ↓
Phase 7: US1 CLI — depends on Phase 5 (needs events for output)
    ↓
Phase 8: US2 WebSocket — depends on Phase 6 (replaces polling in viz)
    ↓
Phase 9: US3 NATS/JetStream — depends on Phase 5 (needs events)
    ↓
Phase 10: US4 Autonomy — depends on Phase 4 (needs full simulation)
    ↓
Phase 11: Polish — depends on all phases
```

### User Story Dependencies

- **US1 (P1)**: Core simulation — no dependencies on other stories. Must be 100% complete first. Phases 4–7 (Safe Zones, Events, Viz, CLI) complete US1.
- **US2 (P2)**: WebSocket/Viz — depends on US1 event emission (Phase 5) and Phase 6 viz scaffolding. Phase 8 upgrades the viz from HTTP polling to WebSocket.
- **US3 (P3)**: NATS/JetStream — depends on US1 event emission (Phase 5). Events flow to JetStream.
- **US4 (P3)**: Autonomy/Scale — depends on US1 core engine (Phase 4). Extends citizen behavior.

### Within Each User Story Phase

- Tasks marked `[P]` can run in parallel (different files, no dependencies)
- Tasks without `[P]` are sequential and build on each other
- Core implementation before integration
- Story complete before moving to next priority

### Parallel Opportunities

- Phase 1 (US1/Grid): T008, T009 can run in parallel (interface + implementation)
- Phase 3 (US1/Hazards): T015 is independent
- Phase 6 (US1/Viz): T028, T029, T030 can all run in parallel (web/ assets)
- Phase 7 (US1/CLI): All tasks are sequential
- Phase 8 (US2/WebSocket): T037, T038 can run in parallel
- Phase 9 (US3/NATS/JetStream): T042, T043 can run in parallel
- Phase 10 (US4/Autonomy): T047 is independent

---

## Parallel Example: Phase 6 (Browser Viz)

```bash
# Launch all parallel tasks for Phase 6 together:
# Task T028 + T029 + T030: web/ assets (HTML + JS + CSS)

# Then assemble:
# Task T031: simviz server (depends on all of the above)
# Task T032: end-to-end verification
```

---

## Implementation Strategy

### MVP Scope (User Story 1 Only)

The MVP covers Phases 1-7 (T008–T036). This delivers:
- A working simulation with citizens, hazards, safe zones
- A* pathfinding with obstacle avoidance
- Event emission in memory
- Browser visualization via HTTP polling
- CLI controls (start, pause, stop, status)
- All testable via `go test ./...`

**Do NOT skip ahead to US2/US3/US4 until all US1 phases are complete.**

### Learning Milestones

| Phase | Go Concepts Mastered | Milestone |
|-------|---------------------|-----------|
| 1 | structs, slices, interfaces, table-driven tests | `go test -v ./internal/pathfinding/` passes |
| 2 | methods, state mutation, tick loops | Citizens move step-by-step each tick |
| 3 | config-driven behavior, lifecycle management | Hazards emerge and expand on schedule |
| 4 | termination conditions, scheduled emergence, edge cases | Simulation auto-completes; safe zones appear mid-run |
| 5 | event emission, UUID, time | Complete event log for any run |
| 6 | `net/http` JSON serving, Canvas 2D | Browser shows live simulation |
| 7 | `flag` package, JSON unmarshal, signal handling | `./simctl start --config x.json` works |
| 8 | WebSocket, hub pattern, event push | Live updates without polling |
| 9 | JetStream producer/consumer, context | Events streaming through NATS JetStream |
| 10 | weighted algorithms, benchmarks, pprof | 100 citizens at <2x real-time |

### Incremental Delivery

1. Complete Phase 1 → Grid + A* pathfinding ready
2. Complete Phases 1–4 → Full simulation with escape/death/completion
3. Complete Phases 1–6 → Simulation with live browser visualization (**huge win!**)
4. Complete Phases 1–7 → Full US1 MVP with CLI
5. Add Phase 8 → Real-time WebSocket updates
6. Add Phase 9 → NATS JetStream event streaming (core learning goal)
7. Add Phase 10 → Scale and autonomy (polish)

### Validation

After each phase, run:
```bash
gofmt -s -w .
go vet ./...
staticcheck ./...
go test -race ./...
```

Before committing, verify nothing is broken.
