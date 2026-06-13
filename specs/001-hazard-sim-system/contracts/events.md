# Event Contracts

## Kafka Topic: `simulation-events`

All simulation state changes are published as JSON events to a single Kafka topic.

## Event Envelope

Every event follows this envelope structure:

```json
{
  "id": "evt_01J...",
  "simulation_id": "sim_01J...",
  "timestamp": "2026-06-13T12:00:00.000Z",
  "tick": 42,
  "event_type": "citizen.moved",
  "entity_id": "cit_01J...",
  "payload": { }
}
```

## Event Types

### simulation.started

Emitted after the generation phase completes. Contains the concrete values generated for this run.

```json
{
  "event_type": "simulation.started",
  "entity_id": "<simulation_id>",
  "payload": {
    "seed": 42,
    "width": 100,
    "height": 100,
    "tick_interval_ms": 100,
    "citizen_count": 10,
    "obstacle_count": 2,
    "safe_zone_count": 1
  }
}
```

### simulation.paused / simulation.stopped / simulation.completed

```json
{
  "event_type": "simulation.paused",
  "entity_id": "<simulation_id>",
  "payload": {
    "tick": 42,
    "escaped_count": 3,
    "dead_count": 1,
    "navigating_count": 6
  }
}
```

### citizen.moved

```json
{
  "event_type": "citizen.moved",
  "entity_id": "<citizen_id>",
  "payload": {
    "from": { "x": 5, "y": 10 },
    "to": { "x": 5, "y": 11 },
    "status": "navigating"
  }
}
```

### citizen.escaped

```json
{
  "event_type": "citizen.escaped",
  "entity_id": "<citizen_id>",
  "payload": {
    "final_position": { "x": 42, "y": 38 },
    "safe_zone_id": "<safe_zone_id>",
    "ticks_survived": 87
  }
}
```

### citizen.died

```json
{
  "event_type": "citizen.died",
  "entity_id": "<citizen_id>",
  "payload": {
    "position": { "x": 20, "y": 15 },
    "killed_by_hazard": "<hazard_id>",
    "ticks_survived": 65
  }
}
```

### citizen.recalculated

```json
{
  "event_type": "citizen.recalculated",
  "entity_id": "<citizen_id>",
  "payload": {
    "reason": "hazard_blocked_path",
    "previous_target_zone": "<safe_zone_id>",
    "new_target_zone": "<safe_zone_id>",
    "path_length": 28
  }
}
```

### hazard.emerged

```json
{
  "event_type": "hazard.emerged",
  "entity_id": "<hazard_id>",
  "payload": {
    "position": { "x": 30, "y": 25 },
    "initial_radius": 3.0,
    "spread_rate": 0.5,
    "max_radius": 15.0,
    "hazard_type": {
      "name": "generic",
      "color": "#ff4444",
      "spread_curve": "linear"
    }
  }
}
```

### hazard.expanded

```json
{
  "event_type": "hazard.expanded",
  "entity_id": "<hazard_id>",
  "payload": {
    "previous_radius": 3.0,
    "current_radius": 3.5,
    "cells_affected": [
      {"x": 30, "y": 25},
      {"x": 31, "y": 25},
      {"x": 30, "y": 26}
    ]
  }
}
```

### hazard.dissipated

```json
{
  "event_type": "hazard.dissipated",
  "entity_id": "<hazard_id>",
  "payload": {
    "final_radius": 12.0,
    "ticks_active": 200
  }
}
```
