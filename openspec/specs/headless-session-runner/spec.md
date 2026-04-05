# Headless Session Runner

## Purpose

Define the required behavior for running a single non-interactive Holy Code session through the Go runtime on top of the shared provider-neutral inference contract.

## Requirements

### Requirement: Headless session execution
The Go rewrite SHALL provide a headless execution path that accepts a single user prompt, constructs a Holy Code session request through the provider-neutral inference contract, and runs the session to completion without requiring the interactive TUI.

#### Scenario: Run a one-shot prompt
- **WHEN** a user invokes the Go CLI in headless mode with a prompt
- **THEN** the runtime starts a single session turn
- **AND** the user prompt is included in the request context
- **AND** provider selection occurs through the shared inference abstraction rather than a provider-specific query path
- **AND** the session runs until the assistant emits a final completion or an unrecoverable error occurs

### Requirement: Streaming assistant output
The headless execution path SHALL stream assistant output incrementally rather than buffering the entire response until the end of the turn.

#### Scenario: Text is streamed during generation
- **WHEN** a supported provider returns assistant content in multiple streamed events
- **THEN** the headless runtime surfaces output incrementally in arrival order through the shared turn-event model
- **AND** the final output preserves the same assembled assistant content as the completed stream

### Requirement: Tool-call continuation
The headless execution path SHALL detect supported tool-call events, execute them through the Go tool runtime, and continue the same turn with tool results until a final assistant response is produced.

#### Scenario: Assistant requests a supported tool
- **WHEN** the streamed assistant response includes a supported tool-call event from any supported provider adapter
- **THEN** the runtime executes the requested tool through the shared tool interface
- **AND** the tool result is added back into the conversation state through the provider-neutral continuation contract
- **AND** the query loop continues instead of terminating after the first tool call

### Requirement: Headless failure reporting
The headless execution path SHALL surface configuration, provider, and tool-execution failures as explicit errors instead of silently truncating the session.

#### Scenario: Provider request fails before completion
- **WHEN** the runtime encounters an unrecoverable provider or tool error during a headless session
- **THEN** the command exits with a non-success status
- **AND** the reported error identifies the failure class closely enough for debugging and retry decisions
