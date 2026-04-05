# Holy Code Multi-Provider Headless Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first runnable Go headless slice of Holy Code on top of a provider-neutral inference core that supports Anthropic API, OpenAI Responses API, and OpenAI-compatible APIs.

**Architecture:** Introduce a shared internal inference contract that the query loop consumes, keep tool execution in Holy Code core, and isolate provider-specific protocol translation inside adapter packages. Land the first executable path with a fake provider for deterministic tests, then wire real adapters behind the same interface.

**Module naming:** Keep the Go module on a custom local name (`module holycode`) for now. Do not bind the rewrite to a personal GitHub import path before the package namespace is intentionally finalized.

**Tech Stack:** Go 1.24+, stdlib `context`/`net/http`/`encoding/json`, Cobra for CLI, Go `_test.go` unit tests, provider adapters for Anthropic, OpenAI Responses, and OpenAI-compatible APIs.

---

## File Structure

**Create:**
- `go.mod`
- `cmd/holy/main.go`
- `internal/core/config.go`
- `internal/core/errors.go`
- `internal/core/messages.go`
- `internal/inference/types.go`
- `internal/inference/provider.go`
- `internal/inference/capabilities.go`
- `internal/inference/registry.go`
- `internal/inference/fake/fake_provider.go`
- `internal/api/runtime.go`
- `internal/query/loop.go`
- `internal/query/loop_test.go`
- `internal/tools/tool.go`
- `internal/tools/registry.go`
- `internal/tools/read.go`
- `internal/tools/edit.go`
- `internal/tools/bash.go`
- `internal/tools/read_test.go`
- `internal/tools/edit_test.go`
- `internal/tools/bash_test.go`
- `internal/providers/anthropic/adapter.go`
- `internal/providers/anthropic/adapter_test.go`
- `internal/providers/openairesponses/adapter.go`
- `internal/providers/openairesponses/adapter_test.go`
- `internal/providers/openaicompat/adapter.go`
- `internal/providers/openaicompat/adapter_test.go`
- `internal/providers/conformance/conformance_test.go`

**Modify later if needed:**
- [spec/13_go_codebase.md](/home/ling/claugo/spec/13_go_codebase.md)
- [openspec/changes/add-multi-provider-inference-core/tasks.md](/home/ling/claugo/openspec/changes/add-multi-provider-inference-core/tasks.md)
- [openspec/changes/bootstrap-go-headless-query-loop/tasks.md](/home/ling/claugo/openspec/changes/bootstrap-go-headless-query-loop/tasks.md)

**Responsibility map:**
- `internal/core`: config, runtime errors, provider selection input, shared message containers.
- `internal/inference`: the provider-neutral request/event/capability contract.
- `internal/providers/*`: translate provider-native protocols into `internal/inference`.
- `internal/api`: runtime-facing provider selection and adapter invocation only.
- `internal/query`: orchestrates a turn using inference events plus Holy Code tools.
- `internal/tools`: shared tool execution and permission semantics.
- `cmd/holy`: argv/stdin UX and headless command wiring.

### Task 1: Scaffold The Module And Headless Entry Skeleton

**Files:**
- Create: `go.mod`
- Create: `cmd/holy/main.go`
- Create: `internal/core/config.go`
- Test: `cmd/holy/main.go` via `go test ./...`

- [ ] **Step 1: Write the failing smoke test expectation**

Document the initial package compile target in comments near the new command skeleton:

```go
// Goal of the first smoke test:
// 1. `go test ./...` compiles the module
// 2. `cmd/holy` parses a prompt source but returns "not implemented" for runtime execution
```

- [ ] **Step 2: Run the empty-module test to verify failure**

Run: `go test ./...`
Expected: FAIL because `go.mod` and Go packages do not exist yet.

- [ ] **Step 3: Write the minimal module and command skeleton**

Create `go.mod` and `cmd/holy/main.go` with a minimal Cobra entrypoint:

```go
package main

import "fmt"

func main() {
    fmt.Println("holy: runtime not implemented")
}
```

Also create `internal/core/config.go` with:
- `ProviderName string`
- `Model string`
- `BaseURL string`
- `Prompt string`

- [ ] **Step 4: Run the smoke test again**

Run: `go test ./...`
Expected: PASS compile-only, with no runtime behavior yet.

- [ ] **Step 5: Commit**

```bash
git add go.mod cmd/holy/main.go internal/core/config.go
git commit -m "chore: scaffold holy go module and command entry"
```

### Task 2: Define The Provider-Neutral Inference Contract

**Files:**
- Create: `internal/inference/types.go`
- Create: `internal/inference/provider.go`
- Create: `internal/core/errors.go`
- Create: `internal/query/loop_test.go`

- [ ] **Step 1: Write the failing shared-contract tests**

Add `internal/query/loop_test.go` with a fake-provider-driven test:

```go
func TestLoopConsumesProviderNeutralEvents(t *testing.T) {
    // fake provider emits text delta -> tool call -> completion
    // query loop should not care which backend produced them
}
```

- [ ] **Step 2: Run the targeted test to verify failure**

Run: `go test ./internal/query -run TestLoopConsumesProviderNeutralEvents -v`
Expected: FAIL because inference types and provider interface do not exist.

- [ ] **Step 3: Write the minimal shared contract**

In `internal/inference/types.go`, define:
- `TurnRequest`
- `TurnEvent`
- `TurnEventType`
- `ToolCall`
- `ToolResult`
- `Usage`
- `StopReason`

In `internal/inference/provider.go`, define:

```go
type Provider interface {
    Name() string
    Capabilities(model string) Capabilities
    RunTurn(ctx context.Context, req TurnRequest) (<-chan TurnEvent, error)
}
```

In `internal/core/errors.go`, define normalized error kinds:
- `Config`
- `Provider`
- `RateLimit`
- `Auth`
- `Tool`
- `Other`

- [ ] **Step 4: Run the targeted test again**

Run: `go test ./internal/query -run TestLoopConsumesProviderNeutralEvents -v`
Expected: still FAIL, but now on missing query loop behavior instead of missing types.

- [ ] **Step 5: Commit**

```bash
git add internal/inference/types.go internal/inference/provider.go internal/core/errors.go internal/query/loop_test.go
git commit -m "feat: add provider-neutral inference contract"
```

### Task 3: Add Capability And Model Registry Support

**Files:**
- Create: `internal/inference/capabilities.go`
- Create: `internal/inference/registry.go`
- Create: `internal/inference/fake/fake_provider.go`
- Test: `internal/inference/registry_test.go`

- [ ] **Step 1: Write the failing capability tests**

Create `internal/inference/registry_test.go`:

```go
func TestRegistryReturnsModelDescriptorWithoutHardcodedGlobalIDs(t *testing.T) {}
func TestCapabilityChecksGateUnsupportedFeatures(t *testing.T) {}
```

- [ ] **Step 2: Run the capability tests to verify failure**

Run: `go test ./internal/inference -run 'TestRegistry|TestCapability' -v`
Expected: FAIL because capability and registry code do not exist.

- [ ] **Step 3: Write the minimal implementation**

In `internal/inference/capabilities.go`, define:
- `Capabilities`
- `SupportsToolCalls bool`
- `SupportsToolArgStreaming bool`
- `SupportsStructuredOutput bool`
- `SupportsConversationState bool`

In `internal/inference/registry.go`, define:
- `ModelDescriptor`
- `Registry`
- `Lookup(provider, model string) (ModelDescriptor, error)`

In `internal/inference/fake/fake_provider.go`, implement a test double that can emit scripted events and advertise scripted capabilities.

- [ ] **Step 4: Run the capability tests again**

Run: `go test ./internal/inference -run 'TestRegistry|TestCapability' -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/inference/capabilities.go internal/inference/registry.go internal/inference/fake/fake_provider.go internal/inference/registry_test.go
git commit -m "feat: add model registry and capability descriptors"
```

### Task 4: Implement The Runtime-Facing API Layer And Query Loop

**Files:**
- Create: `internal/api/runtime.go`
- Create: `internal/query/loop.go`
- Modify: `internal/query/loop_test.go`

- [ ] **Step 1: Expand the failing query loop tests**

Add tests for:

```go
func TestLoopStreamsTextInOrder(t *testing.T) {}
func TestLoopContinuesAfterToolCall(t *testing.T) {}
func TestLoopReturnsProviderErrors(t *testing.T) {}
```

- [ ] **Step 2: Run the query tests to verify failure**

Run: `go test ./internal/query -v`
Expected: FAIL because runtime orchestration is not implemented.

- [ ] **Step 3: Write the minimal runtime and loop**

In `internal/api/runtime.go`:
- accept `core.Config`
- resolve provider from registry/config
- expose one method that returns the shared turn-event stream

In `internal/query/loop.go`:
- call runtime provider
- accumulate `message_delta`
- dispatch `tool_call_completed` requests to the tool runtime
- feed tool results back into subsequent turn requests
- return final assembled text plus normalized usage/error info

- [ ] **Step 4: Run the query tests again**

Run: `go test ./internal/query -v`
Expected: PASS using the fake provider.

- [ ] **Step 5: Commit**

```bash
git add internal/api/runtime.go internal/query/loop.go internal/query/loop_test.go
git commit -m "feat: add provider-neutral query runtime loop"
```

### Task 5: Implement The Minimal Local Toolchain

**Files:**
- Create: `internal/tools/tool.go`
- Create: `internal/tools/registry.go`
- Create: `internal/tools/read.go`
- Create: `internal/tools/edit.go`
- Create: `internal/tools/bash.go`
- Create: `internal/tools/read_test.go`
- Create: `internal/tools/edit_test.go`
- Create: `internal/tools/bash_test.go`

- [ ] **Step 1: Write the failing tool tests**

Add tests for:

```go
func TestReadRegistersFileState(t *testing.T) {}
func TestEditRejectsExistingFileWithoutPriorRead(t *testing.T) {}
func TestBashRejectsUnsafeCommandWithoutApproval(t *testing.T) {}
```

- [ ] **Step 2: Run the tool tests to verify failure**

Run: `go test ./internal/tools -v`
Expected: FAIL because tool runtime and implementations do not exist.

- [ ] **Step 3: Write the minimal tool runtime**

Implement:
- a shared `Tool` interface
- registry lookup by tool name
- `Read` with file-state registration
- `Edit` with read-before-edit enforcement
- `Bash` with explicit safe/unsafe classification hook

- [ ] **Step 4: Run the tool tests again**

Run: `go test ./internal/tools -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/tools/tool.go internal/tools/registry.go internal/tools/read.go internal/tools/edit.go internal/tools/bash.go internal/tools/read_test.go internal/tools/edit_test.go internal/tools/bash_test.go
git commit -m "feat: add headless read edit bash toolchain"
```

### Task 6: Add Adapter Stubs And Conformance Tests

**Files:**
- Create: `internal/providers/anthropic/adapter.go`
- Create: `internal/providers/anthropic/adapter_test.go`
- Create: `internal/providers/openairesponses/adapter.go`
- Create: `internal/providers/openairesponses/adapter_test.go`
- Create: `internal/providers/openaicompat/adapter.go`
- Create: `internal/providers/openaicompat/adapter_test.go`
- Create: `internal/providers/conformance/conformance_test.go`

- [ ] **Step 1: Write the failing adapter conformance tests**

Add one shared conformance test helper:

```go
func RunProviderContractSuite(t *testing.T, provider inference.Provider) {}
```

And per-adapter tests that verify:
- provider name
- capability advertisement
- translation into shared event types
- normalized provider error mapping

- [ ] **Step 2: Run the provider tests to verify failure**

Run: `go test ./internal/providers/... -v`
Expected: FAIL because adapters do not exist.

- [ ] **Step 3: Write minimal adapter shells**

Implement adapters as thin translation layers:
- `anthropic`: native messages/events -> shared events
- `openairesponses`: responses items/events -> shared events
- `openaicompat`: chat-style subset -> shared events, with explicit unsupported-feature errors

Do not add full HTTP production logic yet; start with request/response translation units and scripted stream decoders.

- [ ] **Step 4: Run the provider tests again**

Run: `go test ./internal/providers/... -v`
Expected: PASS for contract and mapping tests.

- [ ] **Step 5: Commit**

```bash
git add internal/providers/anthropic/adapter.go internal/providers/anthropic/adapter_test.go internal/providers/openairesponses/adapter.go internal/providers/openairesponses/adapter_test.go internal/providers/openaicompat/adapter.go internal/providers/openaicompat/adapter_test.go internal/providers/conformance/conformance_test.go
git commit -m "feat: add multi-provider adapter contract stubs"
```

### Task 7: Wire The CLI And Add End-To-End Headless Verification

**Files:**
- Modify: `cmd/holy/main.go`
- Modify: `internal/api/runtime.go`
- Modify: `internal/query/loop.go`
- Test: `go test ./...`

- [ ] **Step 1: Write the failing CLI behavior test expectation**

Document expected behavior:
- prompt from argv or stdin
- provider selected from config/flags
- fake provider usable in tests
- non-zero exit on provider/tool errors

- [ ] **Step 2: Run the full test suite before CLI wiring**

Run: `go test ./...`
Expected: PASS on unit packages, but no real CLI end-to-end path yet.

- [ ] **Step 3: Implement minimal headless CLI wiring**

`cmd/holy/main.go` should:
- parse prompt input
- parse provider/model/base URL options
- instantiate registry/runtime/query/tool registry
- print streamed deltas or final text
- set exit status on normalized runtime errors

- [ ] **Step 4: Run the full verification suite**

Run: `go test ./...`
Expected: PASS.

Run: `go run ./cmd/holy --provider fake --model test \"hello\"`
Expected: deterministic fake-provider output.

- [ ] **Step 5: Commit**

```bash
git add cmd/holy/main.go internal/api/runtime.go internal/query/loop.go
git commit -m "feat: wire holy headless cli to multi-provider runtime"
```

### Task 8: Update Specs And Verification Notes

**Files:**
- Modify: `openspec/changes/add-multi-provider-inference-core/tasks.md`
- Modify: `openspec/changes/bootstrap-go-headless-query-loop/tasks.md`
- Modify: `spec/13_go_codebase.md`

- [ ] **Step 1: Write the failing documentation checklist**

Checklist:
- package boundaries match implementation
- provider-neutral contract is reflected in spec text
- headless runtime plan references the inference core dependency

- [ ] **Step 2: Review docs against the landed implementation**

Run: `rg -n "Anthropic-compatible request payloads|cmd/holy|provider-neutral|openai-compatible" spec openspec/changes`
Expected: identify stale wording before editing.

- [ ] **Step 3: Update the docs minimally**

Adjust only the lines needed to keep specs and implementation aligned. Do not expand scope.

- [ ] **Step 4: Re-run validation**

Run: `openspec validate add-multi-provider-inference-core`
Expected: valid

Run: `openspec validate bootstrap-go-headless-query-loop`
Expected: valid

- [ ] **Step 5: Commit**

```bash
git add openspec/changes/add-multi-provider-inference-core/tasks.md openspec/changes/bootstrap-go-headless-query-loop/tasks.md spec/13_go_codebase.md
git commit -m "docs: align specs with multi-provider headless foundation"
```
