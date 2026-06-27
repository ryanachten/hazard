// Package events defines simulation events
package events

import (
	"encoding/json"
	"hazard/internal/pathfinding"
	"time"

	"github.com/google/uuid"
)

// SimulationEvent represents a domain event that occurred during a simulation
type SimulationEvent struct {
	ID        uuid.UUID
	EventType eventType
	Timestamp time.Time
	EntityID  uuid.UUID
	Metadata  EventMetadata
	Payload   []byte
}

// EventMetadata common for all simulation events
type EventMetadata struct {
	SimulationID uuid.UUID
	Tick         uint64
}

type eventType string

const (
	simulationStarted   eventType = "simulation.started"
	simulationCompleted eventType = "simulation.completed"
	citizenMoved        eventType = "citizen.moved"
	citizenPathUpdated  eventType = "citizen.pathUpdated"
	citizenEscaped      eventType = "citizen.escaped"
	citizenDied         eventType = "citizen.died"
	safeZoneEmerged     eventType = "safeZone.emerged"
	hazardEmerged       eventType = "hazard.emerged"
	hazardExpanded      eventType = "hazard.expanded"
	hazardDissipated    eventType = "hazard.dissipated"
)

// EventEmitter defines how events are emitted in the simulation
type EventEmitter interface {
	Events() []SimulationEvent
	SimulationStarted(metadata EventMetadata) error
	SimulationCompleted(metadata EventMetadata) error
	CitizenMoved(citizenID uuid.UUID, newPosition pathfinding.Position, metadata EventMetadata) error
	CitizenPathUpdated(citizenID uuid.UUID, path []pathfinding.Position, metadata EventMetadata) error
	CitizenEscaped(citizenID uuid.UUID, metadata EventMetadata) error
	CitizenDied(citizenID uuid.UUID, metadata EventMetadata) error
	SafeZoneEmerged(safeZoneID uuid.UUID, position pathfinding.Position, radius int, metadata EventMetadata) error
	HazardEmerged(hazardID uuid.UUID, position pathfinding.Position, metadata EventMetadata) error
	HazardExpanded(hazardID uuid.UUID, radius int, metadata EventMetadata) error
	HazardDissipated(hazardID uuid.UUID, metadata EventMetadata) error
}

func createEvent(eventType eventType, entityID uuid.UUID, metadata EventMetadata, payload any) (SimulationEvent, error) {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return SimulationEvent{}, err
	}

	return SimulationEvent{
		ID:        uuid.New(),
		Timestamp: time.Now().UTC(),
		EventType: eventType,
		EntityID:  entityID,
		Metadata:  metadata,
		Payload:   jsonPayload,
	}, nil
}
