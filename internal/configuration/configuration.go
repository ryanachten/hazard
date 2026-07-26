// Package configuration defines the simulation configuration
package configuration

import (
	"errors"
	"fmt"
	h "hazard/internal/hazard"
	r "hazard/internal/numrange"
	o "hazard/internal/obstacle"
	sz "hazard/internal/safe_zone"
)

// SimulationConfig configuration for a simulation
type SimulationConfig struct {
	TickIntervalMs int
	CitizenCount   int
	Hazard         h.Config
	SafeZone       sz.Config
	Obstacle       o.Config
}

// DefaultConfig for the simulation
var DefaultConfig = SimulationConfig{
	TickIntervalMs: 100,
	CitizenCount:   30,
	Hazard: h.Config{
		Probability:   0.1,
		Count:         4,
		DurationRange: r.PositiveRange{Min: 5, Max: 10},
	},
	SafeZone: sz.Config{
		Probability: 0.06,
		Count:       3,
		RadiusRange: r.Range{Min: 1, Max: 1},
	},
	Obstacle: o.Config{
		CountRange: r.Range{Min: 3, Max: 20},
		SizeRange:  r.PositiveRange{Min: 1, Max: 3},
	},
}

// Validate ensures configuration is valid prior to use
func (s *SimulationConfig) Validate() error {
	var errs []error

	if s.CitizenCount < 1 {
		errs = append(errs, errors.New("CitizenCount must be at least 1"))
	}
	if s.Hazard.Count < 1 {
		errs = append(errs, errors.New("Hazard.Count must be at least 1"))
	}
	if s.SafeZone.Count < 1 {
		errs = append(errs, errors.New("SafeZone.Count must be at least 1"))
	}
	if err := r.ValidatePositive(s.Hazard.DurationRange.Min, s.Hazard.DurationRange.Max); err != nil {
		errs = append(errs, fmt.Errorf("Hazard.DurationRange: %w", err))
	}
	if err := r.Validate(s.SafeZone.RadiusRange.Min, s.SafeZone.RadiusRange.Max); err != nil {
		errs = append(errs, fmt.Errorf("SafeZone.RadiusRange: %w", err))
	}
	if err := r.Validate(s.Obstacle.CountRange.Min, s.Obstacle.CountRange.Max); err != nil {
		errs = append(errs, fmt.Errorf("Obstacle.CountRange: %w", err))
	}
	if err := r.ValidatePositive(s.Obstacle.SizeRange.Min, s.Obstacle.SizeRange.Max); err != nil {
		errs = append(errs, fmt.Errorf("Obstacle.SizeRange: %w", err))
	}
	if s.Hazard.Probability < 0 || s.Hazard.Probability > 1 {
		errs = append(errs, errors.New("Hazard.Probability must be between 0.0 and 1.0"))
	}
	if s.SafeZone.Probability < 0 || s.SafeZone.Probability > 1 {
		errs = append(errs, errors.New("SafeZone.Probability must be between 0.0 and 1.0"))
	}

	return errors.Join(errs...)
}
