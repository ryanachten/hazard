// Package tui provides user interface for hazard simulation
package tui

import (
	"encoding/json"
	c "hazard/internal/common"
	"hazard/internal/events"
	pf "hazard/internal/pathfinding"
	"log"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/google/uuid"
)

// Model represents the TUI state for the hazard simulation
type Model struct {
	SimulationEvents chan events.SimulationEvent
	Grid             [][]string
	Citizens         map[uuid.UUID]pf.Position
	Hazards          map[uuid.UUID]string
}

// InitialModel creates the initial TUI model state
func InitialModel() Model {
	return Model{
		SimulationEvents: events.SimulationEventChannel,
		Grid:             [][]string{},
		Citizens:         map[uuid.UUID]pf.Position{},
		Hazards:          map[uuid.UUID]string{},
	}
}

func (m Model) consumeEvent() tea.Msg {
	event := <-m.SimulationEvents
	return event
}

// Init initializes the Bubble Tea program
func (m Model) Init() tea.Cmd {
	return m.consumeEvent
}

// View renders the current simulation state
func (m Model) View() tea.View {
	var s strings.Builder

	s.WriteString("Running!\n")

	for y := range m.Grid {
		for x := range m.Grid[y] {
			cell := m.Grid[y][x]
			s.WriteString(cell)
		}
		s.WriteString("\n")
	}

	return tea.NewView(s.String())
}

// Update handles messages and updates the TUI model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	case events.SimulationEvent:
		event := events.SimulationEvent(msg)
		switch event.EventType {
		case events.SimulationStarted:
			m.handleSimulationStarted(event)
		case events.SimulationCompleted:
			return m, tea.Quit
		case events.CitizenMoved:
			m.handleCitizenMoved(event)
		case events.CitizenDied:
			m.handleCitizenDied(event)
		case events.SafeZoneEmerged:
			m.handleSafeZoneEmerged(event)
		case events.CitizenEscaped:
			m.handleCitizenEscaped(event)
		case events.HazardEmerged:
			m.handleHazardEmerged(event)
		case events.HazardExpanded:
			m.handleHazardExpanded(event)
		case events.HazardDissipated:
			m.handleHazardDissipated(event)
		}

		return m, m.consumeEvent

	// // Is it a key press?
	case tea.KeyPressMsg:

		// 	// Cool, what was the actual key pressed?
		switch msg.String() {

		// 	// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

			// 	// The "up" and "k" keys move the cursor up
			// 	case "up", "k":
			// 		if m.cursor > 0 {
			// 			m.cursor--
			// 		}

			// 	// The "down" and "j" keys move the cursor down
			// 	case "down", "j":
			// 		if m.cursor < len(m.choices)-1 {
			// 			m.cursor++
			// 		}

			// 	// The "enter" key and the space bar toggle the selected state
			// 	// for the item that the cursor is pointing at.
			// 	case "enter", "space":
			// 		_, ok := m.selected[m.cursor]
			// 		if ok {
			// 			delete(m.selected, m.cursor)
			// 		} else {
			// 			m.selected[m.cursor] = struct{}{}
			// 		}
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m *Model) handleSimulationStarted(event events.SimulationEvent) {
	var payload events.SimulationStartedPayload
	err := json.Unmarshal(event.Payload, &payload)
	if err != nil {
		log.Printf("error parsing event: %v", err)
	}

	// Initialise open cells
	m.Grid = make([][]string, payload.Grid.Height)
	for y := range payload.Grid.Height {
		m.Grid[y] = make([]string, payload.Grid.Width)
		for x := range payload.Grid.Width {
			m.Grid[y][x] = getOpenCell()
		}
	}

	// Initialise safe zones
	for _, safeZone := range payload.SafeZones {
		safeZoneChar := c.RandValInSlice(safeZoneCharacters)
		for _, safeZoneCell := range safeZone.Cells {
			m.Grid[safeZoneCell.Y][safeZoneCell.X] = getSafeZoneCell(safeZoneChar)
		}
	}

	// Initialise citizens
	for _, citizen := range payload.Citizens {
		pos := citizen.CurrentPosition
		m.Grid[pos.Y][pos.X] = getCitizenCell()
		m.Citizens[citizen.ID] = citizen.CurrentPosition
	}
}

func (m *Model) handleCitizenMoved(event events.SimulationEvent) {
	var newPosition pf.Position
	err := json.Unmarshal(event.Payload, &newPosition)
	if err != nil {
		log.Printf("error parsing event: %v", err)
	}

	var currentPosition = m.Citizens[event.EntityID]
	m.Grid[currentPosition.Y][currentPosition.X] = getOpenCell()

	m.Grid[newPosition.Y][newPosition.X] = getCitizenCell()
	m.Citizens[event.EntityID] = newPosition
}

func (m *Model) handleCitizenEscaped(event events.SimulationEvent) {
	var currentPosition = m.Citizens[event.EntityID]
	m.Grid[currentPosition.Y][currentPosition.X] = getEscapedCitizenCell()
}

func (m *Model) handleCitizenDied(event events.SimulationEvent) {
	var currentPosition = m.Citizens[event.EntityID]
	m.Grid[currentPosition.Y][currentPosition.X] = getDeadCitizenCell()
}

func (m *Model) handleSafeZoneEmerged(event events.SimulationEvent) {
	var cells []pf.Position
	err := json.Unmarshal(event.Payload, &cells)
	if err != nil {
		log.Printf("error parsing event: %v", err)
	}

	character := c.RandValInSlice(safeZoneCharacters)

	for _, cell := range cells {
		m.Grid[cell.Y][cell.X] = getSafeZoneCell(character)
	}
}

func (m *Model) handleHazardEmerged(event events.SimulationEvent) {
	var payload events.HazardEmergedPayload
	err := json.Unmarshal(event.Payload, &payload)
	if err != nil {
		log.Printf("error parsing event: %v", err)
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

	m.Grid[payload.Position.Y][payload.Position.X] = character
	m.Hazards[event.EntityID] = character
}

func (m *Model) handleHazardExpanded(event events.SimulationEvent) {
	var updatedCells []pf.Position
	err := json.Unmarshal(event.Payload, &updatedCells)
	if err != nil {
		log.Printf("error parsing event: %v", err)
	}

	character := m.Hazards[event.EntityID]

	for _, cell := range updatedCells {
		m.Grid[cell.Y][cell.X] = character
	}
}

func (m *Model) handleHazardDissipated(event events.SimulationEvent) {
	var updatedCells []pf.Position
	err := json.Unmarshal(event.Payload, &updatedCells)
	if err != nil {
		log.Printf("error parsing event: %v", err)
	}

	for _, cell := range updatedCells {
		m.Grid[cell.Y][cell.X] = getOpenCell()
	}
}
