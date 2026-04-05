## MODIFIED Requirements

### Requirement: Executable parity fixtures
The Go rewrite SHALL include executable fixtures or acceptance cases that encode the expected behaviors for the provider-neutral inference core, headless session runner, and initial toolchain.

#### Scenario: Running the parity suite
- **WHEN** developers run the rewrite's conformance-oriented test suite
- **THEN** the suite exercises the shared turn-event model and supported provider adapters against deterministic expectations
- **AND** failures identify which required behavior regressed

### Requirement: Spec-derived acceptance coverage
The first conformance baseline SHALL be derived from the repository's written specs rather than from copied TypeScript implementation logic.

#### Scenario: Adding an acceptance case
- **WHEN** a developer adds or updates a parity fixture for the initial rewrite foundation
- **THEN** the expected behavior is traceable to the relevant repository spec or approved change artifact
- **AND** the fixture does not require embedding proprietary TypeScript source

### Requirement: Change-gated regression protection
Changes to the inference core, provider adapters, headless runner, or initial toolchain SHALL be validated against the conformance baseline before they are considered complete.

#### Scenario: Modifying the initial runtime foundation
- **WHEN** a change touches the shared inference contract or a supported first-phase provider adapter
- **THEN** the associated test or fixture suite is run during verification
- **AND** an unexplained regression blocks the change from being declared complete
