# Research: Hazard Simulation System

## 1. Go NATS/JetStream Client Library

**Decision**: `github.com/nats-io/nats.go`

**Rationale**:
- Official Go client maintained by Synadia — same team as the NATS server itself
- Pure Go — no CGo, `CGO_ENABLED=0` safe, trivial cross-compilation, fast builds
- JetStream is built into NATS Server — no separate broker, no additional infrastructure
- Single `nats.Conn` handles core pub/sub, JetStream streaming, and key-value store
- Actively developed — v1.40+ with full JetStream API, consumer management, and pull/push subscriptions
- Lighter weight than Kafka: NATS server is a single ~20MB binary, starts in <100ms
- Subjects use hierarchical dot notation (`hazard.sim.events.>`) — simpler than Kafka topic config
- No partition/replication factor management for dev — JetStream streams are ready with a single `nats.Conn.AddStream()`
- Composable: JetStream consumers can be ordered, durable, or ephemeral — matches Kafka consumer group semantics
- Context-aware API — `JetStreamContext.PublishAsync()`, `SubscribeSync()`, `ChanSubscribe()`

**Alternatives considered**:
- **franz-go / sarama (Kafka)**: Kafka requires a separate broker (JVM-based), heavier dev setup, more configuration overhead. Good for production-scale systems but overkill for a learning project. NATS gives the same event sourcing patterns with a fraction of the operational complexity.
- **RabbitMQ + STOMP**: Strong broker but lacks built-in streaming persistence (needs shovel/federation plugins). JetStream provides Kafka-like retention, replay, and consumer groups out of the box.
- **Redis Streams**: Simple but no consumer group semantics as robust as JetStream; persistence model differs from typical event sourcing.

**Local development**:
```yaml
# docker-compose.yml — single-node NATS with JetStream for development
services:
  nats:
    image: nats:2.10-alpine
    ports:
      - "4222:4222"   # client connections
      - "8222:8222"   # HTTP monitoring
    command: ["-js", "-sd", "/data"]
    volumes:
      - nats_data:/data

volumes:
  nats_data:
```

Optionally, run the NATS server directly without Docker:
```bash
# Download single binary from https://nats.io/download/
nats-server -js
```

---

## 2. Go WebSocket Library (optional / remote observation)

**Decision**: `github.com/coder/websocket` (formerly `nhooyr.io/websocket`) — deferred to stretch goal

**Rationale**:
- The primary visualization is the TUI (Bubbletea), which communicates with the engine via Go channels
- WebSocket is only needed if remote observation (browser or second terminal) is desired
- When needed, `coder/websocket` remains the recommended choice for the same reasons as originally documented
- See Phase 8 for the optional remote observer implementation

**Alternatives considered**:
- **gorilla/websocket**: Dormant maintenance cycle; no `context.Context` support; concurrent writes panic (requires manual synchronization). Avoid for new projects.
- **golang.org/x/net/websocket**: Officially deprecated by Go team. Violates RFC 6455 on message framing.

**Broadcast pattern**: Hub-and-spoke. The hub runs a single goroutine for client registration/unregistration. In the TUI path, the hub uses Go channels instead of WebSocket connections.

---

## 3. Event Schema Format

**Decision**: JSON (serialized as UTF-8 bytes)

**Rationale**:
- Zero dependencies — no schema registry or code generation needed for v1
- Human-readable for debugging during development
- `encoding/json` is in the Go standard library
- `coder/websocket`'s `wsjson` subpackage handles marshaling natively
- Can evolve to Avro/Protobuf via NATS message headers or schema registry if needed later

---

## 4. Simulation Tick Rate

**Decision**: Configurable, default ~10 Hz (100ms per tick)

**Rationale**:
- Balances visual smoothness with per-tick computational cost
- Citizens move a configurable number of grid cells per tick
- Each tick: simulate citizen movement, hazard expansion, event emission
- JetStream events are batched per-tick, not per-movement
- Visualization interpolates between tick states for smooth rendering

---

## 5. Go TUI Framework

**Decision**: `github.com/charmbracelet/bubbletea` + `github.com/charmbracelet/lipgloss` + `github.com/charmbracelet/bubbles`

**Rationale**:
- Bubbletea is the most widely adopted Go TUI framework with an active ecosystem and thorough documentation (examples, FAQ, Discord)
- Elm-style Model/Update/View architecture maps naturally to the simulation's tick-driven state model — the `Model` holds the render state, `Update` processes ticks and keypresses, `View` produces the styled string output
- Lipgloss provides composable style definitions with 256-color and truecolor support — directly maps the design language hex colors (e.g., `lipgloss.Color("#ef4444")`)
- Bubbles component library provides pre-built input fields, select lists, spinners, etc. for future interactive controls (config editing, hazard tuning)
- Pure Go — no CGo, no external runtime, trivial cross-compilation
- `tea.Program` with `tea.WithAltScreen()` creates a clean full-screen terminal experience
- Keyboard input handling via `tea.KeyMsg` is straightforward for start/pause/quit controls
- The `tea.Cmd` pattern for asynchronous operations maps naturally to simulation ticks and NATS subscriptions
- `tea.Quit` provides a clean shutdown path

**Alternatives considered**:
- **tview**: More widget-focused, less composable than Bubbletea; uses its own draw model rather than Elm architecture
- **Termui**: Graph/chart focused, less suited for grid-based simulation rendering
- **ANSI escape codes directly**: Too low-level; defeats the learning value of TUI framework patterns

**Workflow**:
```bash
# Run the TUI directly (no build step needed for the viz layer)
go run ./cmd/simviz

# Or build and run
go build -o bin/simviz ./cmd/simviz
./bin/simviz
```

---

## 6. JetStream Stream Configuration

**Decision**: Single stream `simulation-events` with one subject for v1

**Rationale**:
- Simplified producer/consumer setup for learning
- JetStream provides ordered delivery, persistence, and replay — core event sourcing primitives
- Event envelope includes type discriminator (client-side routing)
- Can split into domain subjects (`citizen.events.>`, `hazard.events.>`, `sim.events.>`) within the same stream in v2 if needed

**Stream config**:
- Name: `simulation-events`
- Subjects: `simulation-events.>` (hierarchical subjects allow sub-topics if needed later)
- Storage: file (persisted to disk)
- Retention: `limits` (age-based, configurable default 1h)
- MaxAge: 1h (configurable)
- Storage: `FileStorage` (survives restarts)

**NATS subject hierarchy** (future expansion):
```
simulation-events.start          → simulation.started
simulation-events.citizen.move   → citizen.moved
simulation-events.citizen.escape → citizen.escaped
simulation-events.hazard.emer    → hazard.emerged
simulation-events.hazard.expand  → hazard.expanded
```
