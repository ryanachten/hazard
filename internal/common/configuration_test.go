package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSimulationConfig_Validate(t *testing.T) {
	validConfig := SimulationConfig{
		CitizenCount: 5,
		Hazard: HazardConfig{
			DurationRange: PositiveRange{Min: 1, Max: 10},
			Count:         5,
			Probability:   0.5,
		},
		SafeZone: SafeZoneConfig{
			Count:       1,
			RadiusRange: Range{Min: 1, Max: 1},
		},
		Obstacle: ObstacleConfig{
			CountRange: Range{Min: 1, Max: 5},
			SizeRange:  PositiveRange{Min: 2, Max: 5},
		},
	}

	tests := []struct {
		name    string
		config  SimulationConfig
		wantErr string
	}{
		{
			name:    "valid config",
			config:  validConfig,
			wantErr: "",
		},
		{
			name: "negative hazard probability",
			config: func() SimulationConfig {
				c := validConfig
				c.Hazard.Probability = -0.1
				return c
			}(),
			wantErr: "Hazard.Probability must be between 0.0 and 1.0",
		},
		{
			name: "hazard probability greater than 1",
			config: func() SimulationConfig {
				c := validConfig
				c.Hazard.Probability = 1.5
				return c
			}(),
			wantErr: "Hazard.Probability must be between 0.0 and 1.0",
		},
		{
			name: "safe zone probability negative",
			config: func() SimulationConfig {
				c := validConfig
				c.SafeZone.Probability = -0.1
				return c
			}(),
			wantErr: "SafeZone.Probability must be between 0.0 and 1.0",
		},
		{
			name: "safe zone probability greater than 1",
			config: func() SimulationConfig {
				c := validConfig
				c.SafeZone.Probability = 1.5
				return c
			}(),
			wantErr: "SafeZone.Probability must be between 0.0 and 1.0",
		},
		{
			name: "safe zone count less than 1",
			config: func() SimulationConfig {
				c := validConfig
				c.SafeZone.Count = 0
				return c
			}(),
			wantErr: "SafeZone.Count must be at least 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr)
			}
		})
	}
}
