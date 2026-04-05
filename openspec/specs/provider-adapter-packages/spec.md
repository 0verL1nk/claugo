# Provider Adapter Packages

## Purpose

Define the required adapter boundaries and minimum responsibilities for each supported provider family in the Go rewrite.

## Requirements

### Requirement: Three first-class adapter families
Holy Code SHALL define first-class adapter boundaries for `anthropic`, `openai-responses`, and `openai-compatible`.

#### Scenario: Provider selection
- **WHEN** the runtime is configured to use one of the supported provider families
- **THEN** the corresponding adapter is responsible for request translation, stream decoding, and provider-specific error mapping
- **AND** all adapters target the same internal turn contract

### Requirement: OpenAI-compatible minimum contract
Holy Code SHALL treat OpenAI-compatible backends as a separate adapter family with an explicit minimum supported contract rather than assuming full OpenAI Responses semantics.

#### Scenario: OpenAI-compatible backend lacks Responses features
- **WHEN** an OpenAI-compatible backend supports only a chat-style subset and not full Responses semantics
- **THEN** the adapter exposes only the capabilities actually supported
- **AND** the runtime does not advertise unsupported features as available

### Requirement: Adapter-scoped provider errors
Each provider adapter SHALL normalize provider-specific failures into the shared runtime error model while preserving provider context for debugging.

#### Scenario: Provider returns a backend-specific failure
- **WHEN** a provider returns an authentication, schema, rate-limit, or protocol error in its own format
- **THEN** the adapter maps it into the shared runtime error categories
- **AND** the original provider context remains available for diagnostics
