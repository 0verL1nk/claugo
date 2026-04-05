# Provider Capability Registry

## Purpose

Define how Holy Code represents provider and model capabilities so runtime behavior depends on explicit descriptors instead of hardcoded backend assumptions.
## Requirements
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

### Requirement: OpenAI Responses capability descriptors
Holy Code SHALL register explicit capability descriptors for the OpenAI Responses model set supported by the first OpenAI provider integration slice.

#### Scenario: Runtime resolves an OpenAI Responses model
- **WHEN** a headless session is configured to use a supported OpenAI Responses model
- **THEN** the registry returns a provider-scoped descriptor for that model
- **AND** the descriptor explicitly declares support for streamed text output and any supported advanced features such as tool calls or conversation continuation

### Requirement: OpenAI-compatible minimum-contract descriptors
Holy Code SHALL register explicit capability descriptors for OpenAI-compatible model selection so different compatibility tiers can be represented without hardcoding backend assumptions into the runtime.

#### Scenario: Runtime resolves an OpenAI-compatible backend
- **WHEN** a headless session is configured to use an OpenAI-compatible model or endpoint
- **THEN** the registry returns a provider-scoped descriptor that reflects the actual supported compatibility tier
- **AND** the runtime reads that descriptor rather than inferring feature support from provider identity alone

### Requirement: Capability-driven tool advertisement
The runtime SHALL advertise tool definitions and enable tool-result continuation only when the resolved descriptor declares the required capabilities.

#### Scenario: Selected backend is text-only
- **WHEN** the resolved descriptor does not support tool calls for the active provider and model
- **THEN** the runtime does not advertise unsupported tool usage as available
- **AND** the session either runs in a supported text-only mode or fails clearly if the requested mode requires tools
