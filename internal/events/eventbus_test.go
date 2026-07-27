package events

import (
	c "hazard/internal/citizen"
	h "hazard/internal/hazard"
	pf "hazard/internal/pathfinding"
	sz "hazard/internal/safezone"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCreateEventBus_InitializesFields(t *testing.T) {
	bus := New()

	require.NotNil(t, bus)
	require.NotNil(t, bus.SimulationCommands)
	require.NotNil(t, bus.SimulationEvents)
	require.Empty(t, bus.EventLog)
}

func TestCreateEventBus_ChannelsAreBuffered(t *testing.T) {
	bus := New()

	require.Equal(t, eventBufferSize, cap(bus.SimulationCommands))
	require.Equal(t, eventBufferSize, cap(bus.SimulationEvents))
}

func TestEventLog_AccumulatesEvents(t *testing.T) {
	bus := New()
	simID := uuid.New()
	meta := EventMetadata{SimulationID: simID, Tick: 0}

	bus.SimulationCreated(SimulationCreatedPayload{}, meta)
	bus.CitizenMoved(uuid.New(), pf.Position{}, EventMetadata{SimulationID: simID, Tick: 1})

	require.Len(t, bus.EventLog, 2)
}

func TestSimulationCreated_StoresCorrectType(t *testing.T) {
	bus := New()
	simID := uuid.New()
	grid := pf.NewGrid(3, 3, pf.CellOpen)
	payload := SimulationCreatedPayload{
		Grid:      grid.Copy(),
		Citizens:  []c.Citizen{},
		SafeZones: []sz.SafeZone{},
	}

	bus.SimulationCreated(payload, EventMetadata{SimulationID: simID, Tick: 0})

	require.Len(t, bus.EventLog, 1)
	evt := bus.EventLog[0]
	require.Equal(t, SimulationCreated, evt.EventType)
	require.Equal(t, simID, evt.SimulationID)
	require.Equal(t, uint64(0), evt.Tick)
	_, ok := evt.Payload.(SimulationCreatedPayload)
	require.True(t, ok, "payload should be SimulationCreatedPayload")
}

func TestSimulationCompleted_StoresCorrectType(t *testing.T) {
	bus := New()

	bus.SimulationCompleted(EventMetadata{SimulationID: uuid.New(), Tick: 5})

	require.Len(t, bus.EventLog, 1)
	require.Equal(t, SimulationCompleted, bus.EventLog[0].EventType)
	require.Equal(t, uint64(5), bus.EventLog[0].Tick)
}

func TestHazardEmerged_StoresPayload(t *testing.T) {
	bus := New()
	hazardID := uuid.New()
	hazardType := h.FireHazard
	pos := pf.Position{X: 3, Y: 4}

	bus.HazardEmerged(hazardID, HazardEmergedPayload{Type: hazardType, Position: pos},
		EventMetadata{SimulationID: uuid.New(), Tick: 1})

	require.Len(t, bus.EventLog, 1)
	evt := bus.EventLog[0]
	require.Equal(t, HazardEmerged, evt.EventType)
	require.Equal(t, hazardID, evt.EntityID)

	payload, ok := evt.Payload.(HazardEmergedPayload)
	require.True(t, ok)
	require.Equal(t, hazardType, payload.Type)
	require.Equal(t, pos, payload.Position)
}

func TestHazardExpanded_StoresUpdatedCells(t *testing.T) {
	bus := New()
	hazardID := uuid.New()
	cells := []pf.Position{{X: 0, Y: 0}, {X: 1, Y: 0}}

	bus.HazardExpanded(hazardID, cells, EventMetadata{SimulationID: uuid.New(), Tick: 2})

	require.Len(t, bus.EventLog, 1)
	evt := bus.EventLog[0]
	require.Equal(t, HazardExpanded, evt.EventType)
	require.Equal(t, hazardID, evt.EntityID)

	storedCells, ok := evt.Payload.([]pf.Position)
	require.True(t, ok)
	require.Equal(t, cells, storedCells)
}

func TestHazardDissipated_StoresUpdatedCells(t *testing.T) {
	bus := New()
	hazardID := uuid.New()
	cells := []pf.Position{{X: 5, Y: 5}, {X: 5, Y: 6}}

	bus.HazardDissipated(hazardID, cells, EventMetadata{SimulationID: uuid.New(), Tick: 3})

	require.Len(t, bus.EventLog, 1)
	require.Equal(t, HazardDissipated, bus.EventLog[0].EventType)
}

func TestCitizenEvents_StoreCorrectTypes(t *testing.T) {
	bus := New()
	citizenID := uuid.New()
	simID := uuid.New()

	bus.CitizenMoved(citizenID, pf.Position{X: 1, Y: 0},
		EventMetadata{SimulationID: simID, Tick: 1})
	bus.CitizenPathUpdated(citizenID, []pf.Position{{X: 0, Y: 0}},
		EventMetadata{SimulationID: simID, Tick: 1})
	bus.CitizenEscaped(citizenID, CitizenEscapedPayload{
		SafeZoneID:     uuid.New(),
		TotalEscaped:   1,
		TotalRemaining: 9,
	}, EventMetadata{SimulationID: simID, Tick: 2})
	bus.CitizenDied(citizenID, CitizenDiedPayload{
		TotalDead:      1,
		TotalRemaining: 9,
	}, EventMetadata{SimulationID: simID, Tick: 1})

	require.Len(t, bus.EventLog, 4)
	require.Equal(t, CitizenMoved, bus.EventLog[0].EventType)
	require.Equal(t, CitizenPathUpdated, bus.EventLog[1].EventType)
	require.Equal(t, CitizenEscaped, bus.EventLog[2].EventType)
	require.Equal(t, CitizenDied, bus.EventLog[3].EventType)
}

func TestSafeZoneEmerged_StoresCells(t *testing.T) {
	bus := New()
	safeZoneID := uuid.New()
	cells := []pf.Position{{X: 2, Y: 2}, {X: 3, Y: 2}}

	bus.SafeZoneEmerged(safeZoneID, SafeZoneEmergedPayload{ID: safeZoneID, Cells: cells},
		EventMetadata{SimulationID: uuid.New(), Tick: 1})

	require.Len(t, bus.EventLog, 1)
	require.Equal(t, SafeZoneEmerged, bus.EventLog[0].EventType)
	require.Equal(t, safeZoneID, bus.EventLog[0].EntityID)
}

func TestEventLog_TimestampsAreSet(t *testing.T) {
	bus := New()
	before := time.Now().UTC()

	bus.SimulationCreated(SimulationCreatedPayload{}, EventMetadata{SimulationID: uuid.New(), Tick: 0})

	after := time.Now().UTC()
	evt := bus.EventLog[0]
	require.False(t, evt.Timestamp.IsZero())
	require.True(t, evt.Timestamp.After(before) || evt.Timestamp.Equal(before))
	require.True(t, evt.Timestamp.Before(after) || evt.Timestamp.Equal(after))
}

func TestEventLog_EventIDsAreUnique(t *testing.T) {
	bus := New()
	meta := EventMetadata{SimulationID: uuid.New(), Tick: 0}

	bus.SimulationCreated(SimulationCreatedPayload{}, meta)
	bus.SimulationCompleted(meta)

	require.NotEqual(t, bus.EventLog[0].ID, bus.EventLog[1].ID,
		"each event must have a unique ID")
}

func TestEventBus_SimulationEventsChannelReceivesEvents(t *testing.T) {
	bus := New()
	simID := uuid.New()
	meta := EventMetadata{SimulationID: simID, Tick: 0}

	bus.SimulationCreated(SimulationCreatedPayload{}, meta)

	select {
	case evt := <-bus.SimulationEvents:
		require.Equal(t, SimulationCreated, evt.EventType)
		require.Equal(t, simID, evt.SimulationID)
	default:
		t.Fatal("expected event on SimulationEvents channel")
	}
}
