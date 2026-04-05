## Provider Verification Notes

This change adds two new headless provider families:

- `openai-responses`
- `openai-compatible`

The deterministic verification baseline remains the automated Go test suite. The commands below are manual smoke paths for exercising the adapters through `cmd/holy` once credentials and endpoints are available.

## OpenAI Responses

### Flag-driven invocation

```bash
go run ./cmd/holy \
  --provider openai-responses \
  --model gpt-5 \
  --api-key "$OPENAI_API_KEY" \
  "Say hello in one short sentence."
```

### Settings-file invocation

Project or user settings may provide:

```json
{
  "provider": "openai-responses",
  "model": "gpt-5",
  "api_key": "YOUR_OPENAI_KEY",
  "base_url": "https://api.openai.com/v1"
}
```

Then run:

```bash
go run ./cmd/holy "Say hello in one short sentence."
```

## OpenAI-Compatible

### Flag-driven invocation

```bash
go run ./cmd/holy \
  --provider openai-compatible \
  --model compat-text \
  --base-url "http://localhost:11434/v1" \
  --api-key "$OPENAI_COMPATIBLE_API_KEY" \
  "Say hello in one short sentence."
```

For a tool-capable compatibility tier, use the tool-enabled model descriptor currently wired by the rewrite:

```bash
go run ./cmd/holy \
  --provider openai-compatible \
  --model compat-tools \
  --base-url "http://localhost:11434/v1" \
  --api-key "$OPENAI_COMPATIBLE_API_KEY" \
  "Read a small file and summarize it."
```

### Settings-file invocation

Project or user settings may provide:

```json
{
  "provider": "openai-compatible",
  "model": "compat-text",
  "api_key": "YOUR_COMPAT_KEY",
  "base_url": "http://localhost:11434/v1"
}
```

Then run:

```bash
go run ./cmd/holy "Say hello in one short sentence."
```

## Verification Scope

- Automated verification covers request translation, streaming event mapping, tool-call continuation, capability gating, and normalized provider errors.
- The manual commands above were documented for operator use; they were not treated as a substitute for the automated conformance suite.
