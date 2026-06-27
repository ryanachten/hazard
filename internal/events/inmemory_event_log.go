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
	if err != nil {
		return err
	}
	e.EventLog = append(e.EventLog, event)
	return nil
}

// SimulationCompleted raised when simulation finishes
func (e *InMemoryEventLog) SimulationCompleted(metadata EventMetadata) error {
	event, err := createEvent(simulationCompleted, metadata.SimulationID, metadata, nil)
	if err != nil {
		return err
	}
	e.EventLog = append(e.EventLog, event)
	return nil
}

// CitizenMoved raised when a citizen moves
func (e *InMemoryEventLog) CitizenMoved(citizenID uuid.UUID, newPosition pathfinding.Position, metadata EventMetadata) error {
	event, err := createEvent(citizenMoved, citizenID, metadata, newPosition)
	if err != nil {
		return err
	}
	e.EventLog = append(e.EventLog, event)
	return nil
}

// CitizenPathUpdated raised when a citizen's path is recalculated
func (e *InMemoryEventLog) CitizenPathUpdated(citizenID uuid.UUID, path []pathfinding.Position, metadata EventMetadata) error {
	event, err := createEvent(citizenPathUpdated, citizenID, metadata, path)
	if err != nil {
		return err
	}
	e.EventLog = append(e.EventLog, event)
	return nil
}

// CitizenEscaped raised when a citizen escapes hazards by reaching a safe zone
func (e *InMemoryEventLog) CitizenEscaped(citizenID uuid.UUID, metadata EventMetadata) error {
	// TODO: ideally we would have some sort of relationship between citizens and the safe zone they've occupied
	// - this is out of scope for now
	event, err := createEvent(citizenEscaped, citizenID, metadata, nil)
	if err != nil {
		return err
	}
	e.EventLog = append(e.EventLog, event)
	return nil
}

// CitizenDied raised when a citizen has been killed by a hazard
func (e *InMemoryEventLog) CitizenDied(citizenID uuid.UUID, metadata EventMetadata) error {
	event, err := createEvent(citizenDied, citizenID, metadata, nil)
	if err != nil {
		return err
	}
	e.EventLog = append(e.EventLog, event)
	return nil
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
	if err != nil {
		return err
	}
	e.EventLog = append(e.EventLog, event)
	return nil
}

// HazardEmerged raised when a hazard emerges
func (e *InMemoryEventLog) HazardEmerged(hazardID uuid.UUID, position pathfinding.Position, metadata EventMetadata) error {
	event, err := createEvent(hazardEmerged, hazardID, metadata, position)
	if err != nil {
		return err
	}
	e.EventLog = append(e.EventLog, event)
	return nil
}

// HazardExpanded raised when a hazard expands
func (e *InMemoryEventLog) HazardExpanded(hazardID uuid.UUID, radius int, metadata EventMetadata) error {
	event, err := createEvent(hazardExpanded, hazardID, metadata, radius)
	if err != nil {
		return err
	}
	e.EventLog = append(e.EventLog, event)
	return nil
}

// HazardDissipated raised when a hazard disappears
func (e *InMemoryEventLog) HazardDissipated(hazardID uuid.UUID, metadata EventMetadata) error {
	event, err := createEvent(hazardDissipated, hazardID, metadata, nil)
	if err != nil {
		return err
	}
	e.EventLog = append(e.EventLog, event)
	return nil
}
