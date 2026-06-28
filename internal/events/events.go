// Package events defines simulation events
package events

import (
	"encoding/json"
	c "hazard/internal/common"
	"hazard/internal/pathfinding"
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
	Payload      json.RawMessage
}

// EventMetadata common for all simulation events
type EventMetadata struct {
	SimulationID uuid.UUID
	Tick         uint64
}

// SimulationEventChannel to subscribe to simulation events
var SimulationEventChannel = make(chan SimulationEvent)

type eventType string

const (
	// SimulationStarted event type indicates the simulation has started
	SimulationStarted eventType = "simulation.started"
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

// SimulationStartedPayload for simulation start event
type SimulationStartedPayload struct {
	Grid      *pathfinding.Grid
	Citizens  []c.Citizen
	SafeZones []c.SafeZone
}

// EventEmitter defines how events are emitted in the simulation
type EventEmitter interface {
	Events() []SimulationEvent
	SimulationStarted(payload SimulationStartedPayload, metadata EventMetadata) error
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
		ID:           uuid.New(),
		Timestamp:    time.Now().UTC(),
		SimulationID: metadata.SimulationID,
		Tick:         metadata.Tick,
		EventType:    eventType,
		EntityID:     entityID,
		Payload:      json.RawMessage(jsonPayload),
	}, nil
}
