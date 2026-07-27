package tui

import (
	"hazard/internal/configuration"
	"hazard/internal/events"
	"log/slog"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

var inputWidth = 20
var config = configuration.DefaultConfig

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

// InitialiseController creates the controller and its inputs
func InitialiseController(eventBus *events.EventBus) InputController {
	c := InputController{
		InputIDs: []string{},
		inputs:   make(map[string]input),
		eventBus: eventBus,
	}

	c.createIntInput("tickerInterval", "Ticker interval", config.TickIntervalMs, c.eventBus.UpdateTickerInterval)
	c.createFloat32Input("hazardProbability", "Hazard probability", config.Hazard.Probability, c.eventBus.UpdateHazardProbability)
	c.createIntInput("hazardCount", "Hazard count", config.Hazard.Count, c.eventBus.UpdateHazardCount)
	c.createIntInput("hazardDurationMin", "Hazard duration min", config.Hazard.DurationRange.Min, c.eventBus.UpdateHazardDurationMin)
	c.createIntInput("hazardDurationMax", "Hazard duration max", config.Hazard.DurationRange.Max, c.eventBus.UpdateHazardDurationMax)
	c.createFloat32Input("safeZoneProbability", "Safe zone probability", config.SafeZone.Probability, c.eventBus.UpdateSafeZoneProbability)
	c.createIntInput("safeZoneCount", "Safe zone count", config.SafeZone.Count, c.eventBus.UpdateSafeZoneCount)
	c.createIntInput("safeZoneRadiusMin", "Safe zone radius min", config.SafeZone.RadiusRange.Min, c.eventBus.UpdateSafeZoneRadiusMin)
	c.createIntInput("safeZoneRadiusMax", "Safe zone radius max", config.SafeZone.RadiusRange.Max, c.eventBus.UpdateSafeZoneRadiusMax)

	return c
}

// View builds a string comprised of the input views
func (ic *InputController) View() (string, *tea.Cursor) {
	var b strings.Builder
	var c *tea.Cursor

	b.WriteString(heading.SetString("Settings").Render())

	for i, id := range ic.InputIDs {
		b.WriteString(ic.inputs[id].label)
		b.WriteString(" ")

		inputModel := ic.inputs[id].model
		b.WriteString(inputModel.View())
		b.WriteRune('\n')

		cursor := inputModel.Cursor()
		if inputModel.Focused() && cursor != nil {
			c = cursor
			c.Y += i
		}
	}

	return b.String(), c
}

// Update the inputs and focus input by ID
func (ic *InputController) Update(msg tea.Msg, focusInputID string) tea.Cmd {
	if focusInputID != ic.prevFocusInputID {
		_, ok := ic.inputs[ic.prevFocusInputID]
		if ok {
			input := ic.inputs[ic.prevFocusInputID]
			input.callback()
			input.model.Blur()
			ic.inputs[ic.prevFocusInputID] = input
		}

		_, ok = ic.inputs[focusInputID]
		if ok {
			input := ic.inputs[focusInputID]
			input.model.Focus()
			ic.inputs[focusInputID] = input
			ic.prevFocusInputID = focusInputID
		}
	}

	cmds := make([]tea.Cmd, len(ic.inputs))

	for i, id := range ic.InputIDs {
		var model textinput.Model
		model, cmds[i] = ic.inputs[id].model.Update(msg)
		input := ic.inputs[id]
		input.model = model
		ic.inputs[id] = input
	}

	return tea.Batch(cmds...)
}

func (ic *InputController) createFloat32Input(
	id string, label string, placeholder float32, callback func(state float32)) {

	formattedPlaceholder := strconv.FormatFloat(float64(placeholder), 'f', 2, 32)

	ic.createInput(id, label, formattedPlaceholder, func() {
		val, err := strconv.ParseFloat(ic.inputs[id].model.Value(), 32)
		if err != nil {
			slog.Error("error parsing input value as float", "inputId", id, "err", err)
		} else {
			callback(float32(val))
		}
	})
}

func (ic *InputController) createIntInput(
	id string, label string, placeholder int, callback func(state int)) {

	formattedPlaceholder := strconv.Itoa(placeholder)

	ic.createInput(id, label, formattedPlaceholder, func() {
		val, err := strconv.Atoi(ic.inputs[id].model.Value())
		if err != nil {
			slog.Error("error parsing input value as int", "inputId", id, "err", err)
		} else {
			callback(val)
		}
	})
}

func (ic *InputController) createInput(
	id string, label string, placeholder string, callback func()) {
	model := textinput.New()
	model.Placeholder = placeholder
	model.CharLimit = 4

	model.SetWidth(inputWidth)
	model.SetValue(placeholder)

	ic.InputIDs = append(ic.InputIDs, id)
	ic.inputs[id] = input{
		id:       id,
		label:    label,
		model:    model,
		callback: callback,
	}
}
