## 1. Bootstrap The Go Rewrite Spine

- [x] 1.0 Rebase this change on `add-multi-provider-inference-core` and treat its shared inference contract as a prerequisite input.
- [x] 1.1 Create `go.mod`, `cmd/holy`, and the initial `internal/{core,api,query,tools}` package layout described by the change design.
- [x] 1.2 Define foundational shared types for config, messages, content blocks, tool definitions, tool results, and classified runtime errors in `internal/core`.
- [x] 1.3 Add a headless CLI entrypoint that accepts prompt input from argv and/or stdin and wires the runtime packages together without the TUI.

## 2. Build The Headless Session Runner

- [x] 2.1 Implement the runtime-facing `internal/api` layer as a consumer of the shared provider-neutral inference contract rather than a provider-native transport abstraction.
- [x] 2.2 Implement the query loop that sends the initial request through the selected provider adapter, streams assistant output, detects supported tool-call events, and continues the turn until completion.
- [x] 2.3 Add explicit headless error reporting and process exit behavior for configuration, provider, and tool-execution failures.

## 3. Implement The Minimal Local Toolchain

- [x] 3.1 Implement the shared Go tool runtime contract and tool registry used by the query loop.
- [x] 3.2 Implement `Read` with file-reading constraints and state tracking needed for later edit safety.
- [x] 3.3 Implement `Edit` with read-before-edit enforcement for existing files and tool-result error reporting.
- [x] 3.4 Implement `Bash` with permission-aware execution decisions for read-only versus unsafe commands in the initial headless flow.

## 4. Add Conformance And Verification

- [x] 4.1 Create deterministic unit and acceptance fixtures covering headless execution, shared event-stream assembly, tool-call continuation, and failure reporting.
- [x] 4.2 Add conformance coverage for the initial toolchain, including read-before-edit rejection and permission-gated bash behavior.
- [x] 4.3 Document and run the verification commands required to prove the initial runtime slice is working before the change is marked complete.
