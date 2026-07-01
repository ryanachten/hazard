package tui

import (
	c "hazard/internal/common"
	"hazard/internal/events"
	pf "hazard/internal/pathfinding"
	"log"
)

func (m *Model) handleSimulationCreated(event events.SimulationEvent) {
	payload, ok := event.Payload.(events.SimulationCreatedPayload)
	if !ok {
		log.Printf("error converting payload to SimulationCreatedPayload: %v", event.Payload)
	}

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
		safeZoneChar := c.RandValInSlice(safeZoneCharacters)
		for _, safeZoneCell := range safeZone.Cells {
			m.grid[safeZoneCell.Y][safeZoneCell.X] = getSafeZoneCell(safeZoneChar)
		}
	}

	// Initialise citizens
	for _, citizen := range payload.Citizens {
		pos := citizen.CurrentPosition
		m.grid[pos.Y][pos.X] = getCitizenCell()
		m.citizens[citizen.ID] = citizen.CurrentPosition
	}
}

func (m *Model) handleCitizenMoved(event events.SimulationEvent) {
	newPosition, ok := event.Payload.(pf.Position)
	if !ok {
		log.Printf("error converting payload to pf.Position: %v", event.Payload)
	}

	var currentPosition = m.citizens[event.EntityID]
	m.grid[currentPosition.Y][currentPosition.X] = getOpenCell()

	m.grid[newPosition.Y][newPosition.X] = getCitizenCell()
	m.citizens[event.EntityID] = newPosition
}

func (m *Model) handleCitizenEscaped(event events.SimulationEvent) {
	var currentPosition = m.citizens[event.EntityID]
	m.grid[currentPosition.Y][currentPosition.X] = getEscapedCitizenCell()
}

func (m *Model) handleCitizenDied(event events.SimulationEvent) {
	var currentPosition = m.citizens[event.EntityID]
	m.grid[currentPosition.Y][currentPosition.X] = getDeadCitizenCell()
}

func (m *Model) handleSafeZoneEmerged(event events.SimulationEvent) {
	cells, ok := event.Payload.([]pf.Position)
	if !ok {
		log.Printf("error converting payload to []pf.Position: %v", event.Payload)
	}

	character := c.RandValInSlice(safeZoneCharacters)

	for _, cell := range cells {
		m.grid[cell.Y][cell.X] = getSafeZoneCell(character)
	}
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
