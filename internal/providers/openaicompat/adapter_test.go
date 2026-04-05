package openaicompat

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

func TestDescriptorDefaultsToTextOnlyCompatibilityTier(t *testing.T) {
	provider := &Provider{}
	desc := provider.Descriptor("compat-text", "https://compat.example/v1")
	if desc.Capabilities.SupportsToolCalls {
		t.Fatal("expected text-only descriptor to disable tool calls")
	}
	if desc.Capabilities.SupportsConversation {
		t.Fatal("expected text-only descriptor to disable continuation")
	}
}

func TestRunTurnBuildsChatCompletionsRequestAndStreamsText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("expected path %q, got %q", "/chat/completions", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("expected bearer auth header, got %q", got)
		}

		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if payload["model"] != "compat-text" {
			t.Fatalf("expected model %q, got %#v", "compat-text", payload["model"])
		}
		if payload["stream"] != true {
			t.Fatalf("expected stream=true, got %#v", payload["stream"])
		}
		if _, ok := payload["tools"]; ok {
			t.Fatalf("expected no tools for text-only tier, got %#v", payload["tools"])
		}

		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, strings.Join([]string{
			`data: {"choices":[{"delta":{"role":"assistant","content":"hello "}}]}`,
			"",
			`data: {"choices":[{"delta":{"content":"world"}}]}`,
			"",
			`data: {"choices":[{"finish_reason":"stop"}],"usage":{"prompt_tokens":5,"completion_tokens":2}}`,
			"",
			`data: [DONE]`,
			"",
		}, "\n"))
	}))
	defer server.Close()

	provider := &Provider{Client: server.Client()}
	events, err := provider.RunTurn(context.Background(), provider.Descriptor("compat-text", server.URL), "test-key", inference.TurnRequest{
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
			if event.Usage.InputTokens != 5 || event.Usage.OutputTokens != 2 {
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

func TestRunTurnMapsToolCallsAndContinuationForToolTier(t *testing.T) {
	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		w.Header().Set("Content-Type", "text/event-stream")
		if requestCount == 1 {
			tools, ok := payload["tools"].([]any)
			if !ok || len(tools) != 1 {
				t.Fatalf("expected one tool definition, got %#v", payload["tools"])
			}
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

		messages, ok := payload["messages"].([]any)
		if !ok || len(messages) != 3 {
			t.Fatalf("expected continuation messages, got %#v", payload["messages"])
		}
		_, _ = io.WriteString(w, strings.Join([]string{
			`data: {"choices":[{"delta":{"role":"assistant","content":"done"}}]}`,
			"",
			`data: {"choices":[{"finish_reason":"stop"}],"usage":{"prompt_tokens":9,"completion_tokens":1}}`,
			"",
			`data: [DONE]`,
			"",
		}, "\n"))
	}))
	defer server.Close()

	provider := &Provider{Client: server.Client()}
	desc := provider.Descriptor("compat-tools", server.URL)
	events, err := provider.RunTurn(context.Background(), desc, "test-key", inference.TurnRequest{
		Prompt: "read file",
		ToolDefinitions: []core.ToolDefinition{
			{Name: "Read", Description: "read file", InputSchema: json.RawMessage(`{"type":"object"}`)},
		},
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

func TestRunTurnNormalizesCompatibleErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"error":{"code":"invalid_api_key","message":"bad key"}}`)
	}))
	defer server.Close()

	provider := &Provider{Client: server.Client()}
	_, err := provider.RunTurn(context.Background(), provider.Descriptor("compat-text", server.URL), "bad-key", inference.TurnRequest{
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
