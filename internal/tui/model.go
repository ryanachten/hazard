// Package tui provides user interface for hazard simulation
package tui

import (
	e "hazard/internal/events"
	pf "hazard/internal/pathfinding"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/google/uuid"
)

// Model represents the TUI state for the hazard simulation
type Model struct {
	simulationID uuid.UUID
	grid         [][]string
	citizens     map[uuid.UUID]pf.Position
	hazards      map[uuid.UUID]string
	safeZones    map[pf.Position]string
	eventBus     *e.EventBus
}

// InitialModel creates the initial TUI model state
func InitialModel(eventBus *e.EventBus) Model {
	return Model{
		grid:      [][]string{},
		citizens:  map[uuid.UUID]pf.Position{},
		hazards:   map[uuid.UUID]string{},
		safeZones: map[pf.Position]string{},
		eventBus:  eventBus,
	}
}

func (m Model) consumeEvent() tea.Msg {
	event := <-m.eventBus.SimulationEvents
	return event
}

func (m Model) dispatchEvent(event e.SimulationCommand) tea.Cmd {
	m.eventBus.SimulationCommands <- event
	return nil
}

// Init initializes the Bubble Tea program
func (m Model) Init() tea.Cmd {
	return m.consumeEvent
}

// View renders the current simulation state
func (m Model) View() tea.View {
	var s strings.Builder

	s.WriteString("Running!\n")

	for y := range m.grid {
		for x := range m.grid[y] {
			cell := m.grid[y][x]
			s.WriteString(cell)
		}
		s.WriteString("\n")
	}

	return tea.NewView(s.String())
}

// Update handles messages and updates the TUI model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	case e.SimulationEvent:
		event := e.SimulationEvent(msg)

		// Skip processing any events where the simulation ID doesn't match the current simulation
		// this avoids processing stale events
		if event.EventType != e.SimulationCreated && event.SimulationID != m.simulationID {
			return m, m.consumeEvent
		}

		switch event.EventType {
		case e.SimulationCreated:
			m.handleSimulationCreated(event)
		case e.SimulationCompleted:
			return m, tea.Quit
		case e.CitizenMoved:
			m.handleCitizenMoved(event)
		case e.CitizenDied:
			m.handleCitizenDied(event)
		case e.SafeZoneEmerged:
			m.handleSafeZoneEmerged(event)
		case e.CitizenEscaped:
			m.handleCitizenEscaped(event)
		case e.HazardEmerged:
			m.handleHazardEmerged(event)
		case e.HazardExpanded:
			m.handleHazardExpanded(event)
		case e.HazardDissipated:
			m.handleHazardDissipated(event)
		}

		return m, m.consumeEvent

	case tea.KeyPressMsg:

		switch msg.String() {

		case "ctrl+c", "q", "esc":
			return m, tea.Quit

		case "enter", "space":
			return m, m.dispatchEvent(e.SimulationCommand{
				CommandType: e.PauseSimulation,
			})

		case "r":
			m.simulationID = uuid.Nil
			return m, m.dispatchEvent(e.SimulationCommand{
				CommandType: e.RestartSimulation,
			})
		}
	}

	return m, nil
}
