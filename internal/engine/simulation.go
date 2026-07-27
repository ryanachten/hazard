package engine

import (
	"log/slog"
	"math/rand"
	"slices"

	"github.com/google/uuid"

	"hazard/internal/citizen"
	"hazard/internal/configuration"
	"hazard/internal/events"
	"hazard/internal/hazard"
	"hazard/internal/obstacle"
	"hazard/internal/pathfinding"
	"hazard/internal/safezone"
)

// Simulation engine for hazards
type Simulation struct {
	ID                   uuid.UUID
	Config               configuration.SimulationConfig
	State                SimulationState
	TickCount            uint64
	Grid                 *pathfinding.Grid
	Citizens             []citizen.Citizen
	DeadCitizensCount    int
	EscapedCitizensCount int
	Hazards              []hazard.Hazard
	SafeZones            []*safezone.SafeZone
	eventBus             *events.EventBus
	safeZoneLocations    map[pathfinding.Position]*safezone.SafeZone
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
func NewSimulation(width, height int, config configuration.SimulationConfig, eventBus *events.EventBus) (Simulation, error) {
	grid := pathfinding.NewGrid(width, height, pathfinding.CellOpen)

	safeZone, err := safezone.Create(config.SafeZone, &grid)
	if err != nil {
		return Simulation{}, err
	}

	safeZoneLocations := make(map[pathfinding.Position]*safezone.SafeZone)
	for _, cell := range safeZone.Cells {
		safeZoneLocations[cell] = &safeZone
	}

	obstacles := obstacle.CreateObstacles(config.Obstacle, &grid)

	sim := Simulation{
		ID:                uuid.New(),
		Config:            config,
		State:             SimulationRunning,
		TickCount:         0,
		Grid:              &grid,
		SafeZones:         []*safezone.SafeZone{&safeZone},
		Citizens:          citizen.CreateCitizens(config.CitizenCount, &grid, safeZoneLocations),
		eventBus:          eventBus,
		safeZoneLocations: safeZoneLocations,
	}

	sim.eventBus.SimulationCreated(
		events.SimulationCreatedPayload{
			Grid:      sim.Grid.Copy(),
			Citizens:  sim.Citizens,
			SafeZones: sim.SafeZones,
			Obstacles: obstacles,
		},
		events.EventMetadata{
			SimulationID: sim.ID,
			Tick:         sim.TickCount,
		})

	return sim, nil
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
		c := &s.Citizens[i]

		if c.Status == citizen.StatusDead || c.Status == citizen.StatusEscaped {
			continue
		}

		isDead := s.removeDeadCitizen(c)
		if isDead {
			continue
		}

		s.updateCitizenPath(c, safeZoneCreated)
		s.updateCitizenLocation(c)
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
		hz := &s.Hazards[i]
		if s.TickCount > hz.CreatedAt+uint64(hz.Duration) {
			updatedCells := hz.Remove(s.Grid)
			s.Hazards = slices.Delete(s.Hazards, i, i+1)
			s.eventBus.HazardDissipated(hz.ID, updatedCells, s.getEventMetadata())
		} else {
			updatedCells := hz.Expand(s.Grid)
			s.eventBus.HazardExpanded(hz.ID, updatedCells, s.getEventMetadata())
		}
	}
}

func (s *Simulation) generateIntermittentHazard() {
	hazardConfig := s.Config.Hazard
	if len(s.Hazards) >= hazardConfig.Count || rand.Float32() > hazardConfig.Probability {
		return
	}

	hz, err := hazard.Create(hazardConfig, s.Grid)
	if err != nil {
		slog.Warn("error creating hazard", "err", err)
		return
	}

	hz.CreatedAt = s.TickCount
	s.Hazards = append(s.Hazards, hz)

	s.eventBus.HazardEmerged(hz.ID, events.HazardEmergedPayload{
		Type:     hz.Type,
		Position: hz.Origin,
	}, s.getEventMetadata())
}

func (s *Simulation) generateIntermittentSafeZone() bool {
	safeZoneConfig := s.Config.SafeZone
	if len(s.SafeZones) >= safeZoneConfig.Count || rand.Float32() > safeZoneConfig.Probability {
		return false
	}

	safeZone, err := safezone.Create(safeZoneConfig, s.Grid)
	if err != nil {
		slog.Warn("error creating safe zone", "err", err)
		return false
	}

	s.SafeZones = append(s.SafeZones, &safeZone)

	for _, cell := range safeZone.Cells {
		s.safeZoneLocations[cell] = &safeZone
	}

	s.eventBus.SafeZoneEmerged(safeZone.ID, events.SafeZoneEmergedPayload{
		ID:    safeZone.ID,
		Cells: safeZone.Cells,
	}, s.getEventMetadata())

	return true
}

func (s *Simulation) removeDeadCitizen(c *citizen.Citizen) bool {
	if c.Status == citizen.StatusEscaped {
		return false
	}

	if s.Grid.GetCell(c.CurrentPosition) != pathfinding.CellHazard {
		return false
	}

	c.Status = citizen.StatusDead
	s.Grid.UpdateCell(c.CurrentPosition, pathfinding.CellDeadCitizen)
	s.DeadCitizensCount++
	s.eventBus.CitizenDied(c.ID, events.CitizenDiedPayload{
		TotalDead:      s.DeadCitizensCount,
		TotalRemaining: len(s.Citizens) - s.DeadCitizensCount - s.EscapedCitizensCount,
	}, s.getEventMetadata())

	return true
}

func (s *Simulation) updateCitizenPath(c *citizen.Citizen, safeZoneCreated bool) {
	pathUpdated := false

	// If no safe zone assigned, new safe zone added, or the target has no capacity
	// determine which available safe zone is closest
	targetSafeZone := c.TargetSafeZone
	if safeZoneCreated || targetSafeZone == nil || !targetSafeZone.HasCapacity {
		if err := c.FindNearestSafeZone(s.Grid, s.safeZoneLocations); err != nil {
			c.Path = nil
		}
		pathUpdated = true
	} else {
		pathUpdated = s.verifyAndUpdateBlockedPath(c)
	}

	if pathUpdated {
		s.eventBus.CitizenPathUpdated(c.ID, c.Path, s.getEventMetadata())
	}
}

// Check if next cell intersects with avoidable cell types and needs recalculating
func (s *Simulation) verifyAndUpdateBlockedPath(c *citizen.Citizen) bool {
	curIndex := c.CurrentPathIndex
	nextIndex := curIndex + 1
	if nextIndex < len(c.Path) {
		pos := c.Path[nextIndex]
		if pathfinding.AvoidableCellType[s.Grid.GetCell(pos)] {
			if err := c.RecalculatePath(s.Grid); err != nil {
				// Destination itself is blocked (e.g. occupied by another citizen).
				// Find a new safe zone with capacity.
				if err := c.FindNearestSafeZone(s.Grid, s.safeZoneLocations); err != nil {
					c.Path = nil
				}
			}
			return true
		}
	}
	return false
}

func (s *Simulation) updateCitizenLocation(c *citizen.Citizen) {
	hasMoved, hasEscaped := c.IncrementLocation(s.Grid)
	if hasMoved {
		s.eventBus.CitizenMoved(c.ID, c.CurrentPosition, s.getEventMetadata())
	}

	if hasEscaped {
		safeZone := s.safeZoneLocations[c.CurrentPosition]
		assignedPosition, hasCapacity := safeZone.AddOccupant(c.ID, c.CurrentPosition, s.Grid)

		if !hasCapacity {
			s.updateCitizenPath(c, true)
		} else {
			s.EscapedCitizensCount++
			c.Status = citizen.StatusEscaped

			if assignedPosition != c.CurrentPosition {
				s.Grid.UpdateCell(c.CurrentPosition, c.PreviousCellType)
				c.CurrentPosition = assignedPosition
			}

			s.eventBus.CitizenEscaped(c.ID, events.CitizenEscapedPayload{
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
