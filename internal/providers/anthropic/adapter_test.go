package anthropic

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"holycode/internal/core"
	"holycode/internal/inference"
)

func TestRunTurnBuildsAnthropicRequestAndStreamsText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/messages" {
			t.Fatalf("expected path %q, got %q", "/v1/messages", r.URL.Path)
		}
		if got := r.Header.Get("x-api-key"); got != "test-key" {
			t.Fatalf("expected x-api-key header, got %q", got)
		}
		if got := r.Header.Get("anthropic-version"); got == "" {
			t.Fatal("expected anthropic-version header")
		}

		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if payload["model"] != "claude-test" {
			t.Fatalf("expected model %q, got %#v", "claude-test", payload["model"])
		}
		if payload["stream"] != true {
			t.Fatalf("expected stream=true, got %#v", payload["stream"])
		}
		messages, ok := payload["messages"].([]any)
		if !ok || len(messages) != 1 {
			t.Fatalf("expected one message, got %#v", payload["messages"])
		}
		tools, ok := payload["tools"].([]any)
		if !ok || len(tools) != 1 {
			t.Fatalf("expected one tool definition, got %#v", payload["tools"])
		}

		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, strings.Join([]string{
			"event: message_start",
			`data: {"type":"message_start","message":{"usage":{"input_tokens":11,"output_tokens":0}}}`,
			"",
			"event: content_block_start",
			`data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`,
			"",
			"event: content_block_delta",
			`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"hello "}}`,
			"",
			"event: content_block_delta",
			`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"world"}}`,
			"",
			"event: content_block_stop",
			`data: {"type":"content_block_stop","index":0}`,
			"",
			"event: message_delta",
			`data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":2}}`,
			"",
			"event: message_stop",
			`data: {"type":"message_stop"}`,
			"",
		}, "\n"))
	}))
	defer server.Close()

	provider := &Provider{Client: server.Client()}
	events, err := provider.RunTurn(context.Background(), provider.Descriptor("claude-test", server.URL), "test-key", inference.TurnRequest{
		Prompt: "say hello",
		ToolDefinitions: []core.ToolDefinition{
			{Name: "Read", Description: "read file", InputSchema: json.RawMessage(`{"type":"object"}`)},
		},
	})
	if err != nil {
		t.Fatalf("RunTurn returned error: %v", err)
	}

	var gotText strings.Builder
	var completed bool
	for event := range events {
		switch event.Type {
		case inference.EventTypeTextDelta:
			gotText.WriteString(event.TextDelta)
		case inference.EventTypeCompleted:
			completed = true
			if event.Usage.InputTokens != 11 || event.Usage.OutputTokens != 2 {
				t.Fatalf("unexpected usage: %#v", event.Usage)
			}
		}
	}

	if gotText.String() != "hello world" {
		t.Fatalf("expected streamed text %q, got %q", "hello world", gotText.String())
	}
	if !completed {
		t.Fatal("expected completed event")
	}
}

func TestRunTurnMapsAnthropicToolUseAndContinuationMessages(t *testing.T) {
	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if requestCount == 1 {
			w.Header().Set("Content-Type", "text/event-stream")
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

		messages, ok := payload["messages"].([]any)
		if !ok || len(messages) != 3 {
			t.Fatalf("expected continuation transcript, got %#v", payload["messages"])
		}

		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, strings.Join([]string{
			"event: content_block_start",
			`data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`,
			"",
			"event: content_block_delta",
			`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"done"}}`,
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

	provider := &Provider{Client: server.Client()}
	desc := provider.Descriptor("claude-test", server.URL)
	events, err := provider.RunTurn(context.Background(), desc, "test-key", inference.TurnRequest{
		Prompt: "read file",
	})
	if err != nil {
		t.Fatalf("first RunTurn returned error: %v", err)
	}

	var toolCall *inference.ToolCall
	for event := range events {
		if event.Type == inference.EventTypeToolCall {
			toolCall = event.ToolCall
		}
	}
	if toolCall == nil {
		t.Fatal("expected tool call event")
	}

	events, err = provider.RunTurn(context.Background(), desc, "test-key", inference.TurnRequest{
		Prompt: "read file",
		AssistantTurns: []inference.AssistantTurn{
			{
				ToolCalls: []inference.ToolCall{*toolCall},
			},
		},
		ToolResults: []inference.ToolResult{
			{ToolCallID: toolCall.ID, Name: toolCall.Name, Output: "fixture"},
		},
	})
	if err != nil {
		t.Fatalf("second RunTurn returned error: %v", err)
	}

	var sawDone bool
	for event := range events {
		if event.Type == inference.EventTypeTextDelta && event.TextDelta == "done" {
			sawDone = true
		}
	}
	if !sawDone {
		t.Fatal("expected final text delta after continuation")
	}
}

func TestRunTurnNormalizesAnthropicErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"type":"error","error":{"type":"authentication_error","message":"invalid key"}}`)
	}))
	defer server.Close()

	provider := &Provider{Client: server.Client()}
	_, err := provider.RunTurn(context.Background(), provider.Descriptor("claude-test", server.URL), "bad-key", inference.TurnRequest{
		Prompt: "hello",
	})
	if err == nil {
		t.Fatal("expected provider error")
	}
	var runtimeErr *core.RuntimeError
	if !errors.As(err, &runtimeErr) {
		t.Fatalf("expected RuntimeError, got %T", err)
	}
	if runtimeErr.Kind != core.ErrorKindAuth {
		t.Fatalf("expected auth error kind, got %q", runtimeErr.Kind)
	}
}
