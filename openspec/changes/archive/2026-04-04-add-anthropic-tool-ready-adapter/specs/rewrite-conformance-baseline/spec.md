## ADDED Requirements

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
