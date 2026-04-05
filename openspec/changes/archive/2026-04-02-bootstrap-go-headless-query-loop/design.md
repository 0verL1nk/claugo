## Context

This repository already contains a detailed behavioral specification for a clean-room Go rewrite of Holy Code, but it does not yet contain an implementation. The target system is large: CLI bootstrapping, streaming API I/O, 40+ tools, TUI, commands, MCP, bridge/remote sessions, agents, memory, and special systems.

The first executable runtime slice still needs to prove the architectural spine of the rewrite without taking on the entire product surface. However, after deciding that Holy Code must support Anthropic native API, OpenAI Responses API, and OpenAI-compatible APIs, this change can no longer define provider-shaped API contracts on its own. It must consume the provider-neutral inference layer established by `add-multi-provider-inference-core`.

The most leveraged place to start remains the headless execution path because it exercises the core runtime chain without depending on Bubble Tea, React/Ink parity, or remote-session infrastructure.

## Goals / Non-Goals

**Goals:**
- Establish the initial Go module and package boundaries described in `spec/13_go_codebase.md`.
- Deliver one runnable vertical slice: `cmd/holy -> internal/core -> internal/api -> internal/query -> internal/tools`.
- Consume the shared provider-neutral inference contract rather than introducing provider-specific streaming or request semantics in this change.
- Support one-shot, non-interactive prompt execution with streamed assistant output and tool round-trips.
- Implement the minimum viable local tool surface for coding tasks: `Bash`, `Read`, and `Edit`.
- Add conformance fixtures and acceptance tests so later rewrite stages can extend behavior without drifting from spec.

**Non-Goals:**
- Recreating the interactive TUI, prompts, dialogs, or alternate-screen behavior.
- Designing provider-neutral inference contracts or capability descriptors from scratch inside this change.
- Porting slash commands, MCP integration, bridge/remote sessions, agents, coordinator mode, memory, plugins, or special systems.
- Achieving feature-complete parity with the TypeScript runtime in a single change.
- Solving final package shapes for every future subsystem beyond the first runtime spine.

## Decisions

### 1. Start with a vertical slice, not package-only scaffolding

Decision:
Implement a runnable headless path instead of only generating empty Go packages and interfaces.

Rationale:
Scaffolding alone proves almost nothing. A vertical slice forces the key contracts to converge early: configuration loading, provider selection, streaming, tool execution, turn continuation, and user-visible output.

Alternatives considered:
- Package-first scaffolding only: low risk, but no behavioral validation.
- Full interactive CLI first: too much UI and event-loop surface for an initial change.

### 2. Make the first runtime headless-only

Decision:
The first executable path will be non-interactive and single-session only.

Rationale:
The TUI and remote systems are the largest architectural multipliers in the rewrite. Deferring them keeps the first change bounded while still validating the core agentic loop.

Alternatives considered:
- Bubble Tea from day one: adds terminal state, rendering, input focus, and dialog complexity before the runtime core is stable.
- Bridge-first remote path: depends on auth, polling, and session lifecycle machinery that is orthogonal to the rewrite foundation.

### 3. Depend on the multi-provider inference core

Decision:
This change will treat `add-multi-provider-inference-core` as a prerequisite and will consume its shared request/event/tool contracts instead of defining provider-native transport semantics locally.

Rationale:
If this change implements its own provider-shaped API layer first, the initial runtime slice will harden the wrong abstraction and force a redesign when additional providers land.

Alternatives considered:
- Keep the old Anthropic-first transport design here and abstract later: faster short term, but structurally wrong for the stated product goal.

### 4. Use the `HOLY_CODE_SIMPLE` subset as the initial tool boundary

Decision:
The initial tool set is `Bash`, `Read`, and `Edit`, aligned with the minimal tool subset already described by the specs.

Rationale:
This is the smallest meaningful tool surface for coding workflows and has a clear anchor in the existing behavior documentation. It also gives us one shell tool and two file tools, which is enough to validate permission and file-safety semantics.

Alternatives considered:
- `Read` only: insufficient to prove the agentic edit loop.
- Full file tool suite including `Write`, `Glob`, `Grep`: useful, but expands the first change without changing the architectural proof.

### 5. Treat spec-derived fixtures as a first-class deliverable

Decision:
The change will include acceptance fixtures and tests that encode the required behaviors for the initial slice.

Rationale:
This is a clean-room rewrite. The implementation needs a parity target that is executable, reviewable, and independent from the proprietary TypeScript source. Spec-derived fixtures reduce drift and make later changes safer.

Alternatives considered:
- Ad hoc unit tests only: faster initially, but weak as a long-term parity guardrail.
- Waiting for broader implementation before building fixtures: increases risk that early design mistakes become entrenched.

### 6. Keep later subsystem seams explicit

Decision:
The initial package interfaces must leave clean seams for TUI, commands, MCP, and bridge integration even though those subsystems are out of scope.

Rationale:
A headless-first slice should not force a dead-end architecture. The query loop and tool execution contracts should be usable from both headless and future interactive entrypoints.

Alternatives considered:
- Optimizing strictly for a temporary MVP binary: simpler short term, but likely to require large rewrites when the TUI arrives.

## Risks / Trade-offs

- [Risk] The headless slice may overfit one-shot execution and make the later TUI harder to integrate. → Mitigation: keep query execution separate from CLI presentation and define streaming/event interfaces outside `cmd/holy`.
- [Risk] This change may accidentally reintroduce provider-specific behavior below the abstraction boundary. → Mitigation: treat adapter-facing contracts from `add-multi-provider-inference-core` as fixed inputs and reject provider-native shortcuts in `internal/api` and `internal/query`.
- [Risk] Limiting the first tool surface to `Bash`, `Read`, and `Edit` may leave gaps for real-world workflows. → Mitigation: position this change as a foundation slice and explicitly queue `Write`, `Glob`, and `Grep` as follow-on work.
- [Risk] Spec-derived fixtures may not capture every nuance of the original runtime. → Mitigation: encode only clearly stated behaviors now and expand the fixture suite as later subsystems are specified and implemented.
- [Risk] Deferring compaction, agents, and remote flows means the first binary is intentionally incomplete. → Mitigation: make non-goals explicit in proposal, specs, and tasks so completeness expectations remain realistic.

## Migration Plan

1. Finalize the `add-multi-provider-inference-core` change and treat its artifacts as the contract for this change.
2. Create the Go module, command entrypoint, and initial internal packages.
3. Land the headless runtime slice behind the new Go binary only, using the shared inference interfaces.
4. Add conformance and unit tests for the initial execution path and minimal tools.
5. Use this change as the prerequisite for later OpenSpec changes that layer in additional tools, commands, TUI, MCP, and bridge support.

Rollback is straightforward because this is additive: if the implementation proves unsound, the new Go code can be removed without affecting existing repository behavior.

## Open Questions

- Should `Write` join the first slice if `Edit` alone makes new-file creation awkward in practice?
- How thin should `internal/api` remain once provider adapters own transport-specific parsing?
- Do we want a fake provider adapter from day one for deterministic streaming tests, or rely first on lower-level adapter/parser tests plus a small integration layer?
- Should the first binary read the prompt from argv, stdin, or both? The design should likely support both, but the exact UX can be finalized during implementation.
