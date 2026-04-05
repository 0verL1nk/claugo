## 1. Shared Runtime And Config Preparation

- [x] 1.1 Extend `internal/core` provider config defaults, env lookup, settings-file handling, and validation for `openai-responses` and `openai-compatible`.
- [x] 1.2 Add provider and model descriptors for the new OpenAI adapter families in the shared inference registry.
- [x] 1.3 Update shared runtime and query request preparation so tool advertisement and continuation are driven by resolved capabilities instead of being assumed for every backend.

## 2. OpenAI Responses Adapter

- [x] 2.1 Implement `internal/providers/openairesponses` request translation from the shared turn contract into OpenAI Responses requests.
- [x] 2.2 Implement OpenAI Responses stream decoding, usage mapping, and provider error normalization into the shared event model.
- [x] 2.3 Implement tool-call and tool-result continuation for OpenAI Responses and add deterministic adapter and query/runtime tests.

## 3. OpenAI-Compatible Adapter

- [x] 3.1 Implement `internal/providers/openaicompat` request and response translation for the documented chat-style minimum contract.
- [x] 3.2 Implement capability-scoped behavior for text-only and advanced-feature-compatible descriptors, including explicit unsupported-feature failures.
- [x] 3.3 Add deterministic adapter and runtime tests covering text output mapping, capability gating, and normalized provider errors.

## 4. Conformance And Verification

- [x] 4.1 Add shared provider conformance coverage for Anthropic, OpenAI Responses, and OpenAI-compatible adapters against the shared event model.
- [x] 4.2 Run targeted provider test suites and the full `go test ./...` verification pass.
- [x] 4.3 Document provider-specific verification notes needed to exercise the new adapters through `cmd/holy` with existing settings-file and flag wiring.
