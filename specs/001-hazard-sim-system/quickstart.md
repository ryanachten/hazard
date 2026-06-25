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
| 5 | NATS/JetStream Integration | Produce/consume events via NATS JetStream | nats.go, event streaming, docker-compose |
| 6 | WebSocket Broadcast | Stream events to browser clients | coder/websocket, hub pattern |
| 7 | HTML Canvas Viz | Real-time visualization in browser | Canvas API, JSON event rendering |
| 8 | CLI Controls | Start, pause, stop simulation via CLI | CLI construction, signal handling |
| 9 | Event History | Full event replay from JetStream | Consumer groups, event sourcing |
| 10 | Autonomy + Scaling | Multi-path preference, 100+ citizens | A* variants, performance optimization |
