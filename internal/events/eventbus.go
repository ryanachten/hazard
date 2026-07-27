package events

const eventBufferSize = 256

// EventBus for dispatching events and receiving commands
type EventBus struct {
	SimulationCommands chan SimulationCommand
	SimulationEvents   chan SimulationEvent
	EventLog           []SimulationEvent
}

// New instantiates an event bus
func New() *EventBus {
	return &EventBus{
		SimulationCommands: make(chan SimulationCommand, eventBufferSize),
		SimulationEvents:   make(chan SimulationEvent, eventBufferSize),
	}
}
