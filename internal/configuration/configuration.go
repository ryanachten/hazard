// Package configuration defines the simulation configuration
package configuration

import (
	"errors"
	"fmt"
	"hazard/internal/bounds"
	"hazard/internal/hazard"
	"hazard/internal/obstacle"
	"hazard/internal/safezone"
)

// SimulationConfig configuration for a simulation
type SimulationConfig struct {
	TickIntervalMs int
	CitizenCount   int
	Hazard         hazard.Config
	SafeZone       safezone.Config
	Obstacle       obstacle.Config
}

// DefaultConfig for the simulation
var DefaultConfig = SimulationConfig{
	TickIntervalMs: 100,
	CitizenCount:   30,
	Hazard: hazard.Config{
		Probability:   0.1,
		Count:         4,
		DurationRange: bounds.PositiveRange{Min: 5, Max: 10},
	},
	SafeZone: safezone.Config{
		Probability: 0.06,
		Count:       3,
		RadiusRange: bounds.Range{Min: 1, Max: 1},
	},
	Obstacle: obstacle.Config{
		CountRange: bounds.Range{Min: 3, Max: 20},
		SizeRange:  bounds.PositiveRange{Min: 1, Max: 3},
	},
}

// Validate ensures configuration is valid prior to use
func (s *SimulationConfig) Validate() error {
	var errs []error

	if s.CitizenCount < 1 {
		errs = append(errs, errors.New("citizen count must be at least 1"))
	}
	if s.Hazard.Count < 1 {
		errs = append(errs, errors.New("hazard count must be at least 1"))
	}
	if s.SafeZone.Count < 1 {
		errs = append(errs, errors.New("safe zone count must be at least 1"))
	}
	if err := bounds.ValidatePositive(s.Hazard.DurationRange.Min, s.Hazard.DurationRange.Max); err != nil {
		errs = append(errs, fmt.Errorf("hazard duration range: %w", err))
	}
	if err := bounds.Validate(s.SafeZone.RadiusRange.Min, s.SafeZone.RadiusRange.Max); err != nil {
		errs = append(errs, fmt.Errorf("safe zone radius range: %w", err))
	}
	if err := bounds.Validate(s.Obstacle.CountRange.Min, s.Obstacle.CountRange.Max); err != nil {
		errs = append(errs, fmt.Errorf("obstacle count range: %w", err))
	}
	if err := bounds.ValidatePositive(s.Obstacle.SizeRange.Min, s.Obstacle.SizeRange.Max); err != nil {
		errs = append(errs, fmt.Errorf("obstacle size range: %w", err))
	}
	if s.Hazard.Probability < 0 || s.Hazard.Probability > 1 {
		errs = append(errs, errors.New("hazard probability must be between 0.0 and 1.0"))
	}
	if s.SafeZone.Probability < 0 || s.SafeZone.Probability > 1 {
		errs = append(errs, errors.New("safe zone probability must be between 0.0 and 1.0"))
	}

	return errors.Join(errs...)
}
