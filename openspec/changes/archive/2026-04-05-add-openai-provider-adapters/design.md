## Context

The Go rewrite already has a provider-neutral inference contract, a headless query loop, a shared local tool runtime, and a real Anthropic adapter. That foundation proves the architecture, but the stated goal is broader: the same runtime spine must also support OpenAI Responses and OpenAI-compatible backends without turning `internal/query`, `internal/api`, or `cmd/holy` into provider-specific orchestration layers.

This change is broader than the Anthropic integration because the two OpenAI families are similar in branding but materially different in protocol guarantees. OpenAI Responses is a first-party agentic API with richer event and tool semantics. OpenAI-compatible backends often expose only a chat-style subset and vary widely in tool and continuation support. The design therefore needs to preserve a single shared runtime contract while making capability limits explicit.

## Goals / Non-Goals

**Goals:**
- Add a real `openai-responses` adapter under `internal/providers/openairesponses`.
- Add a real `openai-compatible` adapter under `internal/providers/openaicompat`.
- Extend provider config, defaults, and validation for both OpenAI provider families through existing Holy Code settings paths and CLI flags.
- Register explicit model capability descriptors so runtime behavior depends on descriptors rather than provider-name assumptions.
- Keep tool execution and turn orchestration in the shared runtime, with provider-specific request and stream handling isolated inside adapters.
- Add deterministic adapter and runtime tests that prove both adapters map into the shared event model and fail clearly when capabilities are absent.

**Non-Goals:**
- Rework the current headless slice into the full TUI, slash-command, MCP, or bridge architecture.
- Support every OpenAI or OpenAI-compatible feature surface, including advanced hosted tools, background jobs, or provider-specific reasoning metadata.
- Promise full Responses semantics for arbitrary OpenAI-compatible backends.
- Introduce a generic HTTP abstraction layer before there is a demonstrated reuse need beyond the two new adapters.
- Expand the local tool surface beyond the current `Read`, `Edit`, and `Bash` set.

## Decisions

### Decision: Keep OpenAI Responses and OpenAI-compatible as separate adapter families

`internal/providers/openairesponses` and `internal/providers/openaicompat` will be separate packages with separate descriptors, stream decoders, and error normalization. They will share the `internal/inference` contract, but not a fake equivalence at the provider boundary.

This is preferred over collapsing both into a single `openai` adapter because the repository already documents OpenAI-compatible backends as a lower-guarantee family. Preserving separate adapters keeps capability declarations honest and prevents a broad class of false assumptions.

### Decision: Make capability descriptors drive tool advertisement and continuation

The runtime must resolve the selected model descriptor before deciding whether tool definitions should be advertised and whether tool-result continuation is legal. If a descriptor lacks tool-calling support, the headless path must either run as text-only or fail explicitly when the requested mode requires tools.

This is preferred over the current implicit behavior where the headless loop always sends tool definitions. Without capability-driven request shaping, text-only OpenAI-compatible backends would be rejected before they could serve the minimum supported contract.

### Decision: Define OpenAI-compatible as a minimum contract, not a promise of parity

The OpenAI-compatible adapter will guarantee request translation, streamed or buffered text output, and normalized provider errors for a documented chat-style baseline. Tool calling and same-turn continuation are optional capabilities that may be enabled only for descriptors that explicitly declare them.

This is preferred over assuming all OpenAI-compatible servers implement Responses-like event semantics. The latter would make the adapter brittle and would mislead the runtime into advertising unsupported features.

### Decision: Keep provider-native transport logic inside each adapter

The adapters will own request payload construction, response decoding, event assembly, and provider-specific error mapping. Shared runtime packages may grow provider-neutral helpers, but they must not absorb OpenAI-specific wire semantics.

This is preferred over pushing protocol branches into `internal/api` or `internal/query`, which would directly undermine the provider-neutral architecture already proven by the Anthropic change.

### Decision: Add conformance tests for both happy-path and capability-denied flows

The deterministic baseline must verify not only successful event mapping, but also capability-gated behavior for unsupported tools or continuation on OpenAI-compatible backends. This includes adapter-level tests and query/runtime integration coverage where capability decisions affect shared control flow.

This is preferred over validating only success paths because the main architectural risk in this change is incorrect over-advertising of backend capabilities.

## Risks / Trade-offs

- [Risk] OpenAI Responses streaming may expose lifecycle details that do not map cleanly onto the current shared event model. -> Mitigation: extend `internal/inference` only with provider-neutral fields that can also serve other adapters.
- [Risk] OpenAI-compatible vendors vary widely, making one adapter contract too optimistic or too weak. -> Mitigation: document a strict minimum contract and require explicit descriptor-level opt-in for advanced features such as tool continuation.
- [Risk] Capability-driven request shaping may require small shared-runtime refactors in `internal/api` or `internal/query`. -> Mitigation: keep those refactors limited to descriptor lookup and request preparation, not provider-native protocol formatting.
- [Risk] Config behavior can become inconsistent across providers as new env vars and base URLs are introduced. -> Mitigation: keep one shared config path and define per-provider defaults and validation in `internal/core`.
- [Risk] Adapter tests may overfit one mock shape and miss integration drift. -> Mitigation: include both adapter-local tests and shared runtime conformance tests that assert provider-neutral event behavior.

## Migration Plan

1. Add OpenAI-family config plumbing, provider registration, and capability descriptors while preserving the existing fake and Anthropic paths.
2. Implement the OpenAI Responses adapter and validate streamed text, tool calls, continuation, and error normalization through mocked transport.
3. Implement the OpenAI-compatible adapter with a text-first minimum contract plus explicit capability-gated advanced behavior.
4. Update shared runtime logic so tool advertisement and continuation are descriptor-driven rather than assumed for every backend.
5. Add conformance tests across both adapters and re-run the full headless test suite.

Rollback is straightforward: unregister the new adapters and keep the runtime limited to Anthropic and fake provider paths.

## Open Questions

- Which OpenAI Responses model IDs should be the initial supported descriptor set for the first implementation pass?
- Should the OpenAI-compatible baseline require streaming text, or allow buffered text completion when streaming is unavailable?
- Should `openai-compatible` require an explicit `base_url`, or should the config layer allow a provider-specific default only for testing and local compatibility shims?
