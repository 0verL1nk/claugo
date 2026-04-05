## ADDED Requirements

### Requirement: Executable parity fixtures
The Go rewrite SHALL include executable fixtures or acceptance cases that encode the expected behaviors for the headless session runner, the shared provider-neutral runtime contract, and the initial toolchain.

#### Scenario: Running the parity suite
- **WHEN** developers run the rewrite's conformance-oriented test suite
- **THEN** the suite exercises the initial headless session flow, shared event semantics, and supported tools against deterministic expectations
- **AND** failures identify which required behavior regressed

### Requirement: Spec-derived acceptance coverage
The first conformance baseline SHALL be derived from the repository's written specs rather than from copied TypeScript implementation logic.

#### Scenario: Adding an acceptance case
- **WHEN** a developer adds or updates a parity fixture for the first rewrite slice
- **THEN** the expected behavior is traceable to the relevant repository spec or approved change artifact
- **AND** the fixture does not require embedding proprietary TypeScript source

### Requirement: Change-gated regression protection
Changes to the headless runner, shared runtime foundation, or initial toolchain SHALL be validated against the conformance baseline before they are considered complete.

#### Scenario: Modifying the initial runtime slice
- **WHEN** a change touches the headless session runner, the shared event/runtime path, or the supported first-phase tools
- **THEN** the associated test or fixture suite is run during verification
- **AND** an unexplained regression blocks the change from being declared complete
