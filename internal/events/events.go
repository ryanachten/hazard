// Package events defines simulation events
package events

import (
	"encoding/json"
	"hazard/internal/pathfinding"
	"log"
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
	citizenEscaped      eventType = "citizen.escaped"
	citizenDied         eventType = "citizen.died"
)

// EventLog for storing simulation events
type EventLog struct {
	Events []SimulationEvent
}

// SimulationStarted raised when simulation starts
func (e *EventLog) SimulationStarted(metadata EventMetadata) {
	e.Events = append(e.Events, createEvent(simulationStarted, metadata.SimulationID, metadata, nil))
}

// SimulationCompleted raised when simulation finishes
func (e *EventLog) SimulationCompleted(metadata EventMetadata) {
	e.Events = append(e.Events, createEvent(simulationCompleted, metadata.SimulationID, metadata, nil))
}

// CitizenMoved raised when a citizen moves
func (e *EventLog) CitizenMoved(citizenID uuid.UUID, newPosition pathfinding.Position, metadata EventMetadata) {
	e.Events = append(e.Events, createEvent(citizenMoved, citizenID, metadata, newPosition))
}

// CitizenEscaped raised when a citizen escapes hazards by reaching a safe zone
func (e *EventLog) CitizenEscaped(citizenID uuid.UUID, metadata EventMetadata) {
	// TODO: ideally we would have some sort of relationship between citizens and the safe zone they've occupied
	// - this is out of scope for now
	e.Events = append(e.Events, createEvent(citizenEscaped, citizenID, metadata, nil))
}

// CitizenDied raised when a citizen has been killed by a hazard
func (e *EventLog) CitizenDied(citizenID uuid.UUID, metadata EventMetadata) {
	e.Events = append(e.Events, createEvent(citizenDied, citizenID, metadata, nil))
}

func createEvent(eventType eventType, entityID uuid.UUID, metadata EventMetadata, payload any) SimulationEvent {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("error encoding JSON for %v with payload %v: %v", eventType, payload, err)
	}

	return SimulationEvent{
		ID:        uuid.New(),
		Timestamp: time.Now().UTC(),
		EventType: eventType,
		EntityID:  entityID,
		Metadata:  metadata,
		Payload:   jsonPayload,
	}
}
