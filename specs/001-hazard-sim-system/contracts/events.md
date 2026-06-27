# Event Contracts

## JetStream Stream: `simulation-events`

All simulation state changes are published as JSON events to a single NATS JetStream stream.
When using subject hierarchy, events are published to `simulation-events.<event_type>` (e.g., `simulation-events.citizen.moved`).

## Event Envelope

Every event follows this envelope structure:

```json
{
  "id": "evt_01J...",
  "simulationId": "sim_01J...",
  "timestamp": "2026-06-13T12:00:00.000Z",
  "tick": 42,
  "eventType": "citizen.moved",
  "entityId": "cit_01J...",
  "payload": { }
}
```

## Event Types

### simulation.started

Emitted after the generation phase completes. Contains the concrete values generated for this run.

```json
{
  "eventType": "simulation.started",
  "entityId": "<simulation_id>",
  "payload": {
    "seed": 42,
    "width": 100,
    "height": 100,
    "tickIntervalMs": 100,
    "citizenCount": 10,
    "obstacleCount": 2,
    "safeZoneCount": 1
  }
}
```

### simulation.paused / simulation.stopped / simulation.completed

```json
{
  "eventType": "simulation.paused",
  "entityId": "<simulation_id>",
  "payload": {
    "tick": 42,
    "escapedCount": 3,
    "deadCount": 1,
    "navigatingCount": 6
  }
}
```

### citizen.moved

```json
{
  "eventType": "citizen.moved",
  "entityId": "<citizen_id>",
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
  "eventType": "citizen.escaped",
  "entityId": "<citizen_id>",
  "payload": {
    "finalPosition": { "x": 42, "y": 38 },
    "safeZoneId": "<safe_zone_id>",
    "ticksSurvived": 87
  }
}
```

### citizen.died

```json
{
  "eventType": "citizen.died",
  "entityId": "<citizen_id>",
  "payload": {
    "position": { "x": 20, "y": 15 },
    "killedByHazard": "<hazard_id>",
    "ticksSurvived": 65
  }
}
```

### citizen.recalculated

```json
{
  "eventType": "citizen.recalculated",
  "entityId": "<citizen_id>",
  "payload": {
    "reason": "hazardBlockedPath",
    "previousTargetZone": "<safe_zone_id>",
    "newTargetZone": "<safe_zone_id>",
    "pathLength": 28
  }
}
```

### hazard.emerged

```json
{
  "eventType": "hazard.emerged",
  "entityId": "<hazard_id>",
  "payload": {
    "position": { "x": 30, "y": 25 },
    "initialRadius": 3.0,
    "spreadRate": 0.5,
    "maxRadius": 15.0,
    "hazardType": {
      "name": "fire",
      "kind": "expanding"
    }
  }
}
```

### hazard.expanded

```json
{
  "eventType": "hazard.expanded",
  "entityId": "<hazard_id>",
  "payload": {
    "previousRadius": 3.0,
    "currentRadius": 3.5,
    "cellsAffected": [
      {"x": 30, "y": 25},
      {"x": 31, "y": 25},
      {"x": 30, "y": 26}
    ]
  }
}
```

### hazard.dissipated (terminal — emitted before removal)

```json
{
  "eventType": "hazard.dissipated",
  "entityId": "<hazard_id>",
  "payload": {
    "finalRadius": 12.0,
    "ticksActive": 200
  }
}
```
