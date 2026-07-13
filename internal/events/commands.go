package events

type commandType string

// SimulationCommand represents a command sent to the simulation for processing.
type SimulationCommand struct {
	CommandType commandType
	Payload     any
}

const (
	// InitialiseSimulation passes information required to initialise a simulation
	InitialiseSimulation commandType = "simulation.initialise"
	// PauseSimulation toggles the simulation between running and paused states.
	PauseSimulation commandType = "simulation.pause"
	// UpdateTickerInterval updates the ticket interval in milliseconds
	UpdateTickerInterval commandType = "simulation.updateTicketInterval"
)

// InitialiseSimulationPayload passes information required to initialise a simulation
type InitialiseSimulationPayload struct {
	Width  int
	Height int
}

// InitialiseSimulation dispatches necessary information to create a simulation
func (e *EventBus) InitialiseSimulation(payload InitialiseSimulationPayload) {
	e.dispatchCommand(SimulationCommand{
		CommandType: InitialiseSimulation,
		Payload:     payload,
	})
}

// PauseSimulation toggles the simulation between running and paused states.
func (e *EventBus) PauseSimulation() {
	e.dispatchCommand(SimulationCommand{
		CommandType: PauseSimulation,
	})
}

// UpdateTickerInterval updates the ticket interval in milliseconds
func (e *EventBus) UpdateTickerInterval(intervalMs int) {
	e.dispatchCommand(SimulationCommand{
		CommandType: UpdateTickerInterval,
		Payload:     intervalMs,
	})
}

func (e *EventBus) dispatchCommand(cmd SimulationCommand) {
	e.SimulationCommands <- cmd
}
