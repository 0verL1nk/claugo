## 1. Anthropic Runtime Wiring

- [x] 1.1 Add Anthropic-specific runtime configuration for provider selection, API key, and base URL without breaking the fake-provider path.
- [x] 1.2 Register Anthropic model descriptors and capability metadata in the shared provider registry.
- [x] 1.3 Update runtime selection so `--provider anthropic` resolves the Anthropic adapter through existing provider-neutral interfaces.

## 2. Anthropic Adapter Implementation

- [x] 2.1 Implement Anthropic request translation from the shared turn request contract to Anthropic Messages API payloads.
- [x] 2.2 Implement streamed Anthropic response decoding into shared text, tool-call, completion, usage, and failure events.
- [x] 2.3 Implement Anthropic tool-result continuation so shared tool results resume the same turn without query-loop protocol branching.
- [x] 2.4 Normalize Anthropic authentication, rate-limit, schema, and protocol failures into the shared runtime error model.

## 3. Verification Coverage

- [x] 3.1 Add mocked transport tests for Anthropic request construction, stream-event mapping, and error normalization.
- [x] 3.2 Add end-to-end runtime tests for Anthropic text streaming, tool execution, and tool-result continuation through the headless query loop.
- [x] 3.3 Document and run a manual live smoke-verification path for `cmd/holy --provider anthropic` when credentials are available.
