// Main entrypoint for app
package main

import (
	"context"
	c "hazard/internal/common"
	eng "hazard/internal/engine"
	"hazard/internal/events"
	"hazard/internal/tui"
	"log"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
)

var debugFilename = "debug.log"

func main() {
	config := c.SimulationConfig{
		TickIntervalMs:    100,
		Height:            100,
		Width:             100,
		CitizenCountRange: c.PositiveRange{Min: 25, Max: 40},
		Hazard: c.HazardConfig{
			Probability:   0.1,
			CountRange:    c.Range{Min: 2, Max: 5},
			DurationRange: c.PositiveRange{Min: 5, Max: 10},
		},
		SafeZone: c.SafeZoneConfig{
			Probability: 0.06,
			RadiusRange: c.Range{Min: 1, Max: 1},
			CountRange:  c.Range{Min: 1, Max: 3},
		},
		Obstacle: c.ObstacleConfig{
			CountRange: c.Range{Min: 3, Max: 20},
			SizeRange:  c.PositiveRange{Min: 1, Max: 3},
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

	if len(os.Getenv("DEBUG")) > 0 {
		logToDebugFile()
	}

	p := tea.NewProgram(tui.InitialModel(eventBus))
	if _, err := p.Run(); err != nil {
		log.Fatalf("error running program: %v", err)
	}

	cancel()
}

func logToDebugFile() {
	_ = os.Remove(debugFilename)
	f, err := tea.LogToFile(debugFilename, "")
	if err != nil {
		log.Fatalf("error logging to debug file: %v", err)
	}
	log.SetOutput(f)
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("error closing file: %v", err)

		}
	}()
}
