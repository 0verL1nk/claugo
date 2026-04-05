## ADDED Requirements

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
