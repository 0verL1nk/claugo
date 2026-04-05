# anthropic-provider-adapter Specification

## Purpose
TBD - created by archiving change add-anthropic-tool-ready-adapter. Update Purpose after archive.
## Requirements
### Requirement: Anthropic request translation
Holy Code SHALL provide an Anthropic adapter that translates the shared turn request contract into Anthropic Messages API requests without exposing Anthropic-native payload semantics to the core query loop.

#### Scenario: Headless prompt is sent through Anthropic
- **WHEN** the runtime executes a headless turn with provider `anthropic`
- **THEN** the adapter constructs the Anthropic request from the shared turn request
- **AND** `internal/query` does not need Anthropic-specific request-building logic

### Requirement: Anthropic stream maps to shared events
The Anthropic adapter SHALL decode streamed Anthropic responses into the shared turn-event model for text output, tool invocation, completion, usage, and provider failure.

#### Scenario: Anthropic streams assistant output
- **WHEN** Anthropic returns a streamed assistant response
- **THEN** the adapter emits shared text delta and terminal events in arrival order
- **AND** the query runtime can process the stream without Anthropic-specific parsing branches

### Requirement: Anthropic tool-ready continuation
The Anthropic adapter SHALL support a tool-ready headless flow in which tool calls are surfaced through the shared tool-call contract and tool results are translated back into Anthropic continuation messages for the same turn.

#### Scenario: Anthropic requests a supported tool
- **WHEN** an Anthropic response requests a supported local tool during a headless turn
- **THEN** the adapter emits a shared tool-call event that the runtime can execute
- **AND** the resulting shared tool result can be translated back into the continued Anthropic conversation
- **AND** the turn resumes until Anthropic returns a final assistant completion or failure

### Requirement: Anthropic error normalization
The Anthropic adapter SHALL normalize authentication, rate-limit, schema, and protocol failures into the shared runtime error model while preserving provider-scoped diagnostic context.

#### Scenario: Anthropic request fails
- **WHEN** Anthropic returns an error response or invalid protocol sequence
- **THEN** the adapter maps the failure into the shared error categories
- **AND** provider-specific response details remain available for debugging without changing query-loop control flow

