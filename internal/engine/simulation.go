package engine

import (
	c "hazard/internal/citizen"
	cfg "hazard/internal/configuration"
	"hazard/internal/events"
	h "hazard/internal/hazard"
	o "hazard/internal/obstacle"
	pf "hazard/internal/pathfinding"
	sz "hazard/internal/safe_zone"
	"log/slog"
	"math/rand"
	"slices"

	"github.com/google/uuid"
)

// Simulation engine for hazards
type Simulation struct {
	ID                   uuid.UUID
	Config               cfg.SimulationConfig
	State                SimulationState
	TickCount            uint64
	Grid                 *pf.Grid
	Citizens             []c.Citizen
	DeadCitizensCount    int
	EscapedCitizensCount int
	Hazards              []h.Hazard
	SafeZones            []sz.SafeZone
	eventBus             *events.EventBus
	safeZoneLocations    map[pf.Position]*sz.SafeZone
}

// SimulationState phases of a simulation
type SimulationState string

const (
	// SimulationCreated simulation created but not running
	SimulationCreated SimulationState = "created"
	// SimulationRunning simulation running
	SimulationRunning SimulationState = "running"
	// SimulationPaused simulation paused
	SimulationPaused SimulationState = "paused"
	// SimulationCompleted simulation completed
	SimulationCompleted SimulationState = "completed"
)

// NewSimulation creates a simulation based on configuration
func NewSimulation(width, height int, config cfg.SimulationConfig, eventBus *events.EventBus) (Simulation, error) {
	grid := pf.NewGrid(width, height, pf.CellOpen)

	safeZone, err := sz.Create(config.SafeZone, &grid)
	if err != nil {
		return Simulation{}, err
	}

	safeZoneLocations := make(map[pf.Position]*sz.SafeZone)
	for _, cell := range safeZone.Cells {
		safeZoneLocations[cell] = &safeZone
	}

	obstacles := o.CreateObstacles(config.Obstacle, &grid)

	var simulation = Simulation{
		ID:                uuid.New(),
		Config:            config,
		State:             SimulationRunning,
		TickCount:         0,
		Grid:              &grid,
		SafeZones:         []sz.SafeZone{safeZone},
		Citizens:          c.CreateCitizens(config.CitizenCount, &grid, safeZoneLocations),
		eventBus:          eventBus,
		safeZoneLocations: safeZoneLocations,
	}

	simulation.eventBus.SimulationCreated(
		events.SimulationCreatedPayload{
			Grid:      simulation.Grid.Copy(),
			Citizens:  simulation.Citizens,
			SafeZones: simulation.SafeZones,
			Obstacles: obstacles,
		},
		events.EventMetadata{
			SimulationID: simulation.ID,
			Tick:         simulation.TickCount,
		})

	return simulation, nil
}

// Tick increments a simulation by one tick
func (s *Simulation) Tick() {
	if s.State == SimulationPaused || s.State == SimulationCompleted {
		return
	}
	s.State = SimulationRunning

	s.updateOrRemoveHazards()
	s.generateIntermittentHazard()

	safeZoneCreated := s.generateIntermittentSafeZone()

	for i := range s.Citizens {

		citizen := &s.Citizens[i]

		if citizen.Status == c.StatusDead || citizen.Status == c.StatusEscaped {
			continue
		}

		isDead := s.removeDeadCitizen(citizen)
		if isDead {
			continue
		}

		s.updateCitizenPath(citizen, safeZoneCreated)
		s.updateCitizenLocation(citizen)
	}

	s.TickCount++

	if len(s.Citizens) > 0 && s.DeadCitizensCount+s.EscapedCitizensCount == len(s.Citizens) {
		s.eventBus.SimulationCompleted(s.getEventMetadata())
		s.State = SimulationCompleted
	}
}

// ProcessCommand handles a simulation command, updating state accordingly.
func (s *Simulation) ProcessCommand(cmd events.SimulationCommand) {
	switch cmd.CommandType {
	case events.PauseSimulation:
		if s.State == SimulationRunning {
			s.State = SimulationPaused
		} else {
			s.State = SimulationRunning
		}
	case events.UpdateHazardProbability:
		if probability, ok := parsePayload[float32](cmd.Payload); ok {
			s.Config.Hazard.Probability = probability
		}
	case events.UpdateHazardCount:
		if count, ok := parsePayload[int](cmd.Payload); ok {
			s.Config.Hazard.Count = count
		}
	case events.UpdateHazardDurationMin:
		if duration, ok := parsePayload[int](cmd.Payload); ok {
			s.Config.Hazard.DurationRange.Min = duration
		}
	case events.UpdateHazardDurationMax:
		if duration, ok := parsePayload[int](cmd.Payload); ok {
			s.Config.Hazard.DurationRange.Max = duration
		}
	case events.UpdateSafeZoneProbability:
		if probability, ok := parsePayload[float32](cmd.Payload); ok {
			s.Config.SafeZone.Probability = probability
		}
	case events.UpdateSafeZoneCount:
		if count, ok := parsePayload[int](cmd.Payload); ok {
			s.Config.SafeZone.Count = count
		}
	case events.UpdateSafeZoneRadiusMin:
		if radius, ok := parsePayload[int](cmd.Payload); ok {
			s.Config.SafeZone.RadiusRange.Min = radius
		}
	case events.UpdateSafeZoneRadiusMax:
		if radius, ok := parsePayload[int](cmd.Payload); ok {
			s.Config.SafeZone.RadiusRange.Max = radius
		}
	}
}

func (s *Simulation) updateOrRemoveHazards() {
	for i := len(s.Hazards) - 1; i >= 0; i-- {
		hazard := &s.Hazards[i]
		if s.TickCount > hazard.CreatedAt+uint64(hazard.Duration) {
			updatedCells := hazard.Remove(s.Grid)
			s.Hazards = slices.Delete(s.Hazards, i, i+1)
			s.eventBus.HazardDissipated(hazard.ID, updatedCells, s.getEventMetadata())
		} else {
			updatedCells := hazard.Expand(s.Grid)
			s.eventBus.HazardExpanded(hazard.ID, updatedCells, s.getEventMetadata())
		}
	}
}

func (s *Simulation) generateIntermittentHazard() {
	hazardConfig := s.Config.Hazard
	if len(s.Hazards) >= hazardConfig.Count || rand.Float32() > hazardConfig.Probability {
		return
	}

	hazard, err := h.Create(hazardConfig, s.Grid)
	if err != nil {
		slog.Warn("error creating hazard", "err", err)
		return
	}

	hazard.CreatedAt = s.TickCount
	s.Hazards = append(s.Hazards, hazard)

	s.eventBus.HazardEmerged(hazard.ID, events.HazardEmergedPayload{
		Type:     hazard.Type,
		Position: hazard.Origin,
	}, s.getEventMetadata())
}

func (s *Simulation) generateIntermittentSafeZone() bool {
	safeZoneConfig := s.Config.SafeZone
	if len(s.SafeZones) >= safeZoneConfig.Count || rand.Float32() > safeZoneConfig.Probability {
		return false
	}

	safeZone, err := sz.Create(safeZoneConfig, s.Grid)
	if err != nil {
		slog.Warn("error creating safe zone", "err", err)
		return false
	}

	s.SafeZones = append(s.SafeZones, safeZone)

	for _, cell := range safeZone.Cells {
		s.safeZoneLocations[cell] = &safeZone
	}

	s.eventBus.SafeZoneEmerged(safeZone.ID, events.SafeZoneEmergedPayload{
		ID:    safeZone.ID,
		Cells: safeZone.Cells,
	}, s.getEventMetadata())

	return true
}

func (s *Simulation) removeDeadCitizen(citizen *c.Citizen) bool {
	if citizen.Status == c.StatusEscaped {
		return false
	}

	if s.Grid.GetCell(citizen.CurrentPosition) != pf.CellHazard {
		return false
	}

	citizen.Status = c.StatusDead
	s.Grid.UpdateCell(citizen.CurrentPosition, pf.CellDeadCitizen)
	s.DeadCitizensCount++
	s.eventBus.CitizenDied(citizen.ID, events.CitizenDiedPayload{
		TotalDead:      s.DeadCitizensCount,
		TotalRemaining: len(s.Citizens) - s.DeadCitizensCount - s.EscapedCitizensCount,
	}, s.getEventMetadata())

	return true
}

func (s *Simulation) updateCitizenPath(citizen *c.Citizen, safeZoneCreated bool) {
	pathUpdated := false

	// If no safe zone assigned, new safe zone added, or the target has no capacity
	// determine which available safe zone is closest
	targetSafeZone := citizen.TargetSafeZone
	if safeZoneCreated || targetSafeZone == nil || !targetSafeZone.HasCapacity {
		if err := citizen.FindNearestSafeZone(s.Grid, s.safeZoneLocations); err != nil {
			citizen.Path = nil
		}
		pathUpdated = true
	} else {
		pathUpdated = s.verifyAndUpdateBlockedPath(citizen)
	}

	if pathUpdated {
		s.eventBus.CitizenPathUpdated(citizen.ID, citizen.Path, s.getEventMetadata())
	}
}

// Check if next cell intersects with avoidable cell types and needs recalculating
func (s *Simulation) verifyAndUpdateBlockedPath(citizen *c.Citizen) bool {
	curIndex := citizen.CurrentPathIndex
	nextIndex := curIndex + 1
	if nextIndex < len(citizen.Path) {
		pos := citizen.Path[nextIndex]
		if pf.AvoidableCellType[s.Grid.GetCell(pos)] {
			if err := citizen.RecalculatePath(s.Grid); err != nil {
				// Destination itself is blocked (e.g. occupied by another citizen).
				// Find a new safe zone with capacity.
				if err := citizen.FindNearestSafeZone(s.Grid, s.safeZoneLocations); err != nil {
					citizen.Path = nil
				}
			}
			return true
		}
	}
	return false
}

func (s *Simulation) updateCitizenLocation(citizen *c.Citizen) {
	hasMoved, hasEscaped := citizen.IncrementLocation(s.Grid)
	if hasMoved {
		s.eventBus.CitizenMoved(citizen.ID, citizen.CurrentPosition, s.getEventMetadata())
	}

	if hasEscaped {
		safeZone := s.safeZoneLocations[citizen.CurrentPosition]
		assignedPosition, hasCapacity := safeZone.AddOccupant(citizen.ID, citizen.CurrentPosition, s.Grid)

		if !hasCapacity {
			s.updateCitizenPath(citizen, true)
		} else {
			s.EscapedCitizensCount++
			citizen.Status = c.StatusEscaped

			if assignedPosition != citizen.CurrentPosition {
				s.Grid.UpdateCell(citizen.CurrentPosition, citizen.PreviousCellType)
				citizen.CurrentPosition = assignedPosition
			}

			s.eventBus.CitizenEscaped(citizen.ID, events.CitizenEscapedPayload{
				SafeZoneID:       safeZone.ID,
				AssignedPosition: assignedPosition,
				TotalEscaped:     s.EscapedCitizensCount,
				TotalRemaining:   len(s.Citizens) - s.DeadCitizensCount - s.EscapedCitizensCount,
			}, s.getEventMetadata())
		}
	}
}

func (s *Simulation) getEventMetadata() events.EventMetadata {
	return events.EventMetadata{
		SimulationID: s.ID,
		Tick:         s.TickCount,
	}
}

func parsePayload[T any](payload any) (T, bool) {
	parsedPayload, ok := payload.(T)
	if !ok {
		slog.Error("error parsing payload", "payload", payload)
	}
	return parsedPayload, ok
}
