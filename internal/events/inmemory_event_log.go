package events

import (
	"hazard/internal/pathfinding"

	"github.com/google/uuid"
)

// InMemoryEventLog for creating and story simulation events in-memory
type InMemoryEventLog struct {
	EventLog []SimulationEvent
}

// Events returns list of events in ascending order
func (e *InMemoryEventLog) Events() []SimulationEvent {
	return e.EventLog
}

// SimulationStarted raised when simulation starts
func (e *InMemoryEventLog) SimulationStarted(metadata EventMetadata) error {
	event, err := createEvent(simulationStarted, metadata.SimulationID, metadata, nil)
	e.EventLog = append(e.EventLog, event)
	return err
}

// SimulationCompleted raised when simulation finishes
func (e *InMemoryEventLog) SimulationCompleted(metadata EventMetadata) error {
	event, err := createEvent(simulationCompleted, metadata.SimulationID, metadata, nil)
	e.EventLog = append(e.EventLog, event)
	return err
}

// CitizenMoved raised when a citizen moves
func (e *InMemoryEventLog) CitizenMoved(citizenID uuid.UUID, newPosition pathfinding.Position, metadata EventMetadata) error {
	event, err := createEvent(citizenMoved, citizenID, metadata, newPosition)
	e.EventLog = append(e.EventLog, event)
	return err
}

// CitizenPathUpdated raised when a citizen's path is recalculated
func (e *InMemoryEventLog) CitizenPathUpdated(citizenID uuid.UUID, path []pathfinding.Position, metadata EventMetadata) error {
	event, err := createEvent(citizenMoved, citizenID, metadata, path)
	e.EventLog = append(e.EventLog, event)
	return err
}

// CitizenEscaped raised when a citizen escapes hazards by reaching a safe zone
func (e *InMemoryEventLog) CitizenEscaped(citizenID uuid.UUID, metadata EventMetadata) error {
	// TODO: ideally we would have some sort of relationship between citizens and the safe zone they've occupied
	// - this is out of scope for now
	event, err := createEvent(citizenEscaped, citizenID, metadata, nil)
	e.EventLog = append(e.EventLog, event)
	return err
}

// CitizenDied raised when a citizen has been killed by a hazard
func (e *InMemoryEventLog) CitizenDied(citizenID uuid.UUID, metadata EventMetadata) error {
	event, err := createEvent(citizenDied, citizenID, metadata, nil)
	e.EventLog = append(e.EventLog, event)
	return err
}

type safeZoneEmergedPayload struct {
	Position pathfinding.Position
	Radius   int
}

// SafeZoneEmerged raised when a safe zone emerges
func (e *InMemoryEventLog) SafeZoneEmerged(safeZoneID uuid.UUID, position pathfinding.Position, radius int, metadata EventMetadata) error {

	event, err := createEvent(safeZoneEmerged, safeZoneID, metadata, safeZoneEmergedPayload{
		Position: position,
		Radius:   radius,
	})
	e.EventLog = append(e.EventLog, event)
	return err
}

// HazardEmerged raised when a hazard emerges
func (e *InMemoryEventLog) HazardEmerged(hazardID uuid.UUID, position pathfinding.Position, metadata EventMetadata) error {
	event, err := createEvent(citizenDied, hazardID, metadata, position)
	e.EventLog = append(e.EventLog, event)
	return err
}

// HazardExpanded raised when a hazard expands
func (e *InMemoryEventLog) HazardExpanded(hazardID uuid.UUID, radius int, metadata EventMetadata) error {
	event, err := createEvent(citizenDied, hazardID, metadata, radius)
	e.EventLog = append(e.EventLog, event)
	return err
}

// HazardDissipated raised when a hazard disappears
func (e *InMemoryEventLog) HazardDissipated(hazardID uuid.UUID, metadata EventMetadata) error {
	event, err := createEvent(citizenDied, hazardID, metadata, nil)
	e.EventLog = append(e.EventLog, event)
	return err
}
