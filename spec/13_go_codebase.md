# Holy Code — Go Codebase

## Overview

The Go rewrite in **this repository** is a **complete standalone rewrite** of the TypeScript Holy Code CLI, using `ref/cloud-code/claude-code-source/` as the reference implementation. It is not an FFI binding layer, not a partial port, and shares no runtime code with the TypeScript implementation. It re-implements the same tool names and semantics, permission model, `HOLY.md` discovery, auto-compact logic, MCP (Model Context Protocol) client, bridge protocol, and cron scheduler using goroutines, channels, and `context.Context`.

### Architecture

```
.
├── go.mod                      # Module root for the Go rewrite
├── cmd/
│   └── holy/
│       └── main.go             # Binary entry point
├── internal/
│   ├── core/                   # Shared types, config, permissions, history, hooks
│   ├── api/                    # API client + SSE streaming
│   ├── tools/                  # All tool implementations
│   ├── query/                  # Agentic query loop, compact, cron scheduler
│   ├── tui/                    # Bubble Tea terminal UI
│   ├── commands/               # Slash command implementations
│   ├── mcp/                    # MCP client
│   └── bridge/                 # Bridge to holy.ai web UI
├── spec/                       # Reverse-engineered specs used to drive the rewrite
└── ref/cloud-code/claude-code-source/
    └── src/                    # Reference TypeScript source
```

**Dependency flow:**

```
cmd/holy -> query -> tools -> core
                    ↘        ↗
                     api -> core
                      ↓
                 commands -> core
                      ↓
                    tui -> core
                      ↓
                    mcp -> core
                      ↓
                 bridge -> core
```

---

## Module Root: `go.mod`

**Path:** `./go.mod`

Go module rooted at `go.mod`, targeting modern Go (`go 1.24` or newer) across the command and internal packages.
Until a public repository path is intentionally chosen, the rewrite uses a custom local module path (`module holycode`) rather than a personal GitHub import path.

### Packages

| Package Path | Package Name | Type |
|---|---|---|
| `internal/core` | `core` | Internal package |
| `internal/api` | `api` | Internal package |
| `internal/tools` | `tools` | Internal package |
| `internal/query` | `query` | Internal package |
| `internal/tui` | `tui` | Internal package |
| `internal/commands` | `commands` | Internal package |
| `internal/mcp` | `mcp` | Internal package |
| `internal/bridge` | `bridge` | Internal package |
| `cmd/holy` | `main` | Command (`go build ./cmd/holy`) |

### Key Shared Dependencies

| Package | Version | Purpose |
|---|---|---|
| `context` | stdlib | Cancellation, deadlines, request scoping |
| `net/http` | stdlib | HTTP client and streaming transport |
| `encoding/json` | stdlib | JSON encode/decode |
| `log/slog` | stdlib | Structured logging |
| `os/exec` | stdlib | Shell and subprocess execution |
| `github.com/spf13/cobra` | 1.x | CLI flags and subcommands |
| `github.com/charmbracelet/bubbletea` | 1.x | Terminal UI event loop |
| `github.com/charmbracelet/lipgloss` | 1.x | Terminal styling and layout |
| `github.com/charmbracelet/bubbles` | 0.x | Input, spinner, viewport widgets |
| `golang.org/x/sync/errgroup` | 0.x | Coordinated concurrency |
| `golang.org/x/term` | 0.x | Raw terminal mode |
| `github.com/bmatcuk/doublestar/v4` | 4.x | Glob pattern matching |
| `github.com/google/uuid` | 1.x | UUID generation |

---

## Package: `core`

**Path:** `internal/core/`

Central shared package. Defines the types and helpers consumed by every other package.

### Key Responsibilities

- Canonical message/content types shared between the API, query loop, tools, and bridge.
- Runtime configuration, settings loading, permission modes, and hook configuration.
- History and cost-tracking primitives.
- Shared constants such as model names, tool names, and settings filenames.

### Key Types

**`ClaudeError`** wraps a typed error kind plus optional metadata:

- `Api`
- `ApiStatus`
- `Auth`
- `PermissionDenied`
- `Tool`
- `Io`
- `JSON`
- `HTTP`
- `RateLimit`
- `ContextWindowExceeded`
- `MaxTokensReached`
- `Cancelled`
- `Config`
- `Mcp`
- `Other`

**`Role`**: `user`, `assistant`

**`ContentBlock`** supports the same logical variants as the TypeScript runtime:

- Text
- Image
- ToolUse
- ToolResult
- Thinking
- RedactedThinking
- Document

**`Message`** stores `role` plus either plain text or block content. Helper constructors mirror the TypeScript implementation for user messages, assistant messages, block-based messages, and tool-use inspection.

**`UsageInfo`** tracks input/output tokens plus cache read/create tokens and exposes aggregate helpers.

**`ToolDefinition`** contains the tool name, description, and JSON schema.

**`Config`** contains runtime state such as:

- API key and base URL
- model and token limits
- permission mode
- output format
- max turns
- system-prompt overrides
- auto-compact settings
- thinking budget
- MCP server definitions
- hook registrations

**`Settings`** persists user preferences at `~/.holy/settings.json` using `encoding/json`, `os.ReadFile`, `os.WriteFile`, and `os.MkdirAll`.

### Implementation Notes

- JSON-heavy types use `encoding/json` with explicit marshal/unmarshal helpers where the TypeScript runtime relies on tagged or untagged unions.
- File- and path-based helpers use `os`, `io/fs`, and `path/filepath`.
- Permission and hook enums from the reference implementation map to Go string constants or typed aliases.

### Relationship to TypeScript

`internal/core` corresponds to the scattered TypeScript definitions in `src/constants/`, `src/context.ts`, `src/history.ts`, `src/cost-tracker.ts`, `src/costHook.ts`, `src/schemas/hooks.ts`, and parts of `src/services/api/`.

---

## Package: `api`

**Path:** `internal/api/`

Encapsulates API request construction, SSE parsing, streamed response decoding, retry logic, and usage accounting.

### Key Responsibilities

- Build Anthropic-compatible request payloads from `core.Message` values.
- Decode streamed SSE events into assistant content blocks and usage deltas.
- Apply retry/backoff policy for transient 429/529 and network failures.
- Normalize API errors into `core.ClaudeError`.

### Implementation Notes

- Uses `net/http` with a shared `http.Client`.
- SSE parsing is implemented with `bufio.Reader` and incremental frame parsing.
- Background streaming work runs in goroutines and reports events over channels.
- Request cancellation is wired through `context.Context`.

### Relationship to TypeScript

`internal/api` corresponds to `src/services/api/`, including the Claude client, streaming parser, and API type definitions.

---

## Package: `tools`

**Path:** `internal/tools/`

Contains the full tool catalog and shared execution framework.

### Core Interface

Each tool implements a shared execution interface shaped like:

```go
type Tool interface {
    Definition() core.ToolDefinition
    Execute(ctx context.Context, input json.RawMessage, toolCtx ToolContext) (ToolResult, error)
}
```

### Shared `ToolContext`

- Working directory and environment
- Permission resolver and approval callbacks
- Conversation/task/session identifiers
- MCP registry access
- Hook runner
- Output writer / event sink for streaming updates

### Tool Coverage

The Go rewrite preserves the same logical tool surface as the TypeScript CLI:

- Shell tools: `Bash`, `PowerShell`
- File tools: `Read`, `Edit`, `Write`, `Glob`, `Grep`
- Web tools: `WebFetch`, `WebSearch`
- Notebook / structured tools: `NotebookEdit`, `TodoWrite`, `Config`, `Brief`
- Planning / control tools: `EnterPlanMode`, `ExitPlanMode`, `Sleep`
- Agent / task tools: `Task`, task lifecycle helpers, `SendMessage`
- MCP tools: `ListMcpResources`, `ReadMcpResource`
- Workspace tools: `EnterWorktree`, `ExitWorktree`
- Discovery helpers: `ToolSearch`, `Skill`

### Selected Implementation Notes

**Bash / PowerShell**

- Spawned with `exec.CommandContext`.
- Stdout and stderr are collected with pipes plus `bufio.Reader`.
- Respect the same permission gates and sandbox policy as the TypeScript runtime.

**Read / Edit / Write**

- Use `os.ReadFile`, `os.WriteFile`, and `os.MkdirAll`.
- Preserve exact file-edit semantics from the TypeScript tool contracts.

**Glob / Grep**

- Glob matching uses `doublestar`.
- Grep uses either direct streaming scanners or delegated `rg` subprocesses depending on platform/configuration.

**WebFetch / WebSearch**

- HTTP fetches use `net/http` with timeouts and redirect limits aligned with the TypeScript behavior.
- Search providers remain abstracted behind a provider interface.

**NotebookEdit**

- Parses `.ipynb` payloads with `encoding/json`.
- Preserves notebook cell metadata unless explicitly modified by the tool.

**Sleep**

- Uses `time.Sleep` with the same maximum duration cap.

### Relationship to TypeScript

`internal/tools` corresponds to the TypeScript tool implementations in `src/` and `src/tools/`. Tool names and user-visible semantics remain aligned.

---

## Package: `query`

**Path:** `internal/query/`

Implements the core agentic query loop and the coordination logic around it.

### Key Responsibilities

- Build the request context for a user turn.
- Stream assistant output from `internal/api`.
- Detect tool-use blocks, execute tools, and feed results back into the loop.
- Enforce token budgets, compact context, and handle stop conditions.
- Run sub-agent flows and scheduled background tasks.

### Major Modules

**`loop.go`**

- Main turn execution loop.
- Uses channels and `select` to merge API events, tool results, and cancellation.

**`compact.go`**

- Auto-compact heuristics and summary generation.
- Mirrors the TypeScript compaction thresholds and trigger conditions.

**`agent_tool.go`**

- Sub-agent orchestration, message routing, and result collection.
- Owns the same task-oriented semantics as the TypeScript `Task` tool.

**`cron_scheduler.go`**

- Background cron registration and dispatch.
- Scheduler loop runs in a dedicated goroutine and honors `context.Context` cancellation.

### Relationship to TypeScript

`internal/query` corresponds to `src/query.ts`, `src/query/`, `src/services/compact/autoCompact.ts`, `src/coordinator/`, and the query-facing portions of `src/services/autoDream/`.

---

## Package: `tui`

**Path:** `internal/tui/`

Terminal UI built on Bubble Tea plus Lip Gloss. It replaces the TypeScript `ink`/React rendering layer with a Go-native terminal application model.

### Key Responsibilities

- Render message history, streaming output, input boxes, status bars, and dialogs.
- Handle keyboard input, resize events, and focus transitions.
- Surface permission prompts and progress UI during tool execution.
- Keep interactive and headless rendering paths consistent with the TypeScript CLI.

### Implementation Notes

- Bubble Tea `Model` instances manage update/render cycles.
- Lip Gloss handles layout, colors, borders, and style composition.
- Alternate screen and raw terminal setup use `golang.org/x/term` or `tcell`-style terminal primitives, depending on the final implementation choice.
- The TUI differs architecturally from React/Ink, but preserves equivalent user-visible behavior.

### Relationship to TypeScript

`internal/tui` replaces the TypeScript `src/ink/` rendering system, `src/components/`, and the interactive CLI screen composition.

---

## Package: `commands`

**Path:** `internal/commands/`

Contains slash-command registration, parsing, validation, and execution.

### Key Responsibilities

- Register built-in slash commands and aliases.
- Parse command arguments/options.
- Dispatch to command handlers with the same semantics as the TypeScript runtime.
- Expose command metadata to help, completion, and the interactive prompt.

### Relationship to TypeScript

`internal/commands` corresponds to the TypeScript `src/commands/` directory and preserves command names and behavioral contracts.

---

## Package: `mcp`

**Path:** `internal/mcp/`

Implements the Model Context Protocol client and resource/tool bridge.

### Key Responsibilities

- JSON-RPC transport over stdio and remote transports.
- MCP initialization handshake and capability discovery.
- Tool-definition conversion into Holy Code tool schemas.
- Resource listing, reading, and namespaced tool dispatch.

### Implementation Notes

- Protocol messages are represented as Go structs with `encoding/json` tags.
- Long-lived MCP connections run in goroutines with request/response correlation maps.
- Cancellation and shutdown propagate through `context.Context`.

### Relationship to TypeScript

`internal/mcp` corresponds to `src/services/mcpClient.ts` and the TypeScript MCP tooling surface.

---

## Package: `bridge`

**Path:** `internal/bridge/`

Implements the polling-based bridge protocol used to connect the CLI with remote or browser-driven Claude sessions.

### Key Responsibilities

- Authenticate sessions and manage polling loops.
- Encode/decode bridge messages and events.
- Support JWT helpers, trusted-device tokens, and remote session lifecycle management.

### Implementation Notes

- Long-polling uses `net/http`.
- JWT parsing uses `encoding/base64`, `encoding/json`, and HMAC/crypto helpers from the standard library.
- Trusted-device enrollment and fingerprinting use `os.Hostname`, environment inspection, `crypto/sha256`, and `encoding/hex`.

### Relationship to TypeScript

`internal/bridge` corresponds to `src/bridge/`, including bridge session management, JWT helpers, and trusted-device handling.

---

## Command: `holy`

**Path:** `cmd/holy/main.go`

Binary entry point. Produces the `holy` executable and wires all internal packages together.

### Key Responsibilities

- Parse CLI flags and subcommands.
- Build runtime config from settings, environment, and flags.
- Select interactive vs headless execution paths.
- Load the embedded system prompt and initialize the query loop.

### Implementation Notes

- Flags/subcommands are defined with Cobra or an equivalent Go CLI framework.
- Embedded static prompt assets use `//go:embed`.
- Interactive mode boots the Bubble Tea application.
- Headless mode streams plain text or JSON output while still supporting tool execution.

### Relationship to TypeScript

`cmd/holy` corresponds to `src/entrypoints/cli.tsx`, `src/main.tsx`, and the startup wiring around the TypeScript CLI.

---

## Cross-Cutting Architecture Notes

### Concurrency Model

All concurrent work uses goroutines. Long-lived flows such as SSE streaming, background cron dispatch, task orchestration, and TUI event handling communicate via channels and are canceled with `context.Context`.

### Cancellation

`context.Context` is threaded through the query loop, cron scheduler, tool execution, and bridge interactions. `Ctrl+C` cancels the root context and allows each subsystem to unwind cleanly.

### Shared State

Common process-wide state includes:

- task store for sub-agent lifecycle
- inbox for inter-agent messaging
- cron registry for scheduled work
- active worktree/session coordination

These are typically protected by mutexes or encapsulated behind package-level managers.

### Error Handling

- Packages return wrapped Go `error` values plus typed classification where needed.
- API, permission, and context-window failures map cleanly to the same user-visible outcomes as the TypeScript runtime.

### Prompt Caching

The API client preserves the same prompt-caching behavior as the TypeScript reference implementation. Cache-control markers remain part of the request payload, independent of the implementation language.

### Logging

- Structured logs use `log/slog`.
- Interactive mode keeps logs off the TUI unless explicitly requested.
- Verbose/debug logging is controlled through config and environment flags.

### TypeScript Parity Summary

| TypeScript Area | Go Package |
|---|---|
| `src/entrypoints/cli.tsx`, `src/main.tsx` | `cmd/holy` |
| `src/services/api/` | `internal/api` |
| `src/query.ts`, `src/query/` | `internal/query` |
| `src/components/`, `src/ink/` | `internal/tui` |
| `src/commands/` | `internal/commands` |
| `src/constants/`, `src/context.ts`, etc. | `internal/core` |
| Tool implementations (Bash, Read, Edit, etc.) | `internal/tools` |
| MCP client (`src/services/mcpClient.ts`) | `internal/mcp` |
| `src/bridge/` | `internal/bridge` |
