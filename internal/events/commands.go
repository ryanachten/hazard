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
	// UpdateTickerInterval updates the ticker interval in milliseconds
	UpdateTickerInterval commandType = "simulation.updateTickerInterval"
	// UpdateHazardProbability updates the probability of hazards occurring
	UpdateHazardProbability commandType = "simulation.updateHazardProbability"
	// UpdateHazardCount updates the number of hazards which can exist
	UpdateHazardCount commandType = "simulation.updateHazardCount"
	// UpdateHazardDurationMin updates the min duration of a hazard
	UpdateHazardDurationMin commandType = "simulation.updateHazardDurationMin"
	// UpdateHazardDurationMax updates the max duration of a hazard
	UpdateHazardDurationMax commandType = "simulation.updateHazardDurationMax"
	// UpdateSafeZoneProbability updates the probability of safe zones occurring
	UpdateSafeZoneProbability commandType = "simulation.updateSafeZoneProbability"
	// UpdateSafeZoneCount updates the number of safe zones which can exist
	UpdateSafeZoneCount commandType = "simulation.UpdateSafeZoneCount"
	// UpdateSafeZoneRadiusMin updates the min radius of a safe zone
	UpdateSafeZoneRadiusMin commandType = "simulation.updateSafeZoneRadiusMin"
	// UpdateSafeZoneRadiusMax updates the max radius of a safe zone
	UpdateSafeZoneRadiusMax commandType = "simulation.updateSafeZoneRadiusMax"
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

// UpdateTickerInterval updates the ticker interval in milliseconds
func (e *EventBus) UpdateTickerInterval(intervalMs int) {
	e.dispatchCommand(SimulationCommand{
		CommandType: UpdateTickerInterval,
		Payload:     intervalMs,
	})
}

// UpdateHazardProbability updates the probability of hazards occurring
func (e *EventBus) UpdateHazardProbability(pct float32) {
	e.dispatchCommand(SimulationCommand{
		CommandType: UpdateHazardProbability,
		Payload:     pct,
	})
}

// UpdateHazardCount updates the number of hazards which can exist
func (e *EventBus) UpdateHazardCount(count int) {
	e.dispatchCommand(SimulationCommand{
		CommandType: UpdateHazardCount,
		Payload:     count,
	})
}

// UpdateHazardDurationMax updates the max duration a hazard can exist for
func (e *EventBus) UpdateHazardDurationMax(duration int) {
	e.dispatchCommand(SimulationCommand{
		CommandType: UpdateHazardDurationMax,
		Payload:     duration,
	})
}

// UpdateHazardDurationMin updates the min duration a hazard can exist for
func (e *EventBus) UpdateHazardDurationMin(duration int) {
	e.dispatchCommand(SimulationCommand{
		CommandType: UpdateHazardDurationMin,
		Payload:     duration,
	})
}

// UpdateSafeZoneProbability updates the probability of hazards occurring
func (e *EventBus) UpdateSafeZoneProbability(pct float32) {
	e.dispatchCommand(SimulationCommand{
		CommandType: UpdateSafeZoneProbability,
		Payload:     pct,
	})
}

// UpdateSafeZoneCount updates the number of safe zones which can exist
func (e *EventBus) UpdateSafeZoneCount(count int) {
	e.dispatchCommand(SimulationCommand{
		CommandType: UpdateSafeZoneCount,
		Payload:     count,
	})
}

// UpdateSafeZoneRadiusMin updates the min radius for a safe zone
func (e *EventBus) UpdateSafeZoneRadiusMin(radius int) {
	e.dispatchCommand(SimulationCommand{
		CommandType: UpdateSafeZoneRadiusMin,
		Payload:     radius,
	})
}

// UpdateSafeZoneRadiusMax updates the max radius for a safe zone
func (e *EventBus) UpdateSafeZoneRadiusMax(radius int) {
	e.dispatchCommand(SimulationCommand{
		CommandType: UpdateSafeZoneRadiusMax,
		Payload:     radius,
	})
}

func (e *EventBus) dispatchCommand(cmd SimulationCommand) {
	e.SimulationCommands <- cmd
}
