package tui

import (
	"hazard/internal/events"
	"log"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

var inputWidth = 20

type input struct {
	id       string
	label    string
	model    textinput.Model
	callback func()
}

// InputController manages the different inputs and their state
type InputController struct {
	InputIDs         []string
	prevFocusInputID string
	inputs           map[string]input
	eventBus         *events.EventBus
}

// InitialiseController creates the controller
func InitialiseController(eventBus *events.EventBus) InputController {
	c := InputController{
		InputIDs: []string{},
		inputs:   make(map[string]input),
		eventBus: eventBus,
	}
	c.createInput("tickerInterval", "ticker interval", "100", 4, func() {
		val, err := strconv.Atoi(c.inputs["tickerInterval"].model.Value())
		if err != nil {
			log.Printf("error parsing tickerInterval as int: %v", err)
		} else {
			c.eventBus.UpdateTickerInterval(val)
		}
	})
	c.createInput("hazardProbability", "hazard probability", "0.1", 4, func() {
		val, err := strconv.ParseFloat(c.inputs["hazardProbability"].model.Value(), 32)
		if err != nil {
			log.Printf("error parsing tickerInterval as int: %v", err)
		} else {
			c.eventBus.UpdateHazardProbability(float32(val))
		}
	})

	return c
}

// View builds a string comprised of the input views
func (m *InputController) View() (string, *tea.Cursor) {
	var b strings.Builder
	var c *tea.Cursor

	for i, id := range m.InputIDs {
		b.WriteString(m.inputs[id].label)
		b.WriteString(" ")

		inputModel := m.inputs[id].model
		b.WriteString(inputModel.View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}

		cursor := inputModel.Cursor()
		if inputModel.Focused() && cursor != nil {
			c = cursor
			c.Y += i
		}
	}

	return b.String(), c
}

// Update the inputs and focus input by ID
func (m *InputController) Update(msg tea.Msg, focusInputID string) tea.Cmd {

	if focusInputID != m.prevFocusInputID {
		_, ok := m.inputs[m.prevFocusInputID]
		if ok {
			input := m.inputs[m.prevFocusInputID]
			input.callback()
			input.model.Blur()
			m.inputs[m.prevFocusInputID] = input
		}

		_, ok = m.inputs[focusInputID]
		if ok {
			input := m.inputs[focusInputID]
			input.model.Focus()
			m.inputs[focusInputID] = input
			m.prevFocusInputID = focusInputID
		}
	}

	cmds := make([]tea.Cmd, len(m.inputs))

	for i, id := range m.InputIDs {
		var model textinput.Model
		model, cmds[i] = m.inputs[id].model.Update(msg)
		input := m.inputs[id]
		input.model = model
		m.inputs[id] = input
	}

	return tea.Batch(cmds...)
}

func (m *InputController) createInput(
	id string, label string, placeholder string, charLimit int, callback func()) {
	model := textinput.New()
	model.Placeholder = placeholder
	model.CharLimit = charLimit

	model.SetWidth(inputWidth)
	model.SetValue(placeholder)

	m.InputIDs = append(m.InputIDs, id)
	m.inputs[id] = input{
		id:       id,
		label:    label,
		model:    model,
		callback: callback,
	}
}
