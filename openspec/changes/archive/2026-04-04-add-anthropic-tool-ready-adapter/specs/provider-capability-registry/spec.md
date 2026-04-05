## ADDED Requirements

### Requirement: Anthropic tool-ready descriptors
Holy Code SHALL register explicit Anthropic capability descriptors for the model set supported by the first real-provider runtime slice.

#### Scenario: Runtime resolves an Anthropic model
- **WHEN** a headless session is configured to use an Anthropic model supported by this phase
- **THEN** the registry returns a provider-scoped descriptor for that model
- **AND** the descriptor explicitly declares support for streamed text output, tool calls, and tool-result continuation

### Requirement: Anthropic capability-gated execution
The runtime SHALL gate Anthropic-specific execution modes on the resolved Anthropic capability descriptor rather than assuming that every Anthropic model supports the full tool-ready flow.

#### Scenario: Selected Anthropic model lacks required capability
- **WHEN** the selected Anthropic model descriptor does not support a feature required by the active headless mode
- **THEN** the runtime fails clearly or disables that mode according to documented behavior
- **AND** it does not silently advertise unsupported Anthropic behavior as available
