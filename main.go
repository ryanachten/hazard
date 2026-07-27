// Main entrypoint for app
package main

import (
	"context"
	cfg "hazard/internal/configuration"
	"hazard/internal/engine"
	"hazard/internal/events"
	"hazard/internal/tui"
	"log"
	"log/slog"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
)

var debugFilename = "debug.log"

func main() {
	config := cfg.DefaultConfig

	err := config.Validate()
	if err != nil {
		log.Fatalf("error validation config: %v", err)
	}

	eventBus := events.New()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var simulation engine.Simulation

	go func() {
		ticker := time.NewTicker(time.Duration(config.TickIntervalMs) * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if simulation.State == engine.SimulationRunning {
					simulation.Tick()
				}
			case cmd := <-eventBus.SimulationCommands:
				switch cmd.CommandType {
				case events.InitialiseSimulation:
					payload, ok := cmd.Payload.(events.InitialiseSimulationPayload)
					if !ok {
						slog.Error("error parsing initialisation payload", "payload", payload)
						continue
					}
					newSim, err := engine.NewSimulation(payload.Width, payload.Height, config, eventBus)
					if err != nil {
						slog.Error("error creating simulation", "err", err)
						continue
					}
					simulation = newSim

				case events.UpdateTickerInterval:
					newInterval, ok := cmd.Payload.(int)
					if !ok {
						slog.Error("error parsing ticker payload", "newInterval", newInterval)
						continue
					}
					ticker.Reset(time.Duration(newInterval) * time.Millisecond)
				default:
					simulation.ProcessCommand(cmd)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	_ = os.Remove(debugFilename)
	_, err = os.Create(debugFilename)
	if err != nil {
		log.Fatalf("error creating debug file: %v", err)
	}
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

	p := tea.NewProgram(tui.InitialModel(eventBus))
	if _, err := p.Run(); err != nil {
		log.Fatalf("error running program: %v", err)
	}

	cancel()
}
