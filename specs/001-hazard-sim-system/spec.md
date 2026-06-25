# Feature Specification: Hazard Simulation System

**Feature Branch**: `001-hazard-sim-system`  
**Created**: 2026-06-13  
**Status**: Draft  
**Input**: User description: "lets define the plan for a hazard simulation system. I want to create a system that will allow us to simulate hazards and guide citizens to avoid these hazards. The objective of this system will be for me to gain greater familiarity with golang, event streaming with NATS/JetStream, event driven development practices and pathfinding algorithms. The system will also have a visual component to it, where we can visualise the citizens, their movements and the emergence of hazards, likely using websockets and some simple form of rendered output."

## User Scenarios & Testing

### User Story 1 - Configure and Run a Hazard Simulation (Priority: P1)

A simulation operator defines an environment (including static obstacles like buildings and terrain), places safe zones, seeds citizens, and configures hazard parameters (emergence interval, type, progression speed). The simulation begins with citizens navigating toward safe zones. Hazards emerge periodically, potentially enveloping areas and threatening citizens. Citizens must navigate around obstacles and hazards to reach safety. Simulation ends when all citizens have either escaped to a safe zone or died.

**Why this priority**: Running the simulation is the core capability; without this there is no system.

**Independent Test**: Can be tested by defining a minimal environment with one citizen, one safe zone, and a hazard configured to emerge at a known interval, then verifying the citizen navigates to safety before the hazard reaches them.

**Acceptance Scenarios**:

1. **Given** a configured environment with citizens seeded at start positions and safe zones placed, **When** the operator starts the simulation, **Then** citizens begin moving toward safe zones
2. **Given** an active simulation, **When** a hazard emerges in a citizen's path, **Then** the citizen recalculates its route to avoid the hazard
3. **Given** an active simulation, **When** a hazard overtakes a citizen, **Then** the citizen is marked as dead
4. **Given** an active simulation, **When** all citizens have either reached a safe zone or died, **Then** the simulation ends automatically
5. **Given** an active simulation, **When** the operator stops the simulation, **Then** all citizen movement ceases

---

### User Story 2 - Observe Simulation in Real-Time (Priority: P2)

An observer connects to the simulation and watches a live visualisation of the environment — including static obstacles, citizen positions, hazard zones (with progressive envelopment), safe zones, and navigation paths — updating in real-time.

**Why this priority**: Real-time visualisation is essential for understanding how hazards affect citizen movement and for demonstrating the system's value.

**Independent Test**: Can be tested by connecting a visualisation client to a running simulation and verifying that citizen positions, hazard growth, and state changes update on the display within reasonable latency.

**Acceptance Scenarios**:

1. **Given** a running simulation, **When** an observer connects to the visualisation, **Then** they see the current state of all citizens, hazards, safe zones, static obstacles, and the environment
2. **Given** an observer watching the visualisation, **When** a citizen moves, a hazard expands, or a citizen dies, **Then** the display updates to reflect the change within one second
3. **Given** an observer watching the visualisation, **When** a citizen successfully reaches a safe zone or dies, **Then** this is visually indicated
4. **Given** an observer watching the visualisation, **When** a hazard progressively envelops new area, **Then** the visualisation shows the expanding hazard zone

---

### User Story 3 - Simulation Produces Navigable Event History (Priority: P3)

An operator replays or inspects the sequence of events that occurred during a simulation to analyse citizen behavior, hazard progression, death events, and pathfinding outcomes after the simulation ends.

**Why this priority**: Streaming events is a core learning goal; retaining them enables analysis and debugging of pathfinding and hazard behavior.

**Independent Test**: Can be tested by running a short simulation with at least one citizen death, then inspecting the recorded event stream to verify it contains a complete, ordered sequence of state changes including movement, hazard emergence, envelopment, and death.

**Acceptance Scenarios**:

1. **Given** a completed simulation, **When** the operator requests the event history, **Then** a complete ordered sequence of all state changes is available
2. **Given** a simulation event history, **When** inspected, **Then** each event includes a timestamp, entity identifier, event type, and state change description
3. **Given** a simulation event history, **When** a citizen died during the simulation, **Then** the death event is recorded with the responsible hazard and timestamp

---

### User Story 4 - Multi-Hazard Scenario with Large Citizen Count (Priority: P3)

An operator creates a dense simulation with multiple concurrent hazards (with varied types and progression rates), many citizens, and complex obstacles, verifying that the system handles scale while maintaining real-time performance.

**Why this priority**: Demonstrates the system's capacity and surfaces performance characteristics relevant to the learning objectives.

**Independent Test**: Can be tested by configuring a simulation with 100+ citizens, 5+ concurrent hazards, and static obstacles, running it, and measuring whether visualisation updates remain fluid.

**Acceptance Scenarios**:

1. **Given** a simulation with 100 citizens, 5 hazards, and obstacles, **When** the simulation runs for 60 seconds, **Then** all citizens continue navigating without visible degradation
2. **Given** a large-scale simulation, **When** observed in real-time, **Then** visualisation updates at a consistent rate

---

### Edge Cases

- What happens when a hazard's progressive envelopment completely blocks all paths to all safe zones?
- How does the system handle a citizen that cannot outpace an expanding hazard?
- What happens when all citizens die before any reach a safe zone?
- What happens when one citizen remains blocked from all safe zones while others have escaped or died?
- How does the system handle a simulation where no hazards emerge (e.g., interval set to zero)?
- What happens if a safe zone becomes enveloped by a hazard before citizens can reach it?
- How does the system behave when a hazard emerges directly on top of a citizen (instant death)?
- What happens if the visualisation client disconnects mid-simulation?
- How does the system handle a simulation with zero citizens?
- What happens when a citizen has no valid path to any safe zone at simulation start?

## Requirements

### Functional Requirements

- **FR-001**: Operators MUST be able to define a 2D simulation environment with configurable dimensions and place static obstacles (e.g., buildings, terrain features) that block movement
- **FR-002**: Operators MUST be able to place one or more safe zones in the environment. At least one safe zone MUST be present at simulation start; additional safe zones MAY emerge dynamically during the simulation
- **FR-003**: Operators MUST be able to seed citizens in the environment before hazards emerge, each with a start position and the goal of reaching any safe zone
- **FR-004**: Hazards MUST emerge at configurable regular intervals after citizens have been seeded, with each hazard having a position, initial radius, progression speed, duration, and type
- **FR-005**: Hazards MUST support configurable progressive envelopment — expanding their effective radius over time to cover additional area
- **FR-006**: Citizens MUST autonomously navigate from their start position toward safe zones while avoiding static obstacles and active hazard zones
- **FR-007**: Citizens MUST recalculate their route when a new hazard emerges, a new safe zone appears, or an existing hazard expands to block their current path
- **FR-008**: If a hazard overtakes a citizen (citizen cannot outpace or avoid the expanding hazard), the citizen MUST be marked as dead
- **FR-009**: The simulation MUST end automatically when all citizens have either reached a safe zone or died
- **FR-010**: Operators MUST be able to start, pause, and stop a simulation
- **FR-011**: The system MUST stream all state changes (citizen moves, hazard emergence, hazard expansion, citizen arrival at safe zone, citizen death) as events to connected visualisation clients
- **FR-012**: The visualisation MUST display the environment, static obstacles, safe zones, citizen positions, hazard zones (with progressive growth), and citizen navigation paths
- **FR-013**: The visualisation MUST update in real-time as events are received
- **FR-014**: The system MUST support at least 100 simultaneous citizens in a single simulation
- **FR-015**: The system MUST support at least 10 simultaneous hazards in a single simulation
- **FR-016**: The system MUST record all simulation events in order with timestamps for post-simulation review
- **FR-017**: Hazard types MUST be defined in app code with a built-in registry — initially fire, flood, lava (all `expanding` kind), with the ability to add new kinds (e.g., `strike` for lightning, `global` for earthquake) by extending the registry. Config references them by name only; behavioral properties are not configurable at runtime.
- **FR-018**: Citizens MUST have a configurable degree of autonomy influencing their movement decisions (e.g., path preference, risk tolerance, speed variation)
- **FR-019**: The hazard emergence interval MUST be configurable per simulation

### Key Entities

- **Simulation**: A single run encompassing environment configuration, citizens, hazards, safe zones, and the event stream. Has a lifecycle (created, running, paused, completed).
- **Citizen**: An autonomous agent with start position, current position, status (navigating, escaped, dead), movement speed, autonomy profile, and goal of reaching any safe zone. Responds to hazards and obstacles by recalculating path.
- **Hazard**: An obstacle with position, type (extensible — initially fire, flood, lava), initial radius, progression speed, maximum radius, severity, duration, and creation tick. Created active at emergence and removed when expired. Behavior governed by `HazardKind` (expanding, strike, global).
- **Safe Zone**: A designated area within the environment that citizens must reach to survive. At least one safe zone is placed before simulation start; additional safe zones may emerge dynamically mid-simulation. Citizens are considered escaped upon entry.
- **Environment**: A bounded 2D space with configurable width and height containing static obstacles, safe zones, citizens, and hazards.
- **Static Obstacle**: Impassable terrain features (buildings, landscape) that block citizen movement and hazard progression. Defined at environment setup.
- **Simulation Event**: A time-stamped record of any state change. Includes event type, entity ID, timestamp, and payload describing the change.
- **Visualisation Client**: A connected viewer that receives simulation events and renders the current state.

## Success Criteria

### Measurable Outcomes

- **SC-001**: A simulation with 100 citizens, 5 hazards, and static obstacles runs to completion (all citizens either escaped or dead) within 2x real-time
- **SC-002**: Citizens successfully escape to a safe zone in at least 70% of cases when a viable path exists at simulation start
- **SC-003**: State changes are reflected in the visualisation within one second of occurring in the simulation
- **SC-004**: An operator can configure hazard emergence interval and observe hazards appearing at that interval
- **SC-005**: An operator can start, pause, and stop a simulation through the provided controls
- **SC-006**: A complete ordered event history — including movement, hazard emergence, hazard expansion, escape, and death events — is available for inspection after simulation completion
- **SC-007**: Multiple visualisation clients can observe the same simulation simultaneously
- **SC-008**: The hazard type system can be extended with a new hazard type (different visual, different progression behavior) without modifying core simulation logic

## Assumptions

- The simulation uses a 2D grid-based environment for pathfinding
- Citizens move a configurable number of grid cells per simulation tick
- Hazards expand outward from their origin point at a configurable rate
- At least one safe zone is placed before simulation start; additional safe zones may emerge dynamically mid-simulation
- Static obstacles are impassable to both citizens and hazards
- The visualisation is rendered in a web browser using simple canvas or SVG graphics
- The simulation runs on a single machine for the initial version
- Standard grid-based pathfinding algorithms (e.g., A*) are sufficient for citizen navigation in initial versions
- Citizens have basic autonomy — they can choose between multiple viable paths based on configurable preferences (e.g., shortest path vs. safest path)
- The system will be built in progressive phases: (1) basic movement + static obstacles + expanding hazards, (2) citizen death + dynamic safe zones, (3) hazard types + progressive envelopment, (4) citizen autonomy, (5) additional hazard kinds (strike, global)
