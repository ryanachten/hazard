// Main entrypoint for app
package main

import (
	eng "hazard/internal/engine"
	"hazard/internal/events"
	"log"
	"time"
)

func main() {
	config := eng.SimulationConfig{
		TickIntervalMs:    100,
		Height:            100,
		Width:             100,
		CitizenCountRange: [2]int{5, 10},
	}

	err := config.Validate()
	if err != nil {
		log.Fatalf("error validation config: %v", err)
	}

	simulation, err := eng.NewSimulation(config)
	if err != nil {
		log.Fatalf("error creating simulation: %v", err)
	}

	go func() {
		err := simulation.EventEmitter.SimulationStarted(events.EventMetadata{
			SimulationID: simulation.ID,
			Tick:         simulation.TickCount,
		})
		if err != nil {
			log.Fatalf("error creating SimulationStarted event: %v", err)
		}

		ticker := time.NewTicker(time.Duration(config.TickIntervalMs) * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			simulation.Tick()
		}
	}()

	select {}
}
