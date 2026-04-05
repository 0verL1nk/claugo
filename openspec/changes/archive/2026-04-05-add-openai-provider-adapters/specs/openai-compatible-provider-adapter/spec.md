## ADDED Requirements

### Requirement: OpenAI-compatible minimum request contract
Holy Code SHALL provide an OpenAI-compatible adapter that translates the shared turn request contract into a documented chat-style minimum contract instead of assuming full OpenAI Responses semantics.

#### Scenario: Headless prompt is sent through an OpenAI-compatible backend
- **WHEN** the runtime executes a headless turn with provider `openai-compatible`
- **THEN** the adapter constructs a provider-native request from the shared turn request
- **AND** the adapter does not require Responses-only protocol features to satisfy the minimum contract

### Requirement: OpenAI-compatible capability-scoped behavior
The OpenAI-compatible adapter SHALL expose only the capabilities explicitly declared by the resolved descriptor and SHALL fail clearly or degrade explicitly when advanced features are unavailable.

#### Scenario: Backend lacks tool continuation support
- **WHEN** the selected OpenAI-compatible descriptor does not support tool calls or conversation continuation
- **THEN** the adapter and runtime expose only the supported text-oriented behavior
- **AND** unsupported advanced behavior is rejected explicitly instead of being advertised as available

### Requirement: OpenAI-compatible text output mapping
The OpenAI-compatible adapter SHALL map streamed or completed assistant text into the shared turn-event model used by the headless runtime.

#### Scenario: Compatible backend returns text output
- **WHEN** an OpenAI-compatible backend returns assistant text for a headless turn
- **THEN** the adapter emits shared text and terminal events that preserve the final assistant content
- **AND** the query runtime processes the result through the same provider-neutral control flow used for other adapters

### Requirement: OpenAI-compatible error normalization
The OpenAI-compatible adapter SHALL normalize provider-specific authentication, rate-limit, schema, and protocol failures into the shared runtime error model while preserving provider-scoped debugging context.

#### Scenario: Compatible backend request fails
- **WHEN** an OpenAI-compatible backend returns a backend-specific failure
- **THEN** the adapter maps the failure into the shared runtime error categories
- **AND** provider-specific response details remain available for diagnostics without changing shared runtime semantics
