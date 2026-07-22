package common

import (
	"errors"
	"fmt"
)

// SimulationConfig configuration for a simulation
type SimulationConfig struct {
	TickIntervalMs int
	CitizenCount   int
	Hazard         HazardConfig
	SafeZone       SafeZoneConfig
	Obstacle       ObstacleConfig
}

// HazardConfig configures simulation hazards
type HazardConfig struct {
	Probability   float32
	Count         int
	DurationRange PositiveRange
}

// SafeZoneConfig configures simulation safe zones
type SafeZoneConfig struct {
	Probability float32
	Count       int
	RadiusRange Range
}

// ObstacleConfig configures simulation obstacles
type ObstacleConfig struct {
	CountRange Range
	SizeRange  PositiveRange
}

// DefaultConfig for the simulation
var DefaultConfig = SimulationConfig{
	TickIntervalMs: 100,
	CitizenCount:   30,
	Hazard: HazardConfig{
		Probability:   0.1,
		Count:         4,
		DurationRange: PositiveRange{Min: 5, Max: 10},
	},
	SafeZone: SafeZoneConfig{
		Probability: 0.06,
		Count:       3,
		RadiusRange: Range{Min: 1, Max: 1},
	},
	Obstacle: ObstacleConfig{
		CountRange: Range{Min: 3, Max: 20},
		SizeRange:  PositiveRange{Min: 1, Max: 3},
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
	if err := ValidatePositiveRange(s.Hazard.DurationRange.Min, s.Hazard.DurationRange.Max); err != nil {
		errs = append(errs, fmt.Errorf("Hazard.DurationRange: %w", err))
	}
	if err := ValidateRange(s.SafeZone.RadiusRange.Min, s.SafeZone.RadiusRange.Max); err != nil {
		errs = append(errs, fmt.Errorf("SafeZone.RadiusRange: %w", err))
	}
	if err := ValidateRange(s.Obstacle.CountRange.Min, s.Obstacle.CountRange.Max); err != nil {
		errs = append(errs, fmt.Errorf("Obstacle.CountRange: %w", err))
	}
	if err := ValidatePositiveRange(s.Obstacle.SizeRange.Min, s.Obstacle.SizeRange.Max); err != nil {
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
