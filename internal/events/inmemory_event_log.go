package events

import (
	pf "hazard/internal/pathfinding"

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
func (e *InMemoryEventLog) SimulationStarted(payload SimulationStartedPayload, metadata EventMetadata) error {
	event, err := createEvent(SimulationStarted, metadata.SimulationID, metadata, payload)
	if err != nil {
		return err
	}
	e.EventLog = append(e.EventLog, event)
	return nil
}

// SimulationCompleted raised when simulation finishes
func (e *InMemoryEventLog) SimulationCompleted(metadata EventMetadata) error {
	event, err := createEvent(SimulationCompleted, metadata.SimulationID, metadata, nil)
	if err != nil {
		return err
	}
	e.EventLog = append(e.EventLog, event)
	return nil
}

// CitizenMoved raised when a citizen moves
func (e *InMemoryEventLog) CitizenMoved(citizenID uuid.UUID, newPosition pf.Position, metadata EventMetadata) error {
	event, err := createEvent(CitizenMoved, citizenID, metadata, newPosition)
	if err != nil {
		return err
	}
	e.EventLog = append(e.EventLog, event)
	return nil
}

// CitizenPathUpdated raised when a citizen's path is recalculated
func (e *InMemoryEventLog) CitizenPathUpdated(citizenID uuid.UUID, path []pf.Position, metadata EventMetadata) error {
	event, err := createEvent(CitizenPathUpdated, citizenID, metadata, path)
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
	event, err := createEvent(CitizenEscaped, citizenID, metadata, nil)
	if err != nil {
		return err
	}
	e.EventLog = append(e.EventLog, event)
	return nil
}

// CitizenDied raised when a citizen has been killed by a hazard
func (e *InMemoryEventLog) CitizenDied(citizenID uuid.UUID, metadata EventMetadata) error {
	event, err := createEvent(CitizenDied, citizenID, metadata, nil)
	if err != nil {
		return err
	}
	e.EventLog = append(e.EventLog, event)
	return nil
}

// SafeZoneEmerged raised when a safe zone emerges
func (e *InMemoryEventLog) SafeZoneEmerged(safeZoneID uuid.UUID, cells []pf.Position, metadata EventMetadata) error {
	event, err := createEvent(SafeZoneEmerged, safeZoneID, metadata, cells)
	if err != nil {
		return err
	}
	e.EventLog = append(e.EventLog, event)
	return nil
}

// HazardEmerged raised when a hazard emerges
func (e *InMemoryEventLog) HazardEmerged(hazardID uuid.UUID, payload HazardEmergedPayload, metadata EventMetadata) error {
	event, err := createEvent(HazardEmerged, hazardID, metadata, payload)
	if err != nil {
		return err
	}
	e.EventLog = append(e.EventLog, event)
	return nil
}

// HazardExpanded raised when a hazard expands
func (e *InMemoryEventLog) HazardExpanded(hazardID uuid.UUID, updatedCells []pf.Position, metadata EventMetadata) error {
	event, err := createEvent(HazardExpanded, hazardID, metadata, updatedCells)
	if err != nil {
		return err
	}
	e.EventLog = append(e.EventLog, event)
	return nil
}

// HazardDissipated raised when a hazard disappears
func (e *InMemoryEventLog) HazardDissipated(hazardID uuid.UUID, updatedCells []pf.Position, metadata EventMetadata) error {
	event, err := createEvent(HazardDissipated, hazardID, metadata, updatedCells)
	if err != nil {
		return err
	}
	e.EventLog = append(e.EventLog, event)
	return nil
}
