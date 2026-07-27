// Package tui provides user interface for hazard simulation
package tui

import (
	"hazard/internal/events"
	"hazard/internal/pathfinding"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/google/uuid"
)

type citizenState struct {
	Position     pathfinding.Position
	SafeZoneID   uuid.UUID
	PreviousCell string
}

// Model represents the TUI state for the hazard simulation
type Model struct {
	simulationID        uuid.UUID
	grid                [][]string
	citizens            map[uuid.UUID]citizenState
	escapedCitizenCount int
	deadCitizenCount    int
	activeCitizenCount  int
	hazards             map[uuid.UUID]string
	safeZones           map[uuid.UUID][]pathfinding.Position
	eventBus            *events.EventBus
	width               int
	height              int
	focusIndex          int
	focusTargets        int
	inputs              InputController
	showSidebar         bool
}

var sidebarWidth = 35

// InitialModel creates the initial TUI model state
func InitialModel(eventBus *events.EventBus) Model {
	inputs := InitialiseController(eventBus)
	focusTargets := 1 + len(inputs.inputs)

	return Model{
		grid:         [][]string{},
		citizens:     map[uuid.UUID]citizenState{},
		hazards:      map[uuid.UUID]string{},
		safeZones:    map[uuid.UUID][]pathfinding.Position{},
		eventBus:     eventBus,
		focusIndex:   0,
		focusTargets: focusTargets,
		inputs:       inputs,
	}
}

func (m Model) consumeSimulationEvent() tea.Msg {
	event := <-m.eventBus.SimulationEvents
	return event
}

// Init initializes the Bubble Tea program
func (m Model) Init() tea.Cmd {
	return m.consumeSimulationEvent
}

// View renders the current simulation state
func (m Model) View() tea.View {
	var grid strings.Builder

	for y := range m.grid {
		if y > 0 {
			grid.WriteString("\n")
		}
		for x := range m.grid[y] {
			cell := m.grid[y][x]
			grid.WriteString(cell)
		}
	}

	styledGrid := gridStyle.SetString(grid.String()).Render()

	if m.showSidebar {
		rightColumn, cursor := m.renderSidebar()

		view := tea.NewView(lipgloss.JoinHorizontal(lipgloss.Top, styledGrid, rightColumn))
		view.Cursor = cursor
		return view
	}

	view := tea.NewView(styledGrid)
	return view
}

// Update handles messages and updates the TUI model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		if msg.Width > sidebarWidth*2 {
			m.width = msg.Width - sidebarWidth
			m.showSidebar = true
		} else {
			m.width = msg.Width
			m.showSidebar = false
		}

		m.height = msg.Height

		m.eventBus.InitialiseSimulation(events.InitialiseSimulationPayload{
			Width:  m.width,
			Height: m.height,
		})
		return m, nil

	case events.SimulationEvent:
		event := events.SimulationEvent(msg)

		// Skip processing any events where the simulation ID doesn't match the current simulation
		// this avoids processing stale events
		if event.EventType != events.SimulationCreated && event.SimulationID != m.simulationID {
			return m, m.consumeSimulationEvent
		}

		switch event.EventType {
		case events.SimulationCreated:
			m.handleSimulationCreated(event)
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

		return m, m.consumeSimulationEvent

	case tea.KeyPressMsg:
		switch msg.String() {

		case "tab", "down":
			m.focusIndex++
			if m.focusIndex >= m.focusTargets {
				m.focusIndex = 0
			}

		case "shift+tab", "up":
			m.focusIndex--
			if m.focusIndex < 0 {
				m.focusIndex = m.focusTargets - 1
			}

		case "ctrl+c", "q", "esc":
			return m, tea.Quit

		case "enter", "space":
			m.eventBus.PauseSimulation()
			return m, nil

		case "r":
			m.simulationID = uuid.Nil
			m.eventBus.InitialiseSimulation(events.InitialiseSimulationPayload{
				Width:  m.width,
				Height: m.height,
			})
			return m, nil
		}
	}

	if m.focusIndex > 0 {
		focusedID := m.inputs.InputIDs[m.focusIndex-1]
		cmd := m.inputs.Update(msg, focusedID)
		return m, cmd
	}

	return m, nil
}
