// Main entrypoint for app
package main

import (
	"context"
	c "hazard/internal/common"
	eng "hazard/internal/engine"
	"hazard/internal/events"
	"hazard/internal/tui"
	"log"
	"time"

	tea "charm.land/bubbletea/v2"
)

func main() {
	config := c.SimulationConfig{
		TickIntervalMs:    100,
		Height:            100,
		Width:             100,
		CitizenCountRange: [2]int{5, 10},
		Hazard: c.HazardConfig{
			Probability:   0.3,
			CountRange:    [2]int{2, 5},
			DurationRange: [2]int{5, 10},
		},
		SafeZone: c.SafeZoneConfig{
			Probability: 0.2,
			RadiusRange: [2]int{3, 7},
			CountRange:  [2]int{1, 3},
		},
	}

	err := config.Validate()
	if err != nil {
		log.Fatalf("error validation config: %v", err)
	}

	eventBus := events.CreateEventBus()

	simulation, err := eng.NewSimulation(config, eventBus)
	if err != nil {
		log.Fatalf("error creating simulation: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		ticker := time.NewTicker(time.Duration(config.TickIntervalMs) * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if simulation.State == eng.SimulationRunning {
					simulation.Tick()
				}
			case cmd := <-eventBus.SimulationCommands:
				if cmd.CommandType == events.RestartSimulation {
					newSim, err := eng.NewSimulation(config, eventBus)
					if err != nil {
						log.Printf("error restarting simulation: %v", err)
						continue
					}
					simulation = newSim
				} else {
					simulation.ProcessCommand(cmd)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	p := tea.NewProgram(tui.InitialModel(eventBus))
	if _, err := p.Run(); err != nil {
		log.Fatalf("error running program: %v", err)
	}

	cancel()
}
