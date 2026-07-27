// Program hazard is a 2D grid-based hazard simulation
package main

import (
	"context"
	"hazard/internal/configuration"
	"hazard/internal/engine"
	"hazard/internal/events"
	"hazard/internal/tui"
	"log/slog"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
)

var debugFilename = "debug.log"

func main() {
	// Validate configuration
	config := configuration.DefaultConfig
	err := config.Validate()
	if err != nil {
		slog.Error("error validation config", "err", err)
		os.Exit(1)
	}

	var simulation engine.Simulation
	eventBus := events.New()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	// Log to debug file rather than stdout to avoid disrupting TUI
	_ = os.Remove(debugFilename)
	_, err = os.Create(debugFilename)
	if err != nil {
		slog.Error("error creating debug file", "err", err)
		os.Exit(1)
	}

	f, err := tea.LogToFile(debugFilename, "")
	if err != nil {
		slog.Error("error logging to debug file", "err", err)
		os.Exit(1)
	}

	handler := slog.NewTextHandler(f, nil)
	slog.SetDefault(slog.New(handler))

	defer func() {
		if err := f.Close(); err != nil {
			slog.Error("error closing debug file", "err", err)
		}
	}()

	// Run bubbletea TUI
	p := tea.NewProgram(tui.InitialModel(eventBus))
	if _, err := p.Run(); err != nil {
		slog.Error("error running program", "err", err)
		os.Exit(1)
	}

	cancel()
}
