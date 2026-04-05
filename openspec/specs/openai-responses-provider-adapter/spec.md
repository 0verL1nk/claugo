# OpenAI Responses Provider Adapter

## Purpose

Define the OpenAI Responses adapter behavior required for the Go rewrite's provider-neutral headless runtime.

## Requirements

### Requirement: OpenAI Responses request translation
Holy Code SHALL provide an OpenAI Responses adapter that translates the shared turn request contract into OpenAI Responses API requests without exposing OpenAI-native payload semantics to the core query loop.

#### Scenario: Headless prompt is sent through OpenAI Responses
- **WHEN** the runtime executes a headless turn with provider `openai-responses`
- **THEN** the adapter constructs the OpenAI Responses request from the shared turn request
- **AND** `internal/query` does not need OpenAI-specific request-building logic

### Requirement: OpenAI Responses stream maps to shared events
The OpenAI Responses adapter SHALL decode streamed OpenAI responses into the shared turn-event model for text output, tool invocation, completion, usage, and provider failure.

#### Scenario: OpenAI Responses streams assistant output
- **WHEN** OpenAI Responses returns streamed assistant output for a headless turn
- **THEN** the adapter emits shared text delta and terminal events in arrival order
- **AND** the query runtime can process the stream without OpenAI-specific parsing branches

### Requirement: OpenAI Responses tool-ready continuation
The OpenAI Responses adapter SHALL support a tool-ready headless flow in which tool calls are surfaced through the shared tool-call contract and tool results are translated back into the provider-native continuation format for the same turn.

#### Scenario: OpenAI Responses requests a supported tool
- **WHEN** an OpenAI Responses turn requests a supported local tool
- **THEN** the adapter emits a shared tool-call event that the runtime can execute
- **AND** the resulting shared tool result can be translated back into the continued OpenAI Responses conversation
- **AND** the turn resumes until a final assistant completion or provider failure is returned

### Requirement: OpenAI Responses error normalization
The OpenAI Responses adapter SHALL normalize authentication, rate-limit, schema, and protocol failures into the shared runtime error model while preserving provider-scoped diagnostic context.

#### Scenario: OpenAI Responses request fails
- **WHEN** OpenAI Responses returns an error response or invalid protocol sequence
- **THEN** the adapter maps the failure into the shared runtime error categories
- **AND** provider-specific response details remain available for debugging without changing query-loop control flow
