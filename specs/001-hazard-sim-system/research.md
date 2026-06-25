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

## 2. Go WebSocket Library

**Decision**: `github.com/coder/websocket` (formerly `nhooyr.io/websocket`)

**Rationale**:
- Actively maintained by Coder (v1.8.14, Sep 2025, with ongoing commits into 2026)
- Minimal, idiomatic Go API with `context.Context` throughout
- Concurrent write safe — no panics, no extra synchronization needed
- `wsjson` subpackage provides zero-alloc JSON reads/writes
- Composes naturally with standard `net/http` — `websocket.Accept(w, r, nil)` is a single call
- Used by Traefik, Vault, Cloudflare in production
- Full permessage-deflate compression, WASM compilation support, `net.Conn` wrapper

**Alternatives considered**:
- **gorilla/websocket**: Dormant maintenance cycle; no `context.Context` support; concurrent writes panic (requires manual synchronization). Avoid for new projects.
- **golang.org/x/net/websocket**: Officially deprecated by Go team. Violates RFC 6455 on message framing.

**Broadcast pattern**: Hub-and-spoke. The hub runs a single goroutine for client registration/unregistration. `coder/websocket`'s concurrent write safety enables `range clients { wsjson.Write(ctx, conn, event) }` without per-client write goroutines or mutexes.

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

## 5. Frontend Build Tool

**Decision**: TypeScript + Bun (`bun build`)

**Rationale**:
- TypeScript gives type safety for the Canvas 2D rendering code and WebSocket/HTTP client logic
- Bun's built-in bundler (`bun build`) compiles TypeScript to a single static JS file with zero config
- No package.json or node_modules required for this project — the frontend has no runtime dependencies (Canvas 2D is a browser API; WebSocket is built-in)
- Build command is trivial: `bun build ./web/src/canvas.ts --outdir ./web`
- Output JS is served as a static asset by the Go `simviz` binary
- Fast builds (<50ms) keep the edit-rebuild-refresh cycle tight

**Alternatives considered**:
- **Plain JS**: Simpler but loses type safety, autocomplete, and inline documentation that TS provides for Canvas 2D APIs
- **Deno**: Viable alternative with `deno bundle`/`deno pack`, but Bun's bundler is more established for this use case and slightly simpler to invoke
- **esbuild/webpack/vite**: Overkill for a single-file frontend with no framework, no npm dependencies, no CSS preprocessing
- **tsc only**: Requires separate bundler; Bun replaces both compiler and bundler in one tool

**Workflow**:
```bash
# Build frontend assets
bun build ./web/src/canvas.ts --outdir ./web

# Run the Go server (which serves web/ as static files)
go run ./cmd/simviz
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
