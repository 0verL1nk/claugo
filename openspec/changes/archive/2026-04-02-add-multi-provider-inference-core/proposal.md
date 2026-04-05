## Why

The current rewrite direction still assumes an Anthropic-shaped runtime in the core query path, including provider-specific request and streaming semantics. That would make Holy Code harder to extend to OpenAI Responses API and OpenAI-compatible backends, and would force a second architectural rewrite once multi-provider support becomes necessary.

## What Changes

- Introduce a provider-neutral inference layer for Holy Code instead of binding the core runtime directly to Anthropic message semantics.
- Define shared request, event, tool-call, tool-result, usage, and stop-reason contracts that can represent Anthropic native streaming, OpenAI Responses streaming, and OpenAI-compatible chat-style responses.
- Add an explicit provider capability model so runtime behavior can adapt to backend differences instead of assuming one universal feature set.
- Define three first-class provider adapters: `anthropic`, `openai-responses`, and `openai-compatible`.
- Make the headless runtime and future query loop changes depend on this provider abstraction rather than embedding provider-specific logic in `internal/api` or `internal/query`.
- **BREAKING** Update the rewrite plan so the initial implementation order becomes provider abstraction first, headless runtime second.

## Capabilities

### New Capabilities
- `multi-provider-inference-core`: A provider-neutral inference contract for request building, streaming events, tool-call round-trips, usage accounting, and stop conditions.
- `provider-capability-registry`: A capability model and model-descriptor registry that describes backend features such as streaming shape, tool calling, structured output, reasoning visibility, and conversation state handling.
- `provider-adapter-packages`: Dedicated adapters for Anthropic native API, OpenAI Responses API, and OpenAI-compatible APIs, all targeting the same internal runtime contract.

### Modified Capabilities
- `headless-session-runner`: Change the planned runtime foundation so headless execution depends on the new provider-neutral inference contract rather than an Anthropic-specific transport design.
- `rewrite-conformance-baseline`: Expand the parity baseline so it validates provider-neutral event semantics and adapter conformance, not just a single backend path.

## Impact

- Changes the architectural foundation for `internal/api`, `internal/query`, and related runtime packages in the Go rewrite.
- Prevents Anthropic-specific protocol details from leaking into the long-term core design.
- Creates a clear path to support Anthropic API, OpenAI Responses API, and OpenAI-compatible APIs under one runtime.
- Reorders follow-on changes: the current headless bootstrap change should be revised to depend on this abstraction instead of implementing a provider-specific API layer first.
