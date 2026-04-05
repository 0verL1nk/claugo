## Why

The Go rewrite already has a provider-neutral inference core and a runnable headless query loop, but it still depends on a fake provider for end-to-end execution. The next step is to prove that the shared runtime contract works against a real backend before expanding to additional providers.

## What Changes

- Add a real Anthropic adapter that translates the shared turn request into Anthropic Messages API requests.
- Allow Anthropic provider settings to be sourced from Holy Code settings files, with command-line flags overriding configured values.
- Decode Anthropic streaming responses into the existing provider-neutral turn-event model.
- Support a tool-ready headless flow for Anthropic, including tool-call detection, tool execution, and tool-result continuation.
- Normalize Anthropic authentication, rate-limit, schema, and protocol failures into the shared runtime error model.
- Register Anthropic model capability descriptors needed by the first real-provider runtime slice.
- Add deterministic adapter and runtime tests using mocked Anthropic transport, plus a manual live verification path when credentials are available.

## Capabilities

### New Capabilities
- `anthropic-provider-adapter`: Defines the Anthropic-specific request translation, streaming event mapping, tool continuation, and error normalization required by the first real provider integration.

### Modified Capabilities
- `provider-capability-registry`: Define the Anthropic model capability descriptors required for the initial tool-ready adapter path.
- `rewrite-conformance-baseline`: Extend the baseline to cover adapter-level stream mapping and tool continuation using mocked Anthropic transport.

## Impact

- Affected code: `internal/providers/anthropic`, `internal/inference`, `internal/api`, `internal/query`, `cmd/holy`
- Affected runtime behavior: headless execution can run against Anthropic instead of only the fake provider
- Dependencies: stdlib `net/http`, `encoding/json`, SSE decoding inside the adapter, Anthropic API key and base URL configuration via settings files and/or flags
