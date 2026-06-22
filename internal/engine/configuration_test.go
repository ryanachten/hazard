package engine

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSimulationConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  SimulationConfig
		wantErr string
	}{
		{
			name: "valid config",
			config: SimulationConfig{
				Width:             5,
				Height:            5,
				CitizenCountRange: [2]int{1, 5},
				Hazard: HazardConfig{
					DurationRange: [2]int{1, 10},
					Probability:   0.5,
				},
				SafeZone: SafeZoneConfig{
					CountRange: [2]int{1, 1},
				},
			},
			wantErr: "",
		},
		{
			name: "zero width",
			config: SimulationConfig{
				Width:  0,
				Height: 5,
			},
			wantErr: "width and height must be greater than zero",
		},
		{
			name: "zero height",
			config: SimulationConfig{
				Width:  5,
				Height: 0,
			},
			wantErr: "width and height must be greater than zero",
		},
		{
			name: "negative width",
			config: SimulationConfig{
				Width:  -1,
				Height: 5,
			},
			wantErr: "width and height must be greater than zero",
		},
		{
			name: "citizen count min greater than max",
			config: SimulationConfig{
				Width:             5,
				Height:            5,
				CitizenCountRange: [2]int{5, 3},
			},
			wantErr: "CitizenCountRange[0] must be less than or equal to CitizenCountRange[1]",
		},
		{
			name: "hazard duration min greater than max",
			config: SimulationConfig{
				Width:             5,
				Height:            5,
				CitizenCountRange: [2]int{1, 5},
				Hazard: HazardConfig{
					DurationRange: [2]int{10, 5},
				},
			},
			wantErr: "Hazard.HazardDurationRange[0] must be less than or equal to Hazard.HazardDurationRange[1]",
		},
		{
			name: "hazard duration range zero",
			config: SimulationConfig{
				Width:             5,
				Height:            5,
				CitizenCountRange: [2]int{1, 5},
				Hazard: HazardConfig{
					DurationRange: [2]int{0, 0},
				},
			},
			wantErr: "Hazard.HazardDurationRange values must be above 0",
		},
		{
			name: "hazard count min greater than max",
			config: SimulationConfig{
				Width:             5,
				Height:            5,
				CitizenCountRange: [2]int{1, 5},
				Hazard: HazardConfig{
					DurationRange: [2]int{1, 1},
					CountRange:    [2]int{5, 3},
				},
			},
			wantErr: "Hazard.CountRange[0] must be less than or equal to Hazard.CountRange[1]",
		},
		{
			name: "negative hazard probability",
			config: SimulationConfig{
				Width:             5,
				Height:            5,
				CitizenCountRange: [2]int{1, 5},
				Hazard: HazardConfig{
					DurationRange: [2]int{1, 10},
					Probability:   -0.1,
				},
			},
			wantErr: "Hazard.HazardProbability must be between 0.0 and 1.0",
		},
		{
			name: "hazard probability greater than 1",
			config: SimulationConfig{
				Width:             5,
				Height:            5,
				CitizenCountRange: [2]int{1, 5},
				Hazard: HazardConfig{
					DurationRange: [2]int{1, 10},
					Probability:   1.5,
				},
			},
			wantErr: "Hazard.HazardProbability must be between 0.0 and 1.0",
		},
		{
			name: "safe zone probability negative",
			config: SimulationConfig{
				Width:             5,
				Height:            5,
				CitizenCountRange: [2]int{1, 5},
				SafeZone: SafeZoneConfig{
					Probability: -0.1,
				},
			},
			wantErr: "SafeZone.Probability must be between 0.0 and 1.0",
		},
		{
			name: "safe zone probability greater than 1",
			config: SimulationConfig{
				Width:             5,
				Height:            5,
				CitizenCountRange: [2]int{1, 5},
				SafeZone: SafeZoneConfig{
					Probability: 1.5,
				},
			},
			wantErr: "SafeZone.Probability must be between 0.0 and 1.0",
		},
		{
			name: "safe zone radius range min greater than max",
			config: SimulationConfig{
				Width:             5,
				Height:            5,
				CitizenCountRange: [2]int{1, 5},
				SafeZone: SafeZoneConfig{
					RadiusRange: [2]int{5, 3},
				},
			},
			wantErr: "SafeZone.RadiusRange[0] must be less than or equal to SafeZone.RadiusRange[1]",
		},
		{
			name: "safe zone count range min greater than max",
			config: SimulationConfig{
				Width:             5,
				Height:            5,
				CitizenCountRange: [2]int{1, 5},
				SafeZone: SafeZoneConfig{
					CountRange: [2]int{5, 3},
				},
			},
			wantErr: "SafeZone.CountRange[0] must be less than or equal to SafeZone.CountRange[1]",
		},
		{
			name: "safe zone max count less than 1",
			config: SimulationConfig{
				Width:             5,
				Height:            5,
				CitizenCountRange: [2]int{1, 5},
				SafeZone: SafeZoneConfig{
					CountRange: [2]int{0, 0},
				},
			},
			wantErr: "SafeZone.CountRange[1] must be at least 1",
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
