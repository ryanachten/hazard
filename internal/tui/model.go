// Package tui provides user interface for hazard simulation
package tui

import (
	e "hazard/internal/events"
	pf "hazard/internal/pathfinding"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/google/uuid"
)

type citizenState struct {
	Position     pf.Position
	SafeZoneID   uuid.UUID
	PreviousCell string
}

// Model represents the TUI state for the hazard simulation
type Model struct {
	simulationID uuid.UUID
	grid         [][]string
	citizens     map[uuid.UUID]citizenState
	hazards      map[uuid.UUID]string
	safeZones    map[uuid.UUID][]pf.Position
	eventBus     *e.EventBus
	logs         []string
}

// InitialModel creates the initial TUI model state
func InitialModel(eventBus *e.EventBus) Model {
	return Model{
		grid:      [][]string{},
		citizens:  map[uuid.UUID]citizenState{},
		hazards:   map[uuid.UUID]string{},
		safeZones: map[uuid.UUID][]pf.Position{},
		eventBus:  eventBus,
		logs:      []string{},
	}
}

func (m Model) consumeSimulationEvent() tea.Msg {
	event := <-m.eventBus.SimulationEvents
	return event
}

func (m Model) consumeLogEvent() tea.Msg {
	log := <-m.eventBus.SystemLogs
	return log
}

func (m Model) dispatchSimulationCommand(event e.SimulationCommand) tea.Cmd {
	m.eventBus.SimulationCommands <- event
	return nil
}

// Init initializes the Bubble Tea program
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.consumeSimulationEvent, m.consumeLogEvent)
}

// View renders the current simulation state
func (m Model) View() tea.View {
	var grid strings.Builder

	for y := range m.grid {
		for x := range m.grid[y] {
			cell := m.grid[y][x]
			grid.WriteString(cell)
		}
		grid.WriteString("\n")
	}

	var logOutput strings.Builder

	for _, log := range m.logs {
		logOutput.WriteString(log)
		logOutput.WriteString("\n")
	}

	styledLog := lipgloss.Wrap(logOutput.String(), 40, " ")
	styledGrid := gridStyle.SetString(grid.String()).Render()

	return tea.NewView(lipgloss.JoinHorizontal(lipgloss.Bottom, styledGrid, styledLog))
}

// Update handles messages and updates the TUI model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	case string:
		m.logs = append(m.logs, string(msg))

	case e.SimulationEvent:
		event := e.SimulationEvent(msg)

		// Skip processing any events where the simulation ID doesn't match the current simulation
		// this avoids processing stale events
		if event.EventType != e.SimulationCreated && event.SimulationID != m.simulationID {
			return m, m.consumeSimulationEvent
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

		return m, m.consumeSimulationEvent

	case tea.KeyPressMsg:

		switch msg.String() {

		case "ctrl+c", "q", "esc":
			return m, tea.Quit

		case "enter", "space":
			return m, m.dispatchSimulationCommand(e.SimulationCommand{
				CommandType: e.PauseSimulation,
			})

		case "r":
			m.simulationID = uuid.Nil
			return m, m.dispatchSimulationCommand(e.SimulationCommand{
				CommandType: e.RestartSimulation,
			})
		}
	}

	return m, nil
}
