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
	"holycode/internal/providers/anthropic"
	"holycode/internal/tools"
)

func TestRunAnthropicExecutesToolAndContinuesTurn(t *testing.T) {
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
				"event: content_block_start",
				`data: {"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"tool-1","name":"Read","input":{}}}`,
				"",
				"event: content_block_delta",
				`data: {"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"{\"path\":\"/tmp/demo\"}"}}`,
				"",
				"event: content_block_stop",
				`data: {"type":"content_block_stop","index":0}`,
				"",
				"event: message_delta",
				`data: {"type":"message_delta","delta":{"stop_reason":"tool_use"},"usage":{"output_tokens":3}}`,
				"",
				"event: message_stop",
				`data: {"type":"message_stop"}`,
				"",
			}, "\n"))
			return
		}

		messages := payload["messages"].([]any)
		if len(messages) != 3 {
			t.Fatalf("expected 3 messages in continuation request, got %d", len(messages))
		}
		_, _ = io.WriteString(w, strings.Join([]string{
			"event: content_block_start",
			`data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`,
			"",
			"event: content_block_delta",
			`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"tool complete"}}`,
			"",
			"event: content_block_stop",
			`data: {"type":"content_block_stop","index":0}`,
			"",
			"event: message_delta",
			`data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":1}}`,
			"",
			"event: message_stop",
			`data: {"type":"message_stop"}`,
			"",
		}, "\n"))
	}))
	defer server.Close()

	provider := &anthropic.Provider{Client: server.Client()}
	runtime := api.NewRuntime(inference.NewRegistry(provider))
	result, err := Run(context.Background(), runtime, tools.NewRegistry(testTool{name: "Read", output: "fixture"}), core.Config{
		ProviderName: "anthropic",
		Model:        "claude-test",
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
