// Package tui provides user interface for hazard simulation
package tui

import (
	e "hazard/internal/events"
	pf "hazard/internal/pathfinding"
	"log"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/textinput"
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
	simulationID        uuid.UUID
	grid                [][]string
	citizens            map[uuid.UUID]citizenState
	hazards             map[uuid.UUID]string
	safeZones           map[uuid.UUID][]pf.Position
	eventBus            *e.EventBus
	width               int
	height              int
	focusIndex          int
	focusTargets        int // TODO: this will eventually be driven by an array or something
	tickerIntervalInput textinput.Model
}

var sidebarWidth = 40

// InitialModel creates the initial TUI model state
func InitialModel(eventBus *e.EventBus) Model {
	tickerIntervalInput := textinput.New()
	tickerIntervalInput.Placeholder = "100"
	tickerIntervalInput.Prompt = "Frame rate interval"
	tickerIntervalInput.CharLimit = 4
	tickerIntervalInput.SetWidth(20)

	return Model{
		grid:                [][]string{},
		citizens:            map[uuid.UUID]citizenState{},
		hazards:             map[uuid.UUID]string{},
		safeZones:           map[uuid.UUID][]pf.Position{},
		eventBus:            eventBus,
		focusIndex:          0,
		focusTargets:        2,
		tickerIntervalInput: tickerIntervalInput,
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

	cursor := m.tickerIntervalInput.Cursor()

	view := tea.NewView(lipgloss.JoinHorizontal(lipgloss.Top, styledGrid, m.tickerIntervalInput.View()))
	view.Cursor = cursor
	return view
}

// Update handles messages and updates the TUI model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	// TODO: handle resize events
	case tea.WindowSizeMsg:
		m.width = msg.Width - sidebarWidth
		m.height = msg.Height

		m.eventBus.InitialiseSimulation(e.InitialiseSimulationPayload{
			Width:  m.width,
			Height: m.height,
		})
		return m, nil

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

		case "tab":
			prevFocus := m.focusIndex
			m.focusIndex++
			if m.focusIndex >= m.focusTargets {
				m.focusIndex = 0
			}
			switch m.focusIndex {
			case 1:
				m.tickerIntervalInput.Focus()
			}
			switch prevFocus {
			case 1:
				val, err := strconv.Atoi(m.tickerIntervalInput.Value())
				if err != nil {
					log.Printf("error parsing input value: %v", err)
				}
				m.eventBus.UpdateTickerInterval(val)
				m.tickerIntervalInput.Blur()
			}

		case "ctrl+c", "q", "esc":
			return m, tea.Quit

		case "enter", "space":
			m.eventBus.PauseSimulation()
			return m, nil

		case "r":
			m.simulationID = uuid.Nil
			m.eventBus.InitialiseSimulation(e.InitialiseSimulationPayload{
				Width:  m.width,
				Height: m.height,
			})
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.tickerIntervalInput, cmd = m.tickerIntervalInput.Update(msg)

	return m, cmd
}
