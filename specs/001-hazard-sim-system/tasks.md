---
description: "Task list for Hazard Simulation System - 12 progressive slices for learning Go, NATS/JetStream, and event-driven architecture"
---

# Tasks: Hazard Simulation System

**Input**: Design documents from `specs/001-hazard-sim-system/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Learning Path**: 12 progressive slices, each introducing new Go concepts. Build in order — no skipping ahead.

> **Note**: Old Phases 1 (Setup) and 2 (Foundational types in isolation) were collapsed into Phase 1. Empty directory structures and Docker Compose for NATS add no value early. Types are defined alongside the code that uses them.
>
> **Reorder**: Event emission was brought forward to Phase 5 to enable earlier testing. The TUI was split into two phases: core grid rendering (Phase 6) and config sidebar (Phase 8). Entity-grid occupancy was added as Phase 7 to close deferred design gaps (obstacle provisioning, citizen cell marking, safe zone capacity). CLI controls were deferred to Phase 10 (P2) after Event Fan-Out.

**Organization**: Tasks are grouped by user story. Each user story is independently testable.

**Format**: `[ID] [P?] [Story] Description with file path`

---

## Phase 1: User Story 1 — Grid + A* Pathfinding (P1) 🎯 MVP SLICE 1/8

> **Note**: Old Phase 1 (empty setup) and Phase 2 (isolated type definitions) are folded into this phase. T004 (Position/CellType/Grid types) is merged into T008 below since they live in the same package. T005–T007 (entity structs, config, events) are deferred to their respective phases where they're first used.

**Goal**: Build a 2D grid and A* pathfinder. Citizens will later use this to navigate. This is the mathematical core of the system.

**Independent Test**: `FindPath(grid, {0,0}, {5,5})` returns a valid path avoiding obstacles, or nil when no path exists. Verify with `go test -v ./internal/pathfinding/`.

**Learning Outcome**: Go interfaces, method receivers, slices, `make`, table-driven tests

- [x] T008 [P] [US1] Define `Position`, `CellType`, `Grid` types and `Pathfinder` interface with `FindPath()` and `Name()` methods in `internal/pathfinding/interface.go`
- [x] T009 [P] [US1] Implement A* algorithm with Manhattan distance heuristic in `internal/pathfinding/astar.go` — cardinal-only movement, returns `[]Position` or nil
- [x] T010 [US1] Write table-driven tests in `internal/pathfinding/astar_test.go` testing: open grid path, obstacle avoidance, no-path case, start==goal edge case

**Checkpoint**: `go test -v ./internal/pathfinding/` passes. You understand interfaces, structs, and slices.

---

## Phase 2: User Story 1 — Citizen Movement (P1) 🎯 MVP SLICE 2/8

**Goal**: Citizens exist on the grid and follow paths toward destinations on each tick. This introduces state mutation over time.

**Independent Test**: Create a citizen with a pre-computed path, run N ticks, verify position matches expected point on the path.

**Learning Outcome**: Go methods on structs, state machines, iteration, tick-based simulation

- [x] T011 [US1] Implement `Citizen` struct with `Status` state machine (`idle`, `navigating`, `escaped`, `dead`) and validation in `internal/engine/citizen.go`
- [x] T012 [US1] Implement `Simulation` struct with `NewSimulation()`, `Tick()`, and state management in `internal/engine/simulation.go`
- [x] T013 [US1] Implement citizen path following: on each tick, advance `PathIndex` toward goal and ensure current position is consistent (either derive from `CurrentPath` + `PathIndex`, or update cached `CurrentPos`) in `internal/engine/citizen.go`
- [x] T014 [US1] Write tests in `internal/engine/simulation_test.go` verifying: citizen advances along path each tick, citizen stops at goal, multiple citizens move independently

**Checkpoint**: `go test -v ./internal/engine/` passes. You understand state mutation and tick loops.

---

## Phase 3: User Story 1 — Hazards + Envelopment (P1) 🎯 MVP SLICE 3/8

**Goal**: Hazards emerge at regular intervals and expand their radius over time, blocking grid cells and forcing citizens to recalculate.

**Independent Test**: Create a hazard at a known position with known spread rate, run N ticks, verify radius matches expected value and hazard cells block pathfinding.

**Learning Outcome**: Concurrent state (via sequential tick), config-driven behavior, floating-point radius → grid-cell mapping

- [x] T015 [P] [US1] Implement `Hazard` struct (active when created, removed on expiry) and `HazardType` in `internal/engine/hazard.go`
- [x] T016 [US1] Implement hazard emergence scheduling and radius expansion logic in `Simulation.Tick()` in `internal/engine/simulation.go`
- [x] T017 [US1] Implement hazard-cell-to-Grid integration: mark cells within hazard radius as `CellHazard` so pathfinding avoids them, in `internal/engine/simulation.go`
- [x] T018 [US1] Write tests in `internal/engine/hazard_test.go` verifying: radius grows each tick, hazard cells block pathfinding, hazard is removed after duration expires

**Checkpoint**: `go test -v ./internal/engine/` passes. You understand float-to-grid mapping and lifecycle management.

---

## Phase 4: User Story 1 — Safe Zones + Death (P1) 🎯 MVP SLICE 4/8

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

## Phase 5: User Story 1 — Event Emission (P1) 🎯 MVP SLICE 5/8

**Goal**: The simulation emits structured events for every state change, stored in-memory. This is the bridge to visualization, CLI output, and NATS JetStream.

**Independent Test**: Run a simulation and verify that events are produced for: simulation start, citizen moves, citizen escapes/deaths, hazard emergence/expansion, simulation completion.

**Learning Outcome**: Go event emission patterns, `time.Time`, UUID generation, structured logging

- [x] T024 [P] [US1] Implement event emission helpers in `internal/events/events.go` — constructor functions for each event type that generate IDs and timestamps
- [x] T025 [US1] Integrate event emission into `Simulation.Tick()`: emit events for citizen moves, hazard expansions, safe zone emergence, state changes in `internal/engine/simulation.go`
- [x] T026 [US1] Store emitted events in-memory on the `Simulation` struct and add an `Events()` accessor for later retrieval
- [x] T027 [US1] Write tests in `internal/engine/events_test.go` verifying: each tick produces expected events, event ordering is correct, completed simulation has complete event log

**Checkpoint**: `go test -v ./internal/engine/` passes with event tests. Events capture every state change in the simulation.

---

## Phase 6: User Story 1 — Terminal TUI (Bubbletea) — Core (P1) 🎯 MVP SLICE 6/8

**Goal**: Simulation state renders in a terminal UI via Bubbletea. The TUI connects to the simulation engine through a Go channel, renders the grid using the design-language glyphs and colors (via Lipgloss), and accepts keyboard input for start/pause/quit.

**Independent Test**: Start `simviz`, verify the TUI renders a grid with citizens, hazards, and safe zones using the correct glyphs and colors. Verify keyboard controls work (Enter to start, Space to pause, q to quit).

**Learning Outcome**: External Go library integration, Bubbletea Elm architecture (Model/Update/View), Lipgloss composable styles, terminal event handling, `tea.Program`, alt-screen rendering

- [ ] T028 [P] [US1] Define event channel type in `internal/events/` — add `type EventChan chan SimulationEvent`, decide buffering strategy (e.g., 256-buffered), document producer/consumer roles. This is the bridge contract between engine and TUI.
- [ ] T029 [P] [US1] Define Bubbletea model struct — simulation state, grid dimensions, viewport offset, keyboard mode, plus `events.EventChan` read-end field — in `internal/tui/model.go`
- [ ] T030 [P] [US1] Implement grid view rendering — iterate over grid cells, apply design-language glyphs and Lipgloss colors, compose into a styled string in `internal/tui/grid_view.go`
- [ ] T031 [P] [US1] Implement keyboard controls — Enter starts simulation, Space pauses/resumes, q/Esc quits, r restarts — in `internal/tui/controls.go`
- [ ] T032 [P] [US1] Define Lipgloss style constants mapping design-language colors (e.g., `styleFire = lipgloss.NewStyle().Foreground(lipgloss.Color("#ef4444"))`) in `internal/tui/styles.go`
- [ ] T033 [US1] Wire simulation and TUI via the event channel — Simulation accepts optional `chan<- SimulationEvent` write-end (nil means no streaming, e.g. in tests), writes events at end of each Tick; model Update() reads from channel to build render state; cmd/simviz/main.go owns channel lifecycle (create, pass write-end to engine, pass read-end to model, close on shutdown)
- [ ] T034 [US1] Wire end-to-end: verify TUI starts, renders grid, responds to keyboard, simulation state updates on each tick

**Checkpoint**: You can SEE your simulation running in a terminal. Glyphs and colors match the design language. You understand Bubbletea's Elm architecture and Lipgloss styling.

---

## Phase 7: User Story 1 — Entity-Grid Occupancy (P1) 🎯 MVP SLICE 7/8

**Goal**: Citizens mark their occupied cells on the grid so pathfinding treats them as blocked. Obstacles are placed at simulation start to add environmental complexity. Safe zones enforce a capacity limit — citizens targeting a full zone must recalculate. This closes three deferred design gaps (citizen cell occupancy, obstacle provisioning, safe zone capacity).

**Independent Test**: Run a simulation with obstacles and verify pathfinding routes around them. Run a simulation with one safe zone (`MaxOccupants=1`) and two citizens heading toward it — verify only one enters, the other recalculates. Verify two citizens cannot occupy the same cell.

**Learning Outcome**: Grid cell state management, pathfinding integration with dynamic blocking, capacity-aware navigation state machine

- [ ] T035 [P] [US1] Add `CellCitizen` to `CellType` enum in `internal/pathfinding/grid.go`, update pathfinding (`astar.go`, `dijkstra.go`) and `GetRandomOpenPosition` to treat it as blocked alongside `CellObstacle` and `CellHazard`
- [ ] T036 [P] [US1] Implement citizen grid marking/unmarking — mark `CellCitizen` on citizen creation and each move, unmark previous position on departure, wrap in helpers (`markCell`, `unmarkCell`) in `internal/engine/simulation.go`
- [ ] T037 [P] [US1] Define `StaticObstacle` struct (position, size) and `ObstacleConfig` (count range, size range) in `internal/engine/configuration.go`, add obstacles to `SimulationConfig`
- [ ] T038 [US1] Implement obstacle placement — generate random non-overlapping rectangular obstacles during `NewSimulation`, mark cells as `CellObstacle`, validate no overlap with safe zones or other obstacles
- [ ] T039 [P] [US1] Implement `MaxOccupants` and `OccupantIDs` on `SafeZone` struct in `internal/engine/safe_zone.go` — track which citizens are inside, refuse new entries when at capacity
- [ ] T040 [US1] Wire capacity checking into citizen path selection and write tests: after pathfinding, check if the nearest safe zone is at capacity; if full, recalculate toward the next-nearest available zone. Tests verify: citizen cells block pathfinding, obstacles block pathfinding and avoid overlap, safe zone capacity prevents overfilling

**Checkpoint**: `go test -v ./internal/pathfinding/ && go test -v ./internal/engine/` passes. Obstacles appear on the grid at startup. Citizens cannot occupy the same cell. Safe zones enforce capacity limits.

---

## Phase 8: User Story 1 — Config Sidebar (P1) 🎯 MVP SLICE 8/8

**Goal**: A configuration sidebar sits alongside the grid, exposing live-adjustable simulation parameters mapped to `SimulationConfig`. Fields use Bubble Tea `textinput`-style controls with Tab focus cycling and ↑↓ adjustment.

**Independent Test**: Start `simviz`, verify the config sidebar renders alongside the grid. Verify Tab cycles focus through fields and ↑↓ adjusts values. Verify adjusted values take effect on the next tick.

**Learning Outcome**: Horizontal split-pane composition with Lipgloss, focus management, config-as-UI mapping

- [ ] T041 [P] [US1] Implement config sidebar view — render config sections (Simulation, Hazard, Safe Zone) with adjustable `textinput`-style fields for each `SimulationConfig` value, apply focus highlight on active field in `internal/tui/config_view.go`
- [ ] T042 [US1] Add config focus/adjust keyboard controls — Tab cycles focus focusable fields, ↑↓ adjusts the focused field value — extend controls in `internal/tui/controls.go`
- [ ] T043 [US1] Compose split-pane layout — join grid and config sidebar horizontally with `lipgloss.JoinHorizontal(lipgloss.Top, gridView, sidebarView)`, add status bar above and help bar below in `internal/tui/model.go`
- [ ] T044 [US1] Wire and verify: start `simviz`, confirm split-pane layout renders, config sidebar responds to Tab/↑↓, values propagate to simulation engine

**Checkpoint**: The TUI shows a live config sidebar alongside the simulation grid. You understand split-pane composition with Lipgloss and keyboard-driven focus management.

---

## Phase 9: User Story 2 — Event Fan-Out + Optional Remote Observer (P2)

**Goal**: Simulation events fan out from the engine to multiple subscribers via an in-process event hub. The TUI consumes events via the hub. Optionally, an HTTP/WebSocket server allows remote observation from a secondary terminal or browser.

**Independent Test**: Start simulation, verify TUI receives all state changes via the hub. If WS server enabled, connect `websocat ws://localhost:8080/ws` and verify events stream in real-time.

**Learning Outcome**: Go channels for fan-out, hub-and-spoke pattern (in-process), optional `net/http` + `coder/websocket` for remote observers

- [ ] T045 [P] [US2] Implement event hub in `internal/tui/hub.go` — register/unregister subscribers, broadcast `SimulationEvent` to all subscribers via buffered channels
- [ ] T046 [US2] Wire hub into `cmd/simviz/main.go` — engine publishes events to hub, TUI subscribes as a hub client
- [ ] T047 [US2] Integration test: create hub, register mock subscribers, emit events, verify all subscribers receive them in order
- [ ] T048 [P] [US2, optional stretch] Implement optional WebSocket remote observer — add `cmd/simobs/main.go` with HTTP server, WebSocket endpoint (using `coder/websocket`), subscribe to hub, stream JSON events to connected clients
- [ ] T049 [US2, optional stretch] Integration test: run `simviz` with WS server enabled, connect mock WebSocket client, emit events, verify client receives them

**Checkpoint**: Events fan out to the TUI in real-time via the hub pattern. Optional: remote clients observe the simulation via WebSocket. You understand Go channel-based fan-out and hub-spoke architecture.

---

## Phase 10: User Story 1 — CLI Controls (P2)

**Goal**: An operator can start, pause, resume, stop, and check status of a simulation via a CLI tool.

**Independent Test**: Run `go build ./cmd/simctl && ./simctl start --config examples/simple-sim.json` and verify simulation runs, then `./simctl status` shows progress.

**Learning Outcome**: Go `flag` package, JSON config loading, signal handling, CLI patterns

- [ ] T050 [US1] Create `cmd/simctl/main.go` with command structure and `flag`-based subcommand parsing for `start`, `pause`, `resume`, `stop`, `status`
- [ ] T051 [US1] Implement config loading from JSON file path in `internal/config/config.go` — read file, unmarshal into `SimulationConfig`, validate required fields
- [ ] T052 [US1] Implement simulation lifecycle commands: `start` kicks off tick loop in goroutine, `pause`/`resume` toggle state, `stop` terminates, `status` prints summary
- [ ] T053 [US1] Add `examples/simple-sim.json` with a minimal working config (1 citizen, 1 safe zone, 1 hazard, small grid) and verify end-to-end: build, run, status, stop

**Checkpoint**: `go build ./cmd/simctl` succeeds. `./simctl start --config examples/simple-sim.json` runs a simulation. You understand CLI construction and JSON config.

---

## Phase 11: User Story 3 — NATS/JetStream Integration + Event History (P3) 📡

**Goal**: Simulation events are published to NATS JetStream and can be replayed after the simulation ends.

**Independent Test**: Start NATS via docker-compose, run a simulation, consume events from the `simulation-events` stream, verify complete ordered sequence.

**Learning Outcome**: `nats.go` JetStream client, event streaming, consumer management, ordered delivery, `context.Context`

- [ ] T054 [P] [US3] Implement JetStream producer in `internal/messaging/producer.go` — connect to NATS, create JetStream context, publish JSON-serialized `SimulationEvent` to `simulation-events.<event_type>` subject (e.g., `simulation-events.citizen.moved`)
- [ ] T055 [P] [US3] Implement JetStream consumer in `internal/messaging/consumer.go` — subscribe to stream, consume events in order with configurable durable consumer name
- [ ] T056 [US3] Integrate JetStream producer into simulation tick loop — publish events to JetStream instead of (or in addition to) in-memory storage in `internal/engine/simulation.go`
- [ ] T057 [US3] Implement event replay — add `replay` command to `cmd/simctl/main.go` that creates a JetStream consumer and prints or saves events in order
- [ ] T058 [US3] Write tests in `internal/messaging/producer_test.go` and `consumer_test.go` — use an embedded NATS server (`github.com/nats-io/nats-server/v2/server`) for integration tests or verify serialization/deserialization round-trip

**Checkpoint**: Events flow through NATS JetStream. You can replay a simulation from the event stream. You understand JetStream producers, consumers, and event sourcing.

---

## Phase 12: User Story 4 — Autonomy + Performance (P3) ⚡

**Goal**: Citizens have varied behaviors (risk tolerance, path preference) and the system handles 100+ citizens at <2x real-time.

**Independent Test**: Run a simulation with 100 citizens, 5 hazards, obstacles — verify it completes within 2x real-time and citizens show varied path choices.

**Learning Outcome**: Weighted A*, benchmarking (`go test -bench`), `pprof`, algorithm optimization

- [ ] T059 [P] [US4] Implement weighted A* variant in `internal/pathfinding/astar.go` — accept a weight function to penalize cells near hazards for "safest" path preference
- [ ] T060 [US4] Implement autonomy profile integration — use citizen's `RiskTolerance`, `SpeedVariation`, and `PathPreference` to influence movement and path selection in `internal/engine/citizen.go`
- [ ] T061 [US4] Add benchmarks in `internal/pathfinding/astar_bench_test.go` and `internal/engine/simulation_bench_test.go` — measure pathfinding time for 100 citizens, measure simulation tick time
- [ ] T062 [US4] Optimize: profile with `pprof`, identify bottlenecks, optimize hot paths (grid cell allocation, path copying, hazard cell marking) to meet <2x real-time target

**Checkpoint**: `go test -bench=. -benchtime=10x ./...` shows acceptable performance. 100 citizens navigate with varied behaviors.

---

## Phase 13: Polish & Cross-Cutting Concerns ✨

**Purpose**: Final quality pass, documentation, and end-to-end validation

- [ ] T063 Run full validation: `gofmt -s -w . && go vet ./... && staticcheck ./... && go test -race ./...` — fix all issues
- [ ] T064 Validate all quickstart.md steps work end-to-end: NATS up → simctl start → events in JetStream → simviz → TUI shows simulation

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
Phase 6: US1 Core TUI — depends on Phase 5 (needs events to display)
    ↓
Phase 7: US1 Entity-Grid Occupancy — depends on Phase 4 (needs citizens, safe zones, hazards on grid)
    ↓
Phase 8: US1 Config Sidebar — depends on Phase 6 (extends TUI layout)
    ↓
Phase 9: US2 Event Fan-Out — depends on Phase 6 (decouples TUI from engine)
    ↓
Phase 10: US1 CLI Controls — depends on Phase 5 (needs events for output)
    ↓
Phase 11: US3 NATS/JetStream — depends on Phase 5 (needs events)
    ↓
Phase 12: US4 Autonomy — depends on Phase 4 (needs full simulation)
    ↓
Phase 13: Polish — depends on all phases
```

### User Story Dependencies

- **US1 (P1)**: Core simulation — no dependencies on other stories. Phases 1–8 (Grid, Movement, Hazards, Safe Zones, Events, Core TUI, Entity-Grid Occupancy, Config Sidebar) complete the P1 MVP. CLI Controls (Phase 10) is P2.
- **US2 (P2)**: Event fan-out + optional remote observer — depends on US1 event emission (Phase 5) and Phase 6 TUI. Phase 9 decouples the TUI from the engine via an event hub and optionally adds a WebSocket remote observer.
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
- Phase 6 (US1/TUI Core): T028, T029, T030, T031, T032 can all run in parallel (different tui/ files + events/ channel type)
- Phase 7 (US1/Entity-Grid Occupancy): T035, T036, T037, T039 can run in parallel (different packages/files); T038 depends on T037; T040 depends on all
- Phase 8 (US1/Config Sidebar): T041 is independent; T041, T042 can run in parallel
- Phase 9 (US2/Event Fan-Out): T045 is independent; T048 is stretch
- Phase 10 (US1/CLI): All tasks are sequential
- Phase 11 (US3/NATS/JetStream): T054, T055 can run in parallel
- Phase 12 (US4/Autonomy): T059 is independent

---

## Parallel Example: Phase 6 (Terminal TUI Core)

```bash
# Launch all parallel tasks for Phase 6 together:
# Task T028 + T029 + T030 + T031 + T032: channel type + tui/ components (events + model + grid + controls + styles)

# Then assemble:
# Task T033: simviz main.go — wire engine and TUI via channel (depends on T028, T029, T030, T031, T032)
# Task T034: end-to-end verification
```

## Parallel Example: Phase 8 (Config Sidebar)

```bash
# Launch all parallel tasks for Phase 8 together:
# Task T041: config sidebar view
# Task T042: focus/adjust keyboard controls

# Then assemble:
# Task T043: compose split-pane layout (depends on T041, T042, plus Phase 6 T029, T030)
# Task T044: wire and verify
```

---

## Implementation Strategy

### MVP Scope (User Story 1, P1 Only)

The core MVP covers Phases 1–8 (T008–T044). This delivers:
- A working simulation with citizens, hazards, safe zones
- A* pathfinding with obstacle and citizen avoidance
- Event emission in memory
- Terminal TUI with grid visualization and config sidebar
- Safe zone capacity enforcement with citizen recalculation
- All testable via `go test ./...`

**Do NOT skip ahead to US2/US3/US4 until Phases 1–8 are complete.**

### Learning Milestones

| Phase | Go Concepts Mastered | Milestone |
|-------|---------------------|-----------|
| 1 | structs, slices, interfaces, table-driven tests | `go test -v ./internal/pathfinding/` passes |
| 2 | methods, state mutation, tick loops | Citizens move step-by-step each tick |
| 3 | config-driven behavior, lifecycle management | Hazards emerge and expand on schedule |
| 4 | termination conditions, scheduled emergence, edge cases | Simulation auto-completes; safe zones appear mid-run |
| 5 | event emission, UUID, time | Complete event log for any run |
| 6 | Bubbletea Elm architecture, Lipgloss styling, terminal rendering | TUI shows live simulation with keyboard controls |
| 7 | grid cell state, pathfinding with dynamic blocking, capacity-aware navigation | Citizens block cells; safe zones enforce capacity |
| 8 | split-pane composition, focus management, config-as-UI mapping | TUI shows live config sidebar alongside grid |
| 9 | Go channels for fan-out, hub-spoke pattern | Events reach TUI via hub; optional WS remote observer |
| 10 | `flag` package, JSON unmarshal, signal handling | `./simctl start --config x.json` works |
| 11 | JetStream producer/consumer, context | Events streaming through NATS JetStream |
| 12 | weighted algorithms, benchmarks, pprof | 100 citizens at <2x real-time |

### Incremental Delivery

1. Complete Phase 1 → Grid + A* pathfinding ready
2. Complete Phases 1–4 → Full simulation with escape/death/completion
3. Complete Phases 1–6 → Simulation with live terminal TUI (**huge win!**)
4. Complete Phases 1–7 → Entity-grid occupancy + safe zone capacity
5. Complete Phases 1–8 → Full P1 MVP with config sidebar
6. Add Phase 9 → In-process event hub + optional remote observer
7. Add Phase 10 → CLI controls (headless operation)
8. Add Phase 11 → NATS JetStream event streaming (core learning goal)
9. Add Phase 12 → Scale and autonomy (polish)

### Validation

After each phase, run:
```bash
gofmt -s -w .
go vet ./...
staticcheck ./...
go test -race ./...
```

Before committing, verify nothing is broken.
