package tui

import (
	c "hazard/internal/common"
	"hazard/internal/events"
	pf "hazard/internal/pathfinding"
	"log"

	"github.com/google/uuid"
)

func (m *Model) handleSimulationCreated(event events.SimulationEvent) {
	payload, ok := event.Payload.(events.SimulationCreatedPayload)
	if !ok {
		log.Printf("error converting payload to SimulationCreatedPayload: %v", event.Payload)
	}

	// Reset model state
	m.simulationID = event.SimulationID
	m.citizens = map[uuid.UUID]citizenState{}
	m.hazards = map[uuid.UUID]string{}

	// Initialise open cells
	m.grid = make([][]string, payload.Grid.Height)
	for y := range payload.Grid.Height {
		m.grid[y] = make([]string, payload.Grid.Width)
		for x := range payload.Grid.Width {
			m.grid[y][x] = getOpenCell()
		}
	}

	// Initialise safe zones
	for _, safeZone := range payload.SafeZones {
		m.createSafeZone(safeZone.Cells)
	}

	// Initialise citizens
	for _, citizen := range payload.Citizens {
		pos := citizen.CurrentPosition
		m.grid[pos.Y][pos.X] = getCitizenCell()
		m.citizens[citizen.ID] = citizenState{
			Position: citizen.CurrentPosition,
		}
	}
}

func (m *Model) handleCitizenMoved(event events.SimulationEvent) {
	newPosition, ok := event.Payload.(pf.Position)
	if !ok {
		log.Printf("error converting payload to pf.Position: %v", event.Payload)
	}

	state := m.citizens[event.EntityID]

	// Restore previous cell state after movement
	if state.PreviousCell != "" {
		m.grid[state.Position.Y][state.Position.X] = state.PreviousCell
	} else {
		m.grid[state.Position.Y][state.Position.X] = getOpenCell()
	}

	m.citizens[event.EntityID] = citizenState{
		Position:     newPosition,
		PreviousCell: m.grid[newPosition.Y][newPosition.X],
	}
	m.grid[newPosition.Y][newPosition.X] = getCitizenCell()
}

func (m *Model) handleCitizenEscaped(event events.SimulationEvent) {
	var currentPosition = m.citizens[event.EntityID].Position
	m.grid[currentPosition.Y][currentPosition.X] = getEscapedCitizenCell()
}

func (m *Model) handleCitizenDied(event events.SimulationEvent) {
	var currentPosition = m.citizens[event.EntityID].Position
	m.grid[currentPosition.Y][currentPosition.X] = getDeadCitizenCell()
}

func (m *Model) handleSafeZoneEmerged(event events.SimulationEvent) {
	cells, ok := event.Payload.([]pf.Position)
	if !ok {
		log.Printf("error converting payload to []pf.Position: %v", event.Payload)
	}

	m.createSafeZone(cells)
}

func (m *Model) handleHazardEmerged(event events.SimulationEvent) {
	payload, ok := event.Payload.(events.HazardEmergedPayload)
	if !ok {
		log.Printf("error converting payload to HazardEmergedPayload: %v", event.Payload)
	}

	var character string

	switch payload.Type {
	case c.FireHazard:
		character = getFireCell()
	case c.FloodHazard:
		character = getFloodCell()
	case c.LavaHazard:
		character = getLavaCell()
	}

	m.grid[payload.Position.Y][payload.Position.X] = character
	m.hazards[event.EntityID] = character
}

func (m *Model) handleHazardExpanded(event events.SimulationEvent) {
	updatedCells, ok := event.Payload.([]pf.Position)
	if !ok {
		log.Printf("error converting payload to []pf.Position: %v", event.Payload)
	}

	character := m.hazards[event.EntityID]

	for _, cell := range updatedCells {
		m.grid[cell.Y][cell.X] = character
	}
}

func (m *Model) handleHazardDissipated(event events.SimulationEvent) {
	updatedCells, ok := event.Payload.([]pf.Position)
	if !ok {
		log.Printf("error converting payload to []pf.Position: %v", event.Payload)
	}

	for _, cell := range updatedCells {
		m.grid[cell.Y][cell.X] = getOpenCell()
	}
}

func (m *Model) createSafeZone(cells []pf.Position) {
	safeZoneChar := getSafeZoneCell(c.RandValInSlice(safeZoneCharacters))
	for _, safeZoneCell := range cells {
		m.grid[safeZoneCell.Y][safeZoneCell.X] = safeZoneChar
	}
}
