## Context

The current Go rewrite already has a provider-neutral inference contract, a headless query loop, a shared local tool runtime, and a fake provider for deterministic tests. That foundation proves the internal package boundaries, but it does not yet prove that the runtime contract can drive a real provider with streaming text, tool calls, and turn continuation.

Anthropic is the right first real-provider integration because the runtime is already expected to support it as a first-class adapter family, and its Messages API is sufficient to validate the most important end-to-end path: headless prompt -> streamed output -> tool call -> tool result continuation -> final answer.

The main constraint is architectural: this change must not pull Anthropic protocol details into `internal/query`, `internal/api`, or `cmd/holy`. The adapter has to absorb request translation, stream decoding, and provider-specific error handling while continuing to emit only the shared `internal/inference` contract.

## Goals / Non-Goals

**Goals:**
- Add a real Anthropic adapter under `internal/providers/anthropic`.
- Allow the Anthropic headless path to read provider settings from `~/.holy/settings.json` and project `.holy/settings.json`, with flags taking precedence.
- Translate provider-neutral turn requests into Anthropic Messages requests.
- Decode Anthropic streaming events into shared text, tool-call, completion, usage, and failure events.
- Support a tool-ready continuation flow for existing `Read`, `Edit`, and `Bash` tools.
- Register Anthropic capability descriptors for the first supported model set.
- Add deterministic mocked-transport tests plus a manual live verification path.

**Non-Goals:**
- Implement OpenAI Responses or OpenAI-compatible adapters in this change.
- Introduce an Anthropic-shaped control flow into `internal/query` or `internal/api`.
- Cover every Anthropic model, beta header, or advanced feature surface.
- Add a shared provider HTTP abstraction before a real adapter has validated the runtime boundaries.
- Expand the local tool surface beyond the already implemented headless tools.

## Decisions

### Decision: Keep Anthropic protocol handling inside a thin adapter

The Anthropic adapter will own HTTP request construction, headers, streaming decode, event mapping, and provider error normalization. The query loop will continue to consume only provider-neutral turn events.

This is preferred over making `internal/api` or `internal/query` Anthropic-aware because the repository has already committed to a provider-neutral runtime contract. Reintroducing Anthropic-specific orchestration would make later OpenAI adapters materially harder.

### Decision: Reuse Holy Code settings file locations for Anthropic configuration

The headless Anthropic path will read provider settings from the existing Holy Code settings locations, global `~/.holy/settings.json` and project `.holy/settings.json`, then allow command-line flags to override those values per invocation.

This is preferred over inventing a new provider-specific config file because the repository already documents Holy Code settings at those paths, and the user experience should not split basic runtime configuration across incompatible locations.

### Decision: Treat tool continuation as adapter responsibility, not query-loop protocol logic

The query loop should know only that a tool call was requested and that a tool result must continue the same turn. The Anthropic adapter will map prior assistant output and tool results into the provider-native continuation payload required by Anthropic.

This is preferred over building Anthropic continuation payloads in the query loop because continuation message formatting is provider-specific and would leak protocol semantics upward.

### Decision: Add only shared inference fields that generalize across providers

If the current `internal/inference` contract needs more detail to represent streamed tool arguments, usage, or stop reasons correctly, the contract may be extended only with provider-neutral fields. Anthropic-specific debug or protocol details must remain in provider-scoped metadata.

This is preferred over adding Anthropic-only fields because OpenAI Responses and OpenAI-compatible adapters still need to fit the same runtime model.

### Decision: Verify with mocked transport first, live API second

The primary test surface will use `httptest` or equivalent mocked transport to simulate Anthropic streaming, tool-call events, continuation, and errors. A live Anthropic API run is still useful, but it is a manual verification step rather than the core automated proof.

This is preferred over relying on live API tests because deterministic CI-style verification is needed before additional providers are added.

## Risks / Trade-offs

- [Risk] Anthropic streaming events may expose lifecycle details that do not map cleanly onto the current event model. -> Mitigation: extend `internal/inference` only with provider-neutral lifecycle details that other adapters can also use later.
- [Risk] Tool-call argument streaming may tempt the implementation toward Anthropic-shaped state in the query loop. -> Mitigation: accumulate provider-specific argument state entirely inside the adapter and emit a shared ready-to-execute tool-call event.
- [Risk] The first real provider path may reveal missing config or error classifications in `internal/core` and `internal/api`. -> Mitigation: allow focused shared-runtime adjustments, but only when they remain provider-neutral.
- [Risk] Over-designing a reusable HTTP transport now could delay the first real adapter. -> Mitigation: keep transport logic local to the Anthropic adapter unless a clear second-provider reuse need appears.
- [Risk] Live verification can be flaky due to credentials, quotas, or external API behavior. -> Mitigation: keep live verification manual and non-blocking for automated completion, while requiring robust mocked transport coverage.

## Migration Plan

1. Add Anthropic-specific config plumbing for API key, base URL, provider selection, and settings-file loading while keeping the existing fake provider available.
2. Implement the Anthropic adapter and capability descriptors behind the existing provider registry.
3. Update runtime wiring so `--provider anthropic` resolves and executes through the real adapter.
4. Add mocked transport tests for request translation, stream mapping, tool continuation, and error normalization.
5. Run manual live verification against Anthropic when credentials are available.

Rollback is straightforward: revert runtime registration of the Anthropic adapter and continue using the fake provider path.

## Open Questions

- Which Anthropic model or models should be the initial tool-ready default descriptor for the first live path?
- Do streamed tool-call arguments require a distinct intermediate event in the shared contract, or can the adapter buffer until a complete call is ready?
- Should live smoke verification be documented as an explicit optional task in this change, or only as a post-implementation verification note?
