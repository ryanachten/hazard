package events

const eventBufferSize = 256

// EventBus for dispatching events and receiving commands
type EventBus struct {
	SimulationCommands chan SimulationCommand
	SimulationEvents   chan SimulationEvent // SimulationEventChannel to subscribe to simulation events
	EventLog           []SimulationEvent
}

// CreateEventBus instantiates an event bus
func CreateEventBus() *EventBus {
	return &EventBus{
		SimulationCommands: make(chan SimulationCommand, eventBufferSize),
		SimulationEvents:   make(chan SimulationEvent, eventBufferSize),
		EventLog:           []SimulationEvent{},
	}
}
