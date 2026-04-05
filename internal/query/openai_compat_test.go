package query

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"holycode/internal/api"
	"holycode/internal/core"
	"holycode/internal/inference"
	"holycode/internal/providers/openaicompat"
	"holycode/internal/tools"
)

func TestRunOpenAICompatibleTextOnlySkipsToolAdvertisement(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if _, ok := payload["tools"]; ok {
			t.Fatalf("expected no tools for text-only model, got %#v", payload["tools"])
		}

		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, strings.Join([]string{
			`data: {"choices":[{"delta":{"role":"assistant","content":"plain text"}}]}`,
			"",
			`data: {"choices":[{"finish_reason":"stop"}],"usage":{"prompt_tokens":4,"completion_tokens":2}}`,
			"",
			`data: [DONE]`,
			"",
		}, "\n"))
	}))
	defer server.Close()

	provider := &openaicompat.Provider{Client: server.Client()}
	runtime := api.NewRuntime(inference.NewRegistry(provider))
	result, err := Run(context.Background(), runtime, tools.NewRegistry(testTool{name: "Read", output: "fixture"}), core.Config{
		ProviderName: "openai-compatible",
		Model:        "compat-text",
		BaseURL:      server.URL,
		APIKey:       "test-key",
		Prompt:       "hello",
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.Text != "plain text" {
		t.Fatalf("expected final text %q, got %q", "plain text", result.Text)
	}
}

func TestRunOpenAICompatibleToolTierExecutesToolAndContinuesTurn(t *testing.T) {
	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		w.Header().Set("Content-Type", "text/event-stream")
		if requestCount == 1 {
			_, _ = io.WriteString(w, strings.Join([]string{
				`data: {"choices":[{"delta":{"tool_calls":[{"index":0,"id":"call_1","type":"function","function":{"name":"Read","arguments":"{\"path\":\"/tmp/demo\"}"}}]}}]}`,
				"",
				`data: {"choices":[{"finish_reason":"tool_calls"}],"usage":{"prompt_tokens":7,"completion_tokens":3}}`,
				"",
				`data: [DONE]`,
				"",
			}, "\n"))
			return
		}

		messages := payload["messages"].([]any)
		if len(messages) != 3 {
			t.Fatalf("expected 3 messages in continuation request, got %d", len(messages))
		}
		_, _ = io.WriteString(w, strings.Join([]string{
			`data: {"choices":[{"delta":{"role":"assistant","content":"tool complete"}}]}`,
			"",
			`data: {"choices":[{"finish_reason":"stop"}],"usage":{"prompt_tokens":9,"completion_tokens":1}}`,
			"",
			`data: [DONE]`,
			"",
		}, "\n"))
	}))
	defer server.Close()

	provider := &openaicompat.Provider{Client: server.Client()}
	runtime := api.NewRuntime(inference.NewRegistry(provider))
	result, err := Run(context.Background(), runtime, tools.NewRegistry(testTool{name: "Read", output: "fixture"}), core.Config{
		ProviderName: "openai-compatible",
		Model:        "compat-tools",
		BaseURL:      server.URL,
		APIKey:       "test-key",
		Prompt:       "read file",
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.Text != "tool complete" {
		t.Fatalf("expected final text %q, got %q", "tool complete", result.Text)
	}
	if len(result.ToolResults) != 1 {
		t.Fatalf("expected 1 tool result, got %d", len(result.ToolResults))
	}
}
