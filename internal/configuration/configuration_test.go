package configuration

import (
	"testing"

	"github.com/stretchr/testify/require"

	"hazard/internal/bounds"
	"hazard/internal/hazard"
	"hazard/internal/obstacle"
	"hazard/internal/safezone"
)

func TestSimulationConfig_Validate(t *testing.T) {
	validConfig := SimulationConfig{
		CitizenCount: 5,
		Hazard: hazard.Config{
			DurationRange: bounds.PositiveRange{Min: 1, Max: 10},
			Count:         5,
			Probability:   0.5,
		},
		SafeZone: safezone.Config{
			Count:       1,
			RadiusRange: bounds.Range{Min: 1, Max: 1},
		},
		Obstacle: obstacle.Config{
			CountRange: bounds.Range{Min: 1, Max: 5},
			SizeRange:  bounds.PositiveRange{Min: 2, Max: 5},
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
			wantErr: "hazard probability must be between 0.0 and 1.0",
		},
		{
			name: "hazard probability greater than 1",
			config: func() SimulationConfig {
				c := validConfig
				c.Hazard.Probability = 1.5
				return c
			}(),
			wantErr: "hazard probability must be between 0.0 and 1.0",
		},
		{
			name: "safe zone probability negative",
			config: func() SimulationConfig {
				c := validConfig
				c.SafeZone.Probability = -0.1
				return c
			}(),
			wantErr: "safe zone probability must be between 0.0 and 1.0",
		},
		{
			name: "safe zone probability greater than 1",
			config: func() SimulationConfig {
				c := validConfig
				c.SafeZone.Probability = 1.5
				return c
			}(),
			wantErr: "safe zone probability must be between 0.0 and 1.0",
		},
		{
			name: "safe zone count less than 1",
			config: func() SimulationConfig {
				c := validConfig
				c.SafeZone.Count = 0
				return c
			}(),
			wantErr: "safe zone count must be at least 1",
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
