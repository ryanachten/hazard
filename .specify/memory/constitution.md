<!--
  Sync Impact Report:
  - Version change: (new constitution) → v1.0.0
  - Principles defined: Small Slice Delivery, Agent-Assisted Code Review,
    Idiomatic Go, Event-Driven Architecture, Pathfinding & Autonomy
  - Added sections: Technology Stack & Infrastructure,
    Development Workflow & Quality Gates
  - Templates requiring updates: none (templates are generic scaffolding)
  - Follow-up TODOs: none
-->

# Hazard Constitution

## Core Principles

### I. Small Slice Delivery

Features MUST be developed in small, independently deliverable slices. Each
slice MUST provide incremental value, be independently testable, and be
completable within a focused timeframe. A slice SHOULD NOT exceed what one
person can implement, test, and review in a single session. Explicit
justification is required for any feature branch accumulating changes beyond
this scope.

### II. Agent-Assisted Code Review

Agentic systems MUST be used for code review rather than direct code
generation. All production code MUST be written by the developer; agentic
systems review for correctness, idiomatic style, test coverage, and adherence
to this constitution. Code review feedback from agents MUST be addressed
before merging. Agents MAY generate small snippets (tests, configuration,
boilerplate) but MUST NOT author feature logic.

### III. Idiomatic Go

All Go code MUST conform to established Go idioms and conventions. This
includes: passing `go vet` and `staticcheck` without warnings, compliance with
`gofmt -s`, proper error handling (errors MUST be checked and handled, not
ignored), appropriate use of interfaces (accept interfaces, return structs),
idiomatic naming conventions, and standard package organization. Any deviation
MUST be justified with an inline comment referencing the specific reason.

### IV. Event-Driven Architecture

The system MUST use event-driven architecture with Kafka as the core event
backbone. Components MUST communicate through events, not direct synchronous
calls. Event schemas MUST be versioned and backwards-compatible. Event
sourcing patterns SHOULD be preferred where state persistence is required.
Each event consumer MUST be independently testable using recorded or simulated
event streams.

### V. Pathfinding & Autonomy

Pathfinding algorithms MUST be implemented as self-contained modules with
well-defined interfaces, enabling substitution of different algorithms.
Autonomy mechanisms MUST follow a layered approach: reactive
(sensor → immediate action) at the lowest level, deliberative
(planning/scheduling) above, with event-driven coordination spanning all
layers. Each autonomy layer MUST be independently testable. Algorithm
selection MUST be documented with trade-off rationale.

## Technology Stack & Infrastructure

- **Language**: Go 1.26+ with standard toolchain
  (`go build`, `go test`, `go vet`)
- **Event Streaming**: Apache Kafka (event backbone for all inter-component
  communication)
- **Testing**: Go standard `testing` package with table-driven tests;
  `go test -race` is required
- **Static Analysis**: `gofmt -s`, `go vet`, `staticcheck` — all MUST pass
  before commit
- **Pathfinding**: Pure Go implementations; no external graph library
  dependencies
- **Dependency Management**: Go modules (`go.mod` / `go.sum`);
  minimal external dependencies

## Development Workflow & Quality Gates

- All features MUST follow the spec → plan → tasks → implement → review flow
- Each feature slice MUST result in a working, testable increment
- Before committing: `gofmt -s .`, `go vet ./...`, and
  `staticcheck ./...` MUST pass
- Tests MUST be written as table-driven tests following Go conventions;
  test coverage MUST not regress
- Every PR MUST reference which constitution principles it addresses
- Agentic code review is REQUIRED before merging any PR
- Complexity MUST be justified — if a simpler alternative exists, document
  why it was rejected

## Governance

This constitution supersedes all other project practices and guidelines.
Amendments MUST:

1. Be documented with a clear rationale
2. Be reviewed and approved through the standard review process
3. Update the constitution version according to MAJOR.MINOR.PATCH:
   - MAJOR: Backward-incompatible governance or principle changes
   - MINOR: New principles or materially expanded guidance
   - PATCH: Clarifications, wording fixes, non-semantic refinements
4. Include a migration plan for affected processes

All PRs and code reviews MUST verify compliance with this constitution.
The governance section itself is subject to the same amendment process.

**Version**: 1.0.0 | **Ratified**: 2026-06-13 | **Last Amended**: 2026-06-13
