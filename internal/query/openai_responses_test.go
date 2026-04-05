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
	"holycode/internal/providers/openairesponses"
	"holycode/internal/tools"
)

func TestRunOpenAIResponsesExecutesToolAndContinuesTurn(t *testing.T) {
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
				"event: response.output_item.added",
				`data: {"type":"response.output_item.added","output_index":0,"item":{"type":"function_call","id":"fc_1","call_id":"call_1","name":"Read","arguments":""}}`,
				"",
				"event: response.function_call_arguments.delta",
				`data: {"type":"response.function_call_arguments.delta","item_id":"fc_1","output_index":0,"delta":"{\"path\":\"/tmp/demo\"}"}`,
				"",
				"event: response.function_call_arguments.done",
				`data: {"type":"response.function_call_arguments.done","item_id":"fc_1","output_index":0,"name":"Read","arguments":"{\"path\":\"/tmp/demo\"}"}`,
				"",
				"event: response.completed",
				`data: {"type":"response.completed","response":{"id":"resp_1","status":"completed","usage":{"input_tokens":10,"output_tokens":3}}}`,
				"",
			}, "\n"))
			return
		}

		input := payload["input"].([]any)
		if len(input) != 3 {
			t.Fatalf("expected 3 input items in continuation request, got %d", len(input))
		}
		_, _ = io.WriteString(w, strings.Join([]string{
			"event: response.output_item.added",
			`data: {"type":"response.output_item.added","output_index":0,"item":{"type":"message","id":"msg_2","status":"in_progress","role":"assistant","content":[]}}`,
			"",
			"event: response.output_text.delta",
			`data: {"type":"response.output_text.delta","item_id":"msg_2","output_index":0,"content_index":0,"delta":"tool complete"}`,
			"",
			"event: response.completed",
			`data: {"type":"response.completed","response":{"id":"resp_2","status":"completed","usage":{"input_tokens":12,"output_tokens":1}}}`,
			"",
		}, "\n"))
	}))
	defer server.Close()

	provider := &openairesponses.Provider{Client: server.Client()}
	runtime := api.NewRuntime(inference.NewRegistry(provider))
	result, err := Run(context.Background(), runtime, tools.NewRegistry(testTool{name: "Read", output: "fixture"}), core.Config{
		ProviderName: "openai-responses",
		Model:        "gpt-5",
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
