package tui

import (
	"log/slog"

	"github.com/google/uuid"

	"hazard/internal/citizen"
	"hazard/internal/events"
	"hazard/internal/hazard"
	"hazard/internal/pathfinding"
	"hazard/internal/random"
)

func (m *Model) handleSimulationCreated(event events.SimulationEvent) {
	payload, ok := event.Payload.(events.SimulationCreatedPayload)
	if !ok {
		slog.Error("error converting payload to SimulationCreatedPayload", "payload", event.Payload)
		return
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
		m.createSafeZone(safeZone.ID, safeZone.Cells)
	}

	// Initialise citizens
	for _, c := range payload.Citizens {
		pos := c.CurrentPosition
		m.grid[pos.Y][pos.X] = citizenCharacter
		m.citizens[c.ID] = citizenState{
			Position: c.CurrentPosition,
			Status:   c.Status,
		}
		m.paths[c.ID] = c.Path
	}
	m.activeCitizenCount = len(payload.Citizens)

	// Initialise obstacles
	for _, obstacle := range payload.Obstacles {
		obstacleChar := getObstacleCell()
		for _, cellPos := range obstacle.Cells {
			m.grid[cellPos.Y][cellPos.X] = obstacleChar
		}
	}

	if payload.UseLogoObstacles && logoCharWidth > 0 && logoCharHeight > 0 && len(m.grid) > 0 && len(m.grid[0]) > 0 {
		startX := (len(m.grid[0]) - logoCharWidth) / 2
		startY := (len(m.grid) - logoCharHeight) / 2
		if startX < 0 {
			startX = 0
		}
		if startY < 0 {
			startY = 0
		}
		placeLogoOnGrid(m.grid, startX, startY)
	}
}

func (m *Model) handleCitizenMoved(event events.SimulationEvent) {
	newPosition, ok := event.Payload.(pathfinding.Position)
	if !ok {
		slog.Error("error converting payload to pf.Position", "payload", event.Payload)
		return
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
		Status:       citizen.StatusNavigating,
	}
	m.grid[newPosition.Y][newPosition.X] = citizenCharacter
}

func (m *Model) handleCitizenPathUpdated(event events.SimulationEvent) {
	path, ok := event.Payload.([]pathfinding.Position)
	if !ok {
		slog.Error("error converting payload to []pathfinding.Position", "payload", event.Payload)
		return
	}

	m.paths[event.EntityID] = path
}

func (m *Model) handleCitizenEscaped(event events.SimulationEvent) {
	payload, ok := event.Payload.(events.CitizenEscapedPayload)
	if !ok {
		slog.Error("error converting payload to CitizenEscapedPayload", "payload", event.Payload)
		return
	}

	assignedPos := payload.AssignedPosition
	curPos := m.citizens[event.EntityID].Position

	if assignedPos != curPos {
		// Citizen was redirected; restore the arrival cell
		m.grid[curPos.Y][curPos.X] = m.citizens[event.EntityID].PreviousCell
	}

	m.grid[assignedPos.Y][assignedPos.X] = citizenEscapedCharacter

	state := m.citizens[event.EntityID]
	state.Position = assignedPos
	state.Status = citizen.StatusEscaped
	m.citizens[event.EntityID] = state

	delete(m.paths, event.EntityID)

	m.escapedCitizenCount = payload.TotalEscaped
	m.activeCitizenCount = payload.TotalRemaining
}

func (m *Model) handleCitizenDied(event events.SimulationEvent) {
	payload, ok := event.Payload.(events.CitizenDiedPayload)
	if !ok {
		slog.Error("error converting payload to CitizenDiedPayload", "payload", event.Payload)
		return
	}

	currentPosition := m.citizens[event.EntityID].Position
	m.grid[currentPosition.Y][currentPosition.X] = citizenDeadCharacter

	state := m.citizens[event.EntityID]
	state.Status = citizen.StatusDead
	m.citizens[event.EntityID] = state

	delete(m.paths, event.EntityID)

	m.deadCitizenCount = payload.TotalDead
	m.activeCitizenCount = payload.TotalRemaining
}

func (m *Model) handleSafeZoneEmerged(event events.SimulationEvent) {
	payload, ok := event.Payload.(events.SafeZoneEmergedPayload)
	if !ok {
		slog.Error("error converting payload to SafeZoneEmergedPayload", "payload", event.Payload)
		return
	}

	m.createSafeZone(payload.ID, payload.Cells)
}

func (m *Model) handleHazardEmerged(event events.SimulationEvent) {
	payload, ok := event.Payload.(events.HazardEmergedPayload)
	if !ok {
		slog.Error("error converting payload to HazardEmergedPayload", "payload", event.Payload)
		return
	}

	var character string

	switch payload.Type {
	case hazard.FireHazard:
		character = getFireCell()
	case hazard.FloodHazard:
		character = getFloodCell()
	case hazard.LavaHazard:
		character = getLavaCell()
	}

	m.grid[payload.Position.Y][payload.Position.X] = character
	m.hazards[event.EntityID] = character
}

func (m *Model) handleHazardExpanded(event events.SimulationEvent) {
	updatedCells, ok := event.Payload.([]pathfinding.Position)
	if !ok {
		slog.Error("error converting payload to []pf.Position", "payload", event.Payload)
		return
	}

	character := m.hazards[event.EntityID]

	for _, cell := range updatedCells {
		m.grid[cell.Y][cell.X] = character
	}
}

func (m *Model) handleHazardDissipated(event events.SimulationEvent) {
	updatedCells, ok := event.Payload.([]pathfinding.Position)
	if !ok {
		slog.Error("error converting payload to []pf.Position", "payload", event.Payload)
		return
	}

	for _, cell := range updatedCells {
		m.grid[cell.Y][cell.X] = getOpenCell()
	}


}

func (m *Model) createSafeZone(safeZoneID uuid.UUID, cells []pathfinding.Position) {
	safeZoneChar := getSafeZoneCell(random.ValInSlice(safeZoneCharacters))
	m.safeZones[safeZoneID] = make([]pathfinding.Position, len(cells))

	for i, cellPos := range cells {
		m.grid[cellPos.Y][cellPos.X] = safeZoneChar
		m.safeZones[safeZoneID][i] = cellPos
	}
}
