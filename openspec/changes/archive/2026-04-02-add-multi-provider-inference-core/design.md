## Context

The current rewrite plan for Holy Code still describes the API layer in Anthropic-shaped terms: Anthropic-compatible payloads, Anthropic SSE parsing, and a query loop that implicitly assumes one provider's request and stream semantics. That is acceptable for a single-backend clone, but it is the wrong foundation for a runtime that must support Anthropic native API, OpenAI Responses API, and OpenAI-compatible APIs.

Those three surfaces overlap, but they are not isomorphic:

- Anthropic native API is message/block oriented and uses provider-specific tool-use/result semantics.
- OpenAI Responses API is item/event oriented and treats output, tools, and multi-step execution as a unified response object.
- OpenAI-compatible APIs often expose only a chat-completions-style subset and cannot be assumed to implement full Responses semantics.

If Holy Code picks any one of these protocols as its internal truth, one or both of the others will be forced through a leaky translation layer. The right architectural move is to define a provider-neutral turn model in the core runtime, then adapt each backend to it explicitly.

## Goals / Non-Goals

**Goals:**
- Define a provider-neutral inference contract for request submission, streaming output, tool-call round-trips, usage accounting, and stop conditions.
- Make provider capability differences explicit through typed descriptors rather than hidden conditionals.
- Keep Anthropic, OpenAI Responses, and OpenAI-compatible support as first-class, parallel adapters.
- Ensure the future headless runtime and query loop can consume one internal stream/event shape regardless of provider.
- Add conformance expectations for adapter behavior so future providers can be added without changing runtime semantics.

**Non-Goals:**
- Implementing all provider adapters in this change.
- Reaching perfect feature parity across all providers on day one.
- Designing around OpenAI Responses as the internal canonical protocol.
- Designing around Anthropic Messages as the internal canonical protocol.
- Solving UI, TUI, bridge, agent, or MCP concerns beyond the minimum event semantics they depend on.

## Decisions

### 1. Use an internal provider-neutral turn model

Decision:
Define Holy Code's own internal turn contract rather than adopting Anthropic Messages or OpenAI Responses as the core runtime model.

Rationale:
The runtime needs one stable language for query execution. If the core model is provider-owned, cross-provider support becomes lossy and provider-specific details leak into unrelated packages.

Alternatives considered:
- Anthropic-first model: simplest against the current spec, but would make OpenAI Responses feel bolted on.
- Responses-first model: attractive for agentic workflows, but too opinionated for broad OpenAI-compatible support and not a clean fit for Anthropic-native streaming semantics.

### 2. Separate provider capabilities from provider identity

Decision:
Represent backend support through explicit capability flags and descriptors instead of `if provider == x` behavior scattered through the runtime.

Rationale:
Different providers, and even different endpoints within a provider family, vary along multiple axes: tool calling, tool argument streaming, response statefulness, reasoning visibility, structured output, image input, and conversation resume support. The runtime should branch on capability, not on brand name.

Alternatives considered:
- Provider-specific conditionals in the query loop: easy initially, but becomes unmaintainable as support broadens.

### 3. Make OpenAI-compatible support a lowest-common-denominator adapter

Decision:
Treat OpenAI-compatible backends as a separate adapter with a documented minimum contract, not as a full alias of OpenAI Responses support.

Rationale:
Many OpenAI-compatible endpoints only implement a subset of OpenAI semantics, often centered on chat completions. Assuming full Responses behavior would produce brittle integrations and false guarantees.

Alternatives considered:
- Fold OpenAI-compatible into the OpenAI Responses adapter: convenient naming, but technically misleading and likely to break against real deployments.

### 4. Keep tool execution outside provider adapters

Decision:
Provider adapters translate provider-specific tool-call representations into a shared internal tool-call event model. Actual tool execution remains in the core/query/tool runtime.

Rationale:
The tool system belongs to Holy Code, not to any provider. Providers may describe tool requests differently, but once translated, execution and permission handling should be uniform.

Alternatives considered:
- Let each adapter execute tools directly: would duplicate permission logic and fragment query semantics.

### 5. Update the implementation order

Decision:
The multi-provider inference core becomes a prerequisite for the headless runtime bootstrap change.

Rationale:
If the first runtime slice is implemented against Anthropic-shaped internals, later provider support will force a redesign of the API layer and the query event pipeline. Paying the abstraction cost first is cheaper overall.

Alternatives considered:
- Build Anthropic first and abstract later: faster short term, but creates exactly the wrong incentives for the rewrite foundation.

## Risks / Trade-offs

- [Risk] The internal abstraction may become too generic and erase important provider semantics. → Mitigation: keep provider-specific metadata attached to events while standardizing only the runtime-critical contract.
- [Risk] Supporting three backend families early may slow the first executable runtime slice. → Mitigation: separate “contract first” from “adapter parity first”; the abstraction lands now, adapters can arrive in phases.
- [Risk] OpenAI-compatible APIs vary too widely for one meaningful adapter contract. → Mitigation: document a minimum supported feature baseline and fail explicitly when required capabilities are absent.
- [Risk] The existing headless bootstrap change is now partially mis-scoped. → Mitigation: treat this change as its dependency and revise the earlier change before implementation begins.

## Migration Plan

1. Define the internal provider-neutral request/event/tool contracts.
2. Define the capability matrix and model descriptor structure.
3. Specify adapter responsibilities for Anthropic, OpenAI Responses, and OpenAI-compatible backends.
4. Revise the headless bootstrap change so `internal/api` and `internal/query` depend on the new abstraction.
5. Implement adapters incrementally behind the common runtime contract.

Rollback is straightforward because this is still pre-implementation architecture work: the main recovery path is to revise the change artifacts before code is written.

## Open Questions

- How much provider-specific metadata should remain accessible on the unified event stream for debugging and advanced features?
- Should OpenAI-compatible support require tool calling in the initial supported baseline, or allow a text-only first tier?
- Should model capability descriptors be fully static config, runtime-discovered, or a hybrid of both?
