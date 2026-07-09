// Package events defines simulation events
package events

import (
	c "hazard/internal/common"
	pf "hazard/internal/pathfinding"
	"time"

	"github.com/google/uuid"
)

// SimulationEvent represents a domain event that occurred during a simulation
type SimulationEvent struct {
	ID           uuid.UUID
	EventType    eventType
	SimulationID uuid.UUID
	Tick         uint64
	Timestamp    time.Time
	EntityID     uuid.UUID
	Payload      any
}

// EventMetadata common for all simulation events
type EventMetadata struct {
	SimulationID uuid.UUID
	Tick         uint64
}

type eventType string

const (
	// SimulationCreated event type indicates a new simulation has been created
	SimulationCreated eventType = "simulation.created"
	// SimulationCompleted event type indicates the simulation has completed
	SimulationCompleted eventType = "simulation.completed"
	// CitizenMoved event type for citizen movement
	CitizenMoved eventType = "citizen.moved"
	// CitizenPathUpdated event type for citizen path recalculation
	CitizenPathUpdated eventType = "citizen.pathUpdated"
	// CitizenEscaped event type for citizen reaching safety
	CitizenEscaped eventType = "citizen.escaped"
	// CitizenDied event type for citizen overtaken by hazard
	CitizenDied eventType = "citizen.died"
	// SafeZoneEmerged event type for new safe zone appearing on the grid
	SafeZoneEmerged eventType = "safeZone.emerged"
	// HazardEmerged event type for new hazard appearing on the grid
	HazardEmerged eventType = "hazard.emerged"
	// HazardExpanded event type for hazard radius increase
	HazardExpanded eventType = "hazard.expanded"
	// HazardDissipated event type for hazard dissipation
	HazardDissipated eventType = "hazard.dissipated"
)

// SimulationCreatedPayload for simulation start event
type SimulationCreatedPayload struct {
	Grid      pf.Grid
	Citizens  []c.Citizen
	SafeZones []c.SafeZone
	Obstacles []c.Obstacle
}

// SimulationCreated raised when simulation starts
func (e *EventBus) SimulationCreated(payload SimulationCreatedPayload, metadata EventMetadata) {
	citizenSnapshot := make([]c.Citizen, len(payload.Citizens))
	for i, citizen := range payload.Citizens {
		citizenSnapshot[i] = citizen.Copy()
	}

	safeZoneSnapshot := make([]c.SafeZone, len(payload.SafeZones))
	for i, safeZone := range payload.SafeZones {
		safeZoneSnapshot[i] = safeZone.Copy()
	}

	obstacleSnapshot := make([]c.Obstacle, len(payload.Obstacles))
	for i, obstacle := range payload.Obstacles {
		obstacleSnapshot[i] = obstacle.Copy()
	}

	e.createEvent(SimulationCreated, metadata.SimulationID, metadata, SimulationCreatedPayload{
		Grid:      payload.Grid,
		Citizens:  citizenSnapshot,
		SafeZones: safeZoneSnapshot,
		Obstacles: obstacleSnapshot,
	})
}

// SimulationCompleted raised when simulation finishes
func (e *EventBus) SimulationCompleted(metadata EventMetadata) {
	e.createEvent(SimulationCompleted, metadata.SimulationID, metadata, nil)
}

// CitizenMoved raised when a citizen moves
func (e *EventBus) CitizenMoved(citizenID uuid.UUID, newPosition pf.Position, metadata EventMetadata) {
	e.createEvent(CitizenMoved, citizenID, metadata, newPosition)
}

// CitizenPathUpdated raised when a citizen's path is recalculated
func (e *EventBus) CitizenPathUpdated(citizenID uuid.UUID, path []pf.Position, metadata EventMetadata) {
	e.createEvent(CitizenPathUpdated, citizenID, metadata, getPositionSnapshot(path))
}

// CitizenEscapedPayload for citizen escape event
type CitizenEscapedPayload struct {
	SafeZoneID       uuid.UUID
	AssignedPosition pf.Position
}

// CitizenEscaped raised when a citizen escapes hazards by reaching a safe zone
func (e *EventBus) CitizenEscaped(citizenID uuid.UUID, payload CitizenEscapedPayload, metadata EventMetadata) {
	e.createEvent(CitizenEscaped, citizenID, metadata, payload)
}

// CitizenDied raised when a citizen has been killed by a hazard
func (e *EventBus) CitizenDied(citizenID uuid.UUID, metadata EventMetadata) {
	e.createEvent(CitizenDied, citizenID, metadata, nil)
}

// SafeZoneEmergedPayload for safe zone emergence event
type SafeZoneEmergedPayload struct {
	ID    uuid.UUID
	Cells []pf.Position
}

// SafeZoneEmerged raised when a safe zone emerges
func (e *EventBus) SafeZoneEmerged(safeZoneID uuid.UUID, payload SafeZoneEmergedPayload, metadata EventMetadata) {
	e.createEvent(SafeZoneEmerged, safeZoneID, metadata, SafeZoneEmergedPayload{
		ID:    payload.ID,
		Cells: getPositionSnapshot(payload.Cells),
	})
}

// HazardEmergedPayload for hazard emergence event
type HazardEmergedPayload struct {
	Type     c.HazardType
	Position pf.Position
}

// HazardEmerged raised when a hazard emerges
func (e *EventBus) HazardEmerged(hazardID uuid.UUID, payload HazardEmergedPayload, metadata EventMetadata) {
	e.createEvent(HazardEmerged, hazardID, metadata, payload)
}

// HazardExpanded raised when a hazard expands
func (e *EventBus) HazardExpanded(hazardID uuid.UUID, updatedCells []pf.Position, metadata EventMetadata) {
	e.createEvent(HazardExpanded, hazardID, metadata, getPositionSnapshot(updatedCells))
}

// HazardDissipated raised when a hazard disappears
func (e *EventBus) HazardDissipated(hazardID uuid.UUID, updatedCells []pf.Position, metadata EventMetadata) {
	e.createEvent(HazardDissipated, hazardID, metadata, getPositionSnapshot(updatedCells))
}

func (e *EventBus) createEvent(eventType eventType, entityID uuid.UUID, metadata EventMetadata, payload any) {
	event := SimulationEvent{
		ID:           uuid.New(),
		Timestamp:    time.Now().UTC(),
		SimulationID: metadata.SimulationID,
		Tick:         metadata.Tick,
		EventType:    eventType,
		EntityID:     entityID,
		Payload:      payload,
	}

	// Sending to channel is blocking to ensure event dispatch and event log stay in sync
	e.SimulationEvents <- event
	e.EventLog = append(e.EventLog, event)
}

func getPositionSnapshot(src []pf.Position) []pf.Position {
	snapshot := make([]pf.Position, len(src))
	copy(snapshot, src)

	return snapshot
}
