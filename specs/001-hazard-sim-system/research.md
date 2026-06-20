# Research: Hazard Simulation System

## 1. Go Kafka Client Library

**Decision**: franz-go (`github.com/twmb/franz-go`)

**Rationale**:
- Pure Go — no CGo, `CGO_ENABLED=0` safe, trivial cross-compilation, fast builds
- Modern, idiomatic API with functional options and `context.Context` support
- Single `kgo.Client` can both produce and consume — simplifies `internal/messaging/`
- Consistently fastest Go Kafka client in benchmarks
- KIP-comprehensive (all client features from Kafka 0.8 to 4.2+)
- Small dependency footprint — single `go get github.com/twmb/franz-go`
- Built-in schema registry client (`pkg/sr`) for future schema evolution

**Alternatives considered**:
- **sarama** (`github.com/IBM/sarama`): Larger community, more tutorials, but heavier API with separate producer/consumer clients and struct-based config. More historical edge-case bugs. Good backup choice.
- **confluent-kafka-go** (`github.com/confluentinc/confluent-kafka-go`): Requires CGo (librdkafka). Adds build complexity, prevents `CGO_ENABLED=0`, complicates cross-compilation. Avoid for this project.

**Local development**:
```yaml
# docker-compose.yml — single-node Kafka for development
services:
  kafka:
    image: apache/kafka:4.1.0
    ports:
      - "9092:9092"
    environment:
      KAFKA_NODE_ID: 1
      KAFKA_PROCESS_ROLES: broker,controller
      KAFKA_CONTROLLER_QUORUM_VOTERS: 1@localhost:9093
      KAFKA_LISTENERS: PLAINTEXT://0.0.0.0:9092,CONTROLLER://0.0.0.0:9093
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092
      KAFKA_CONTROLLER_LISTENER_NAMES: CONTROLLER
      KAFKA_LISTENER_SECURITY_PROTOCOL_MAP: PLAINTEXT:PLAINTEXT,CONTROLLER:PLAINTEXT
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
      KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR: 1
      KAFKA_TRANSACTION_STATE_LOG_MIN_ISR: 1
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
- Can evolve to Avro/Protobuf with franz-go's schema registry client if needed later

---

## 4. Simulation Tick Rate

**Decision**: Configurable, default ~10 Hz (100ms per tick)

**Rationale**:
- Balances visual smoothness with per-tick computational cost
- Citizens move a configurable number of grid cells per tick
- Each tick: simulate citizen movement, hazard expansion, event emission
- Kafka events are batched per-tick, not per-movement
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

## 6. Kafka Topics Design

**Decision**: Single topic `simulation-events` for v1

**Rationale**:
- Simplified producer/consumer setup for learning
- Partition by simulation run ID for ordering guarantees within a run
- Event envelope includes type discriminator (client-side routing)
- Can split into domain topics (citizen-events, hazard-events, sim-lifecycle) in v2 if needed

**Topic config**:
- Name: `simulation-events`
- Partitions: 1 (single machine, single simulation)
- Replication factor: 1 (single broker)
- Cleanup policy: `delete` with retention (configurable, default 1 hour)
