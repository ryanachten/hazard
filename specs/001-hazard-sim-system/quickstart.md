# Quickstart: Hazard Simulation System

## Prerequisites

- Go 1.26+
- Docker (for NATS, or run `nats-server` directly)
- `gofmt`, `go vet`, `staticcheck` (install: `go install honnef.co/go/tools/cmd/staticcheck@latest`)

## Setup

```bash
# Clone and enter the project
git checkout 001-hazard-sim-system

# Start NATS with JetStream
docker compose up -d

# Verify NATS is running
docker compose logs nats | head -5

# Run tests
go test ./...
```

## Run a Simulation

```bash
# Build the CLI
go build -o bin/simctl ./cmd/simctl

# Start a simulation with default config
./bin/simctl start --config examples/simple-sim.json

# View status
./bin/simctl status

# Stop the simulation
./bin/simctl stop
```

## View the Visualization

```bash
# Start the visualization server
go build -o bin/simviz ./cmd/simviz
./bin/simviz

# Open in browser
open http://localhost:8080
```

## Run All Checks

```bash
gofmt -s -w .
go vet ./...
staticcheck ./...
go test -race ./...
```

## Project Layout

```
├── cmd/
│   ├── simctl/           # CLI operator (start, pause, stop)
│   └── simviz/           # WebSocket visualization server
├── internal/
│   ├── engine/           # Core simulation loop
│   ├── pathfinding/      # A* pathfinder
│   ├── messaging/        # NATS JetStream producer/consumer
│   ├── events/           # Event types
│   ├── vis/              # WebSocket hub
│   └── config/           # Config loading
├── web/                  # HTML/Canvas frontend
├── docker-compose.yml    # NATS for development
└── specs/001-hazard-sim-system/
    ├── plan.md
    ├── research.md
    ├── data-model.md
    └── contracts/
```

## Implementation Slices

This project is built in small, progressive slices. Each slice is independently testable.

| # | Slice | What You'll Build | Learning Goal |
|---|---|---|---|
| 1 | Grid + A* Pathfinding | 2D grid, A* algorithm, path visualization test | Go structs, slices, interfaces, unit tests |
| 2 | Citizen Movement | Citizens move along paths on the grid | Go methods, tick loop, state management |
| 3 | Hazards + Envelopment | Hazard emergence, radius expansion, grid blocking | Concurrent state, tick-based simulation |
| 4 | Safe Zones + Death | Citizens escape or die, simulation completion | State machines, edge cases |
| 5 | Event Emission | Event type constructors, tick integration, in-memory storage | time.Time, UUID, event patterns |
| 6 | Browser Viz (HTTP poll) | Simulation state exposed as JSON, Canvas rendering via TypeScript | net/http, bun build, Canvas 2D API, requestAnimationFrame |
| 7 | CLI Controls | simctl start, pause, stop, status | flag package, JSON config, signal handling |
| 8 | WebSocket Upgrade | Replace HTTP polling with WebSocket push in TypeScript frontend | coder/websocket, hub-and-spoke pattern, bun build |
| 9 | NATS/JetStream Integration | Produce/consume simulation events | nats.go, JetStream, docker-compose, event serialization |
| 10 | Autonomy + Performance | Risk tolerance, path preference, 100+ citizens | A* variants, benchmarking, pprof |
