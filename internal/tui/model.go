// Package tui provides user interface for hazard simulation
package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/google/uuid"

	"hazard/internal/citizen"
	"hazard/internal/events"
	"hazard/internal/pathfinding"
)

type citizenState struct {
	Position     pathfinding.Position
	SafeZoneID   uuid.UUID
	PreviousCell string
	Status       citizen.Status
}

// Model represents the TUI state for the hazard simulation
type Model struct {
	simulationID        uuid.UUID
	grid                [][]string
	citizens            map[uuid.UUID]citizenState
	paths               map[uuid.UUID][]pathfinding.Position
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
	showPaths           bool
}

var sidebarWidth = 35

// InitialModel creates the initial TUI model state
func InitialModel(eventBus *events.EventBus) Model {
	inputs := InitialiseController(eventBus)
	focusTargets := 1 + len(inputs.inputs)

	return Model{
		grid:         [][]string{},
		citizens:     map[uuid.UUID]citizenState{},
		paths:        map[uuid.UUID][]pathfinding.Position{},
		hazards:      map[uuid.UUID]string{},
		safeZones:    map[uuid.UUID][]pathfinding.Position{},
		eventBus:     eventBus,
		focusIndex:   0,
		focusTargets: focusTargets,
		inputs:       inputs,
	}
}

func (m *Model) updateFocusTargets() {
	used := lineCount(logo)
	used += lineCount(m.renderCitizenStatus())
	used += lineCount(m.renderControls())
	used += lineCount(m.renderKey())
	inputView, _ := m.inputs.View()
	if used+lineCount(inputView) <= m.height {
		m.focusTargets = 1 + len(m.inputs.InputIDs)
	} else {
		m.focusTargets = 1
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
	pathOverlay := m.buildPathOverlay()

	for y := range m.grid {
		if y > 0 {
			grid.WriteString("\n")
		}
		for x := range m.grid[y] {
			pos := pathfinding.Position{X: x, Y: y}
			if pathChar, ok := pathOverlay[pos]; ok {
				grid.WriteString(pathChar)
			} else {
				grid.WriteString(m.grid[y][x])
			}
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

func (m Model) buildPathOverlay() map[pathfinding.Position]string {
	if !m.showPaths {
		return nil
	}

	overlay := make(map[pathfinding.Position]string)

	for id, citizenPath := range m.paths {
		state, ok := m.citizens[id]
		if !ok {
			continue
		}

		if state.Status != citizen.StatusIdle && state.Status != citizen.StatusNavigating {
			continue
		}

		startIndex := -1
		for i, pos := range citizenPath {
			if pos == state.Position {
				startIndex = i
				break
			}
		}
		if startIndex == -1 {
			continue
		}

		remaining := citizenPath[startIndex+1:]
		for i := range remaining {
			var from pathfinding.Position
			if i == 0 {
				from = state.Position
			} else {
				from = remaining[i-1]
			}
			overlay[remaining[i]] = pathDirection(remaining, i, from)
		}
	}

	return overlay
}

func pathDirection(path []pathfinding.Position, i int, from pathfinding.Position) string {
	if i < len(path)-1 {
		if path[i].Y == path[i+1].Y {
			return pathHorizontalCharacter
		}
		return pathVerticalCharacter
	}

	if path[i].Y == from.Y {
		return pathHorizontalCharacter
	}
	return pathVerticalCharacter
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
		m.updateFocusTargets()

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
		case events.CitizenPathUpdated:
			m.handleCitizenPathUpdated(event)
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
			m.paths = map[uuid.UUID][]pathfinding.Position{}
			m.showPaths = false
			m.updateFocusTargets()
			m.eventBus.InitialiseSimulation(events.InitialiseSimulationPayload{
				Width:  m.width,
				Height: m.height,
			})
			return m, nil

		case "p":
			m.showPaths = !m.showPaths
		}
	}

	if m.focusIndex > 0 {
		focusedID := m.inputs.InputIDs[m.focusIndex-1]
		cmd := m.inputs.Update(msg, focusedID)
		return m, cmd
	}

	return m, nil
}
