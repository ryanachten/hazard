package common

import (
	"errors"
	"fmt"
)

// SimulationConfig configuration for a simulation
type SimulationConfig struct {
	TickIntervalMs    int
	CitizenCountRange PositiveRange
	Hazard            HazardConfig
	SafeZone          SafeZoneConfig
	Obstacle          ObstacleConfig
}

// HazardConfig configures simulation hazards
type HazardConfig struct {
	Probability   float32
	CountRange    Range
	DurationRange PositiveRange
}

// SafeZoneConfig configures simulation safe zones
type SafeZoneConfig struct {
	Probability float32
	CountRange  Range
	RadiusRange Range
}

// ObstacleConfig configures simulation obstacles
type ObstacleConfig struct {
	CountRange Range
	SizeRange  PositiveRange
}

// Validate ensures configuration is valid prior to use
func (s *SimulationConfig) Validate() error {
	var errs []error

	if err := ValidatePositiveRange(s.CitizenCountRange.Min, s.CitizenCountRange.Max); err != nil {
		errs = append(errs, fmt.Errorf("CitizenCountRange: %w", err))
	}
	if err := ValidateRange(s.Hazard.CountRange.Min, s.Hazard.CountRange.Max); err != nil {
		errs = append(errs, fmt.Errorf("Hazard.CountRange: %w", err))
	}
	if err := ValidatePositiveRange(s.Hazard.DurationRange.Min, s.Hazard.DurationRange.Max); err != nil {
		errs = append(errs, fmt.Errorf("Hazard.DurationRange: %w", err))
	}
	if err := ValidateRange(s.SafeZone.CountRange.Min, s.SafeZone.CountRange.Max); err != nil {
		errs = append(errs, fmt.Errorf("SafeZone.CountRange: %w", err))
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
	if s.SafeZone.CountRange.Max < 1 {
		errs = append(errs, errors.New("SafeZone.CountRange must be at least 1"))
	}

	return errors.Join(errs...)
}
