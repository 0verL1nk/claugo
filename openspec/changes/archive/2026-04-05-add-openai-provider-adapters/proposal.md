## Why

The Go rewrite has already proven the provider-neutral runtime contract with a fake provider and a real Anthropic path, but the multi-provider foundation is still incomplete. The next step is to validate that the same runtime spine can support both OpenAI Responses and OpenAI-compatible backends without leaking provider-specific protocol logic into the core query loop.

## What Changes

- Add a real `openai-responses` adapter that translates the shared turn request into OpenAI Responses API requests and maps streamed response items/events into the shared event model.
- Add a real `openai-compatible` adapter that targets a documented minimum chat-style contract instead of assuming full Responses semantics.
- Extend runtime configuration so provider selection, API key lookup, base URL handling, and model validation work cleanly for the OpenAI provider families through existing Holy Code settings paths and CLI flags.
- Register explicit OpenAI Responses and OpenAI-compatible capability descriptors in the shared provider registry.
- Add deterministic adapter and runtime conformance coverage for both OpenAI provider families, including capability-gated behavior and normalized provider error mapping.
- Preserve the existing provider-neutral query loop so tool execution and turn continuation remain shared runtime behavior rather than adapter-specific orchestration branches.

## Capabilities

### New Capabilities
- `openai-responses-provider-adapter`: Defines OpenAI Responses request translation, streamed event mapping, tool-call handling, continuation behavior, and provider error normalization for the Go rewrite.
- `openai-compatible-provider-adapter`: Defines the minimum supported OpenAI-compatible adapter contract, including capability limits, chat-style request mapping, and explicit unsupported-feature behavior.

### Modified Capabilities
- `provider-capability-registry`: Add explicit capability descriptors for OpenAI Responses and OpenAI-compatible model selection and capability-gated runtime behavior.
- `rewrite-conformance-baseline`: Extend deterministic conformance coverage to include OpenAI Responses and OpenAI-compatible adapter behavior against the shared event model.

## Impact

- Affected code: `internal/providers/openairesponses`, `internal/providers/openaicompat`, `internal/providers/conformance`, `internal/inference`, `internal/api`, `internal/query`, `internal/core`, `cmd/holy`
- Affected runtime behavior: headless execution can run against OpenAI Responses and OpenAI-compatible backends in addition to Anthropic and the fake provider
- Dependencies: stdlib `net/http`, `encoding/json`, streaming decoders for Responses events, chat-style request/response handling for compatible backends, provider credentials and base URLs sourced through `.holy/settings.json` and CLI flags
