## ADDED Requirements

### Requirement: Explicit capability descriptors
Holy Code SHALL describe provider and model behavior through explicit capability descriptors rather than implicit assumptions in the runtime.

#### Scenario: Runtime checks backend support
- **WHEN** the runtime needs to know whether a selected backend supports a feature such as tool calling, tool-argument streaming, structured output, or stateful response continuation
- **THEN** it reads that information from the provider or model capability descriptor
- **AND** it does not infer support solely from provider identity

### Requirement: Capability-gated behavior
The query runtime SHALL gate provider-dependent behavior on capabilities instead of hardcoding provider-name branches throughout the orchestration flow.

#### Scenario: Backend lacks a required feature
- **WHEN** a selected provider or model does not support a feature required by the current execution mode
- **THEN** the runtime fails clearly or degrades according to documented behavior
- **AND** it does not silently pretend the capability exists

### Requirement: Model descriptor separation
Holy Code SHALL separate model identity from model capability metadata so provider-specific model IDs do not become core runtime constants.

#### Scenario: Selecting a model
- **WHEN** a user or config selects a model for a provider
- **THEN** the runtime resolves a provider-scoped model descriptor
- **AND** the core runtime depends on the descriptor's capabilities rather than on a globally hardcoded model ID list
