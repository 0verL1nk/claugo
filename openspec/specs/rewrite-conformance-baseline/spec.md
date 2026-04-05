# Rewrite Conformance Baseline

## Purpose

Define the executable verification baseline that protects the first Go rewrite foundation from regressions in shared runtime behavior, provider integration, and the initial toolchain.
## Requirements
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

### Requirement: Mocked Anthropic adapter conformance
The rewrite conformance baseline SHALL include deterministic tests that validate Anthropic request translation, streamed event mapping, tool-call continuation, and error normalization through mocked transport fixtures.

#### Scenario: Running Anthropic adapter conformance tests
- **WHEN** developers run the conformance-oriented test suite for the Anthropic adapter change
- **THEN** the suite exercises Anthropic-specific stream and continuation behavior through mocked transport responses
- **AND** failures identify whether the regression is in request translation, event mapping, tool continuation, or error handling

### Requirement: Real-provider headless smoke path
The rewrite SHALL document a manual smoke-verification path for running the headless CLI against a real Anthropic backend once credentials are configured.

#### Scenario: Manually verifying the first real provider
- **WHEN** a developer provides valid Anthropic credentials and runs the headless CLI with provider `anthropic`
- **THEN** the documented verification path demonstrates streamed output and at least one successful end-to-end turn
- **AND** the live smoke step complements but does not replace deterministic mocked transport tests

### Requirement: Mocked OpenAI Responses adapter conformance
The rewrite conformance baseline SHALL include deterministic tests that validate OpenAI Responses request translation, streamed event mapping, tool-call continuation, and error normalization through mocked transport fixtures.

#### Scenario: Running OpenAI Responses adapter conformance tests
- **WHEN** developers run the conformance-oriented test suite for the OpenAI Responses adapter
- **THEN** the suite exercises OpenAI Responses-specific request and stream behavior through mocked transport responses
- **AND** failures identify whether the regression is in request translation, event mapping, tool continuation, or error handling

### Requirement: Mocked OpenAI-compatible adapter conformance
The rewrite conformance baseline SHALL include deterministic tests that validate OpenAI-compatible minimum-contract behavior, capability-gated unsupported-feature handling, text output mapping, and error normalization through mocked transport fixtures.

#### Scenario: Running OpenAI-compatible adapter conformance tests
- **WHEN** developers run the conformance-oriented test suite for the OpenAI-compatible adapter
- **THEN** the suite exercises both supported text behavior and unsupported advanced-feature behavior through mocked transport responses
- **AND** failures identify whether the regression is in request translation, capability gating, event mapping, or error handling
