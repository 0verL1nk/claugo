# Anthropic Live Smoke Verification

## Purpose

Manual smoke path for validating the first real-provider headless execution flow against Anthropic when credentials are available.

## Prerequisites

- `ANTHROPIC_API_KEY` is set, or `api_key` is configured in `~/.holy/settings.json` or `.holy/settings.json`
- a tool-ready Anthropic model name is chosen
- outbound network access to the configured Anthropic base URL is available

## Command

```bash
go run ./cmd/holy --provider anthropic --model <model> "say hello in one short sentence"
```

Settings may also be sourced from:
- `~/.holy/settings.json`
- `.holy/settings.json`

Flags override settings-file values for the current invocation.

## Tool Continuation Smoke

Use a prompt that encourages a local tool call against a harmless test fixture, for example a temporary file created in the workspace:

```bash
printf 'fixture text\n' > /tmp/holy-anthropic-smoke.txt
go run ./cmd/holy --provider anthropic --model <model> "Read /tmp/holy-anthropic-smoke.txt and answer with its contents."
```

Expected behavior:
- text is streamed incrementally
- the runtime can execute the requested local tool
- the tool result is continued through the same Anthropic turn
- the final assistant output is printed to stdout

## Current Session Status

Live smoke was run successfully in this session against a configured Anthropic-compatible endpoint via `.holy/settings.json`.

Validated commands:

```bash
go run ./cmd/holy "Reply with exactly: smoke-ok"
go run ./cmd/holy "Read /tmp/holy-anthropic-smoke.txt and reply with exactly its contents."
```

Observed results:
- text-only smoke returned `smoke-ok`
- tool smoke returned `fixture text`
