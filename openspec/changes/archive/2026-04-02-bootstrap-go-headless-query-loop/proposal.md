## Why

The current specs describe a full clean-room Go rewrite of Holy Code, but the target surface is too large to implement safely as a single first change. After defining a separate multi-provider inference core, this change should now focus on the first executable headless slice that proves the runtime can sit on top of that abstraction without reintroducing provider-specific assumptions.

## What Changes

- Introduce a headless Go CLI entrypoint that can execute a single prompt without the React/Ink TUI.
- Define the foundational Go package layout for `cmd/holy` and `internal/{core,api,query,tools}` in a way that consumes the provider-neutral inference core.
- Implement the minimum query loop needed to send a request through the shared inference layer, stream assistant output, execute tool calls, and continue until a stop condition is reached.
- Introduce an initial local tool set for the rewrite: `Bash`, `Read`, and `Edit`, with permission checks aligned to the existing specs.
- Add conformance-oriented tests and fixtures derived from the current spec set so future rewrite steps have a stable behavioral baseline across the shared runtime contract.
- Explicitly depend on the `add-multi-provider-inference-core` change for provider-neutral request and streaming semantics.
- Explicitly defer TUI, slash commands, MCP, bridge/remote sessions, background agents, memory systems, plugins, and special systems to later changes.

## Capabilities

### New Capabilities
- `headless-session-runner`: Run a non-interactive Holy Code session through the Go runtime, including provider-neutral request construction, streaming output, tool execution, and final assistant completion.
- `minimal-local-toolchain`: Provide the initial Go tool surface for local coding workflows with `Bash`, `Read`, and `Edit`, including read/write safety and permission gating semantics required by the rewrite specs.
- `rewrite-conformance-baseline`: Define executable parity checks and fixtures that anchor the Go rewrite to the repository's behavioral specs and the shared provider-neutral runtime contract before larger subsystems are added.

### Modified Capabilities

None.

## Impact

- Adds the first real implementation target for the Go rewrite under the planned package structure from `spec/13_go_codebase.md`.
- Depends on the `add-multi-provider-inference-core` change to avoid baking Anthropic-specific protocol details into the initial runtime slice.
- Establishes the dependency spine for follow-on changes in commands, TUI, MCP, bridge, and agents.
- Creates initial test fixtures and acceptance criteria that reduce rewrite drift as more subsystems are ported.
- Keeps the first change intentionally headless so we can validate core runtime behavior before taking on the far larger UI and remote-session surfaces.
