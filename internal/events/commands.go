package events

type commandType string

const (
	// PauseSimulation toggles the simulation between running and paused states.
	PauseSimulation commandType = "simulation.pause"
)

// SimulationCommand represents a command sent to the simulation for processing.
type SimulationCommand struct {
	CommandType commandType
	Payload     any
}
