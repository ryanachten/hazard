package engine

import (
	"errors"
	"math/rand"
)

// SimulationConfig configuration for a simulation
type SimulationConfig struct {
	TickIntervalMs    int
	Width             int
	Height            int
	CitizenCountRange [2]int
	Hazard            HazardConfig
	SafeZone          SafeZoneConfig
}

// HazardConfig configures simulation hazards
type HazardConfig struct {
	Probability   float32
	CountRange    [2]int
	DurationRange [2]int
}

// SafeZoneConfig configures simulation safe zones
type SafeZoneConfig struct {
	Probability float32
	CountRange  [2]int
	RadiusRange [2]int
}

// Validate ensures configuration is valid prior to use
func (s *SimulationConfig) Validate() error {
	var err []error

	if s.Width <= 0 || s.Height <= 0 {
		err = append(err, errors.New("simulation width and height must be greater than zero"))
	}

	if s.CitizenCountRange[0] > s.CitizenCountRange[1] {
		err = append(err, errors.New("CitizenCountRange[0] must be less than or equal to CitizenCountRange[1]"))
	}

	if s.Hazard.DurationRange[0] > s.Hazard.DurationRange[1] {
		err = append(err, errors.New("Hazard.HazardDurationRange[0] must be less than or equal to Hazard.HazardDurationRange[1]"))
	}

	if s.Hazard.DurationRange[0] <= 0 {
		err = append(err, errors.New("Hazard.HazardDurationRange values must be above 0"))
	}

	if s.Hazard.CountRange[0] > s.Hazard.CountRange[1] {
		err = append(err, errors.New("Hazard.CountRange[0] must be less than or equal to Hazard.CountRange[1]"))
	}

	if s.Hazard.Probability < 0 || s.Hazard.Probability > 1 {
		err = append(err, errors.New("Hazard.HazardProbability must be between 0.0 and 1.0"))
	}

	if s.SafeZone.Probability < 0 || s.SafeZone.Probability > 1 {
		err = append(err, errors.New("SafeZone.Probability must be between 0.0 and 1.0"))
	}

	if s.SafeZone.RadiusRange[0] > s.SafeZone.RadiusRange[1] {
		err = append(err, errors.New("SafeZone.RadiusRange[0] must be less than or equal to SafeZone.RadiusRange[1]"))
	}

	if s.SafeZone.CountRange[0] > s.SafeZone.CountRange[1] {
		err = append(err, errors.New("SafeZone.CountRange[0] must be less than or equal to SafeZone.CountRange[1]"))
	}

	if s.SafeZone.CountRange[1] < 1 {
		err = append(err, errors.New("SafeZone.CountRange[1] must be at least 1"))
	}

	return errors.Join(err...)
}

func randIntInRange(valueRange [2]int) int {
	configMin := valueRange[0]
	configMax := valueRange[1]
	return configMin + rand.Intn(configMax-configMin+1)
}
