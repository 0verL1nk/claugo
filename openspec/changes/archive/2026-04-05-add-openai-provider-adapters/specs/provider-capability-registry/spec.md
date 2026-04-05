## ADDED Requirements

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
