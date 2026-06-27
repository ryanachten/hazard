# WebSocket Remote Observer (Optional / Stretch Goal)

> **Note**: The primary visualization for this project is the terminal TUI (Bubbletea), which receives events via Go channels. This document defines an **optional** WebSocket protocol for remote observation — a stretch goal that can be implemented after the core TUI is working.

## Connection

1. Client opens WebSocket connection to `ws://localhost:8080/ws`
2. Server accepts upgrade, registers client for broadcast
3. Client is now subscribed to all simulation events

## Server-to-Client Messages

Messages use the same event envelope as the in-memory event system, streamed in real-time as they occur.

```json
{
  "event_type": "citizen.moved",
  "entity_id": "cit_01J...",
  "timestamp": "2026-06-13T12:00:00.000Z",
  "tick": 42,
  "payload": { }
}
```

## Client-to-Server Messages

For v1, clients are receive-only. No client-to-server messages are expected.
If a client sends data, the server may ignore or log it.

## Client Lifecycle

| Event | Server Action |
|---|---|
| Client connects | Register client, begin broadcasting current simulation state |
| Client disconnects | Unregister client, stop sending events |
| Network error | Close connection, unregister client |

## Error Handling

- Server closes connection with status `1011` (Internal Error) on unrecoverable errors
- Server closes connection with status `1000` (Normal Closure) on simulation end
- Clients should implement reconnection with exponential backoff
- Disconnected mid-simulation: Client misses events; on reconnect, client should request current state snapshot (v2 feature)

## Relationship to In-Process Event Hub

The optional WebSocket server subscribes to the same in-process event hub that feeds the TUI. The hub's channel-based fan-out supports both local (TUI) and remote (WebSocket) subscribers transparently. See `plan.md` Phase 8 for the hub implementation.
