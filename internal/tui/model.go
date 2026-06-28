// Package tui provides user interface for hazard simulation
package tui

import (
	"encoding/json"
	"hazard/internal/common"
	"hazard/internal/events"
	"hazard/internal/pathfinding"
	"log"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// Model represents the TUI state for the hazard simulation
type Model struct {
	SimulationEvents chan events.SimulationEvent
	Grid             [][]rune
	Citizens         []common.Citizen
}

// InitialModel creates the initial TUI model state
func InitialModel() Model {
	return Model{
		SimulationEvents: events.SimulationEventChannel,
		Grid:             [][]rune{},
		Citizens:         []common.Citizen{},
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
			s.WriteString(string(cell))
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
	m.Grid = make([][]rune, payload.Grid.Height)
	for y := range payload.Grid.Height {
		m.Grid[y] = make([]rune, payload.Grid.Width)
		for x := range payload.Grid.Width {
			chars := cellCharacters[pathfinding.CellOpen]
			m.Grid[y][x] = common.RandValInSlice(chars[:])
		}
	}

	// Initialise safe zones
	safeZoneChars := cellCharacters[pathfinding.CellSafeZone]
	for _, safeZone := range payload.SafeZones {
		safeZoneChar := common.RandValInSlice(safeZoneChars[:])
		for _, safeZoneCell := range safeZone.Cells {
			m.Grid[safeZoneCell.Y][safeZoneCell.X] = safeZoneChar
		}
	}

	// Initialise citizens
	for _, citizen := range payload.Citizens {
		pos := citizen.CurrentPosition
		m.Grid[pos.Y][pos.X] = citizenCharacter
	}

	m.Citizens = payload.Citizens
}
