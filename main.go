// Main entrypoint for app
package main

import (
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
	simulation := eng.NewSimulation(config)

	go func() {
		ticker := time.NewTicker(time.Duration(config.TickIntervalMs) * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			simulation.Tick()
		}
	}()

	select {}
}
