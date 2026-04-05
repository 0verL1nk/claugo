# Holy Code — Progress

> Project dashboard for the current Go rewrite status.

## Status Summary

Foundation complete.
The rewrite now has a provider-neutral inference core, a runnable headless CLI slice, synced main OpenSpec capability specs, and working real-provider paths for Anthropic, OpenAI Responses, and OpenAI-compatible backends.
Next phase is improving provider capability configuration and expanding beyond the current headless foundation.

## Current Phase

The project has moved from staged provider integration into post-foundation consolidation.
Anthropic, OpenAI Responses, and OpenAI-compatible headless execution are now proven through deterministic adapter tests, shared conformance coverage, and full Go test-suite verification.
The baseline is now "preserve the shared runtime contract while making provider capability selection and real-backend verification more robust."

## Completed

- Renamed the rewrite target to `Holy Code`
- Standardized the Go entrypoint as `cmd/holy`
- Kept the Go module path custom and local as `holycode`
- Completed the multi-provider inference core change
- Completed the headless query loop bootstrap change
- Completed the Anthropic tool-ready adapter change
- Completed the OpenAI Responses adapter change
- Completed the OpenAI-compatible adapter change
- Verified the Anthropic-compatible live smoke path through `.holy/settings.json`
- Verified a real OpenAI-compatible text path against a LiteLLM endpoint
- Archived all completed OpenSpec changes to date
- Synced the resulting capability specs into `openspec/specs/`
- Restored OpenSpec validation for the main specs

## In Progress

- None

## Next

- Replace hardcoded OpenAI-compatible capability tiers with configurable provider/model descriptors
- Expand provider-specific manual verification coverage for real backends, especially tool-capable OpenAI-compatible deployments
- Decide whether to expose think/reasoning channels distinctly from final output in the headless runtime
- Continue beyond the headless slice into the next rewrite surface once provider behavior is stable

## Key Decisions

- Product name is `Holy Code`
- The Go rewrite must not reintroduce `claude` naming in the new entrypoint
- The command path is `cmd/holy`
- The module path stays custom and local for now as `holycode`
- The runtime core stays provider-neutral
- Provider adapters are translation layers, not orchestration layers
- Initial provider targets are Anthropic API, OpenAI Responses API, and OpenAI-compatible APIs

## Linked Artifacts

### Analysis Specs

- `spec/00_overview.md`
- `spec/13_go_codebase.md`

### Main OpenSpec Specs

- `openspec/specs/anthropic-provider-adapter/spec.md`
- `openspec/specs/headless-session-runner/spec.md`
- `openspec/specs/minimal-local-toolchain/spec.md`
- `openspec/specs/multi-provider-inference-core/spec.md`
- `openspec/specs/openai-compatible-provider-adapter/spec.md`
- `openspec/specs/openai-responses-provider-adapter/spec.md`
- `openspec/specs/provider-adapter-packages/spec.md`
- `openspec/specs/provider-capability-registry/spec.md`
- `openspec/specs/rewrite-conformance-baseline/spec.md`

### Active Planning

- `docs/superpowers/plans/2026-04-02-holy-code-multi-provider-headless-foundation.md`

### Archived Changes

- `openspec/changes/archive/2026-04-02-add-multi-provider-inference-core/`
- `openspec/changes/archive/2026-04-02-bootstrap-go-headless-query-loop/`
- `openspec/changes/archive/2026-04-04-add-anthropic-tool-ready-adapter/`
- `openspec/changes/archive/2026-04-05-add-openai-provider-adapters/`
