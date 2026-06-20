// Main entrypoint for app
package main

import (
	"fmt"
	eng "hazard/internal/engine"
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
		msg := fmt.Sprintf("simulation config invalid: %v", err)
		panic(msg)
	}

	simulation, err := eng.NewSimulation(config)
	if err != nil {
		msg := fmt.Sprintf("error creating simulation: %v", err)
		panic(msg)
	}

	go func() {
		ticker := time.NewTicker(time.Duration(config.TickIntervalMs) * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			simulation.Tick()
		}
	}()

	select {}
}
