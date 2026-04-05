# Multi-Provider Inference Core

## Purpose

Define the shared provider-neutral inference contract that allows Holy Code runtime logic to work across Anthropic, OpenAI Responses, and OpenAI-compatible adapters.

## Requirements

### Requirement: Provider-neutral turn contract
Holy Code SHALL define an internal provider-neutral turn contract for inference requests, streamed turn events, tool-call events, tool-result continuation, usage accounting, and terminal stop conditions.

#### Scenario: Query runtime consumes a provider stream
- **WHEN** the query runtime executes a turn through any supported provider adapter
- **THEN** it receives events through the shared internal turn contract
- **AND** it does not need provider-specific parsing logic in the core query loop

### Requirement: Unified streaming event model
The provider-neutral turn contract SHALL represent incremental assistant text, tool-call lifecycle events, completion, and provider failures through a unified event model.

#### Scenario: Different providers stream the same logical turn
- **WHEN** Anthropic native streaming and OpenAI Responses streaming each produce a tool-using assistant turn
- **THEN** both streams are translated into the same internal event categories
- **AND** the runtime can process them with the same turn orchestration logic

### Requirement: Provider metadata preservation
The shared turn contract SHALL allow provider-specific metadata to remain available without making that metadata part of the core runtime control flow.

#### Scenario: Adapter surfaces provider-specific details
- **WHEN** a provider emits metadata that has diagnostic or advanced-feature value but no direct effect on the shared runtime contract
- **THEN** the adapter preserves it as provider-scoped metadata
- **AND** the runtime can ignore it without losing core correctness
