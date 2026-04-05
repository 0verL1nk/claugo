package conformance

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"holycode/internal/core"
	"holycode/internal/inference"
	"holycode/internal/providers/anthropic"
	"holycode/internal/providers/openaicompat"
	"holycode/internal/providers/openairesponses"
)

func collectEvents(t *testing.T, events <-chan inference.Event) []inference.Event {
	t.Helper()
	var collected []inference.Event
	for event := range events {
		collected = append(collected, event)
	}
	return collected
}

func assertTextAndCompletion(t *testing.T, events []inference.Event, wantText string) {
	t.Helper()
	var gotText strings.Builder
	var sawCompleted bool
	for _, event := range events {
		switch event.Type {
		case inference.EventTypeTextDelta:
			gotText.WriteString(event.TextDelta)
		case inference.EventTypeCompleted:
			sawCompleted = true
		}
	}
	if gotText.String() != wantText {
		t.Fatalf("expected text %q, got %q", wantText, gotText.String())
	}
	if !sawCompleted {
		t.Fatal("expected completed event")
	}
}

func assertToolCall(t *testing.T, events []inference.Event, wantName string) {
	t.Helper()
	for _, event := range events {
		if event.Type == inference.EventTypeToolCall {
			if event.ToolCall == nil {
				t.Fatal("expected tool call payload")
			}
			if event.ToolCall.Name != wantName {
				t.Fatalf("expected tool call %q, got %q", wantName, event.ToolCall.Name)
			}
			return
		}
	}
	t.Fatalf("expected tool call %q", wantName)
}

func TestProvidersMapTextOutputToSharedEvents(t *testing.T) {
	t.Run("anthropic", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = io.WriteString(w, strings.Join([]string{
				"event: content_block_start",
				`data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`,
				"",
				"event: content_block_delta",
				`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"hello anthropic"}}`,
				"",
				"event: message_stop",
				`data: {"type":"message_stop"}`,
				"",
			}, "\n"))
		}))
		defer server.Close()

		provider := &anthropic.Provider{Client: server.Client()}
		events, err := provider.RunTurn(context.Background(), provider.Descriptor("claude-test", server.URL), "test-key", inference.TurnRequest{Prompt: "hello"})
		if err != nil {
			t.Fatalf("RunTurn returned error: %v", err)
		}
		assertTextAndCompletion(t, collectEvents(t, events), "hello anthropic")
	})

	t.Run("openai-responses", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = io.WriteString(w, strings.Join([]string{
				"event: response.output_text.delta",
				`data: {"type":"response.output_text.delta","delta":"hello responses"}`,
				"",
				"event: response.completed",
				`data: {"type":"response.completed","response":{"usage":{"input_tokens":5,"output_tokens":2}}}`,
				"",
			}, "\n"))
		}))
		defer server.Close()

		provider := &openairesponses.Provider{Client: server.Client()}
		events, err := provider.RunTurn(context.Background(), provider.Descriptor("gpt-5", server.URL), "test-key", inference.TurnRequest{Prompt: "hello"})
		if err != nil {
			t.Fatalf("RunTurn returned error: %v", err)
		}
		assertTextAndCompletion(t, collectEvents(t, events), "hello responses")
	})

	t.Run("openai-compatible", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = io.WriteString(w, strings.Join([]string{
				`data: {"choices":[{"delta":{"role":"assistant","content":"hello compatible"}}]}`,
				"",
				`data: {"choices":[{"finish_reason":"stop"}],"usage":{"prompt_tokens":4,"completion_tokens":2}}`,
				"",
				`data: [DONE]`,
				"",
			}, "\n"))
		}))
		defer server.Close()

		provider := &openaicompat.Provider{Client: server.Client()}
		events, err := provider.RunTurn(context.Background(), provider.Descriptor("compat-text", server.URL), "test-key", inference.TurnRequest{Prompt: "hello"})
		if err != nil {
			t.Fatalf("RunTurn returned error: %v", err)
		}
		assertTextAndCompletion(t, collectEvents(t, events), "hello compatible")
	})
}

func TestProvidersMapToolCallsToSharedEvents(t *testing.T) {
	toolDefs := []core.ToolDefinition{
		{Name: "Read", Description: "read file", InputSchema: json.RawMessage(`{"type":"object"}`)},
	}

	t.Run("anthropic", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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
				"event: message_stop",
				`data: {"type":"message_stop"}`,
				"",
			}, "\n"))
		}))
		defer server.Close()

		provider := &anthropic.Provider{Client: server.Client()}
		events, err := provider.RunTurn(context.Background(), provider.Descriptor("claude-test", server.URL), "test-key", inference.TurnRequest{
			Prompt:          "read file",
			ToolDefinitions: toolDefs,
		})
		if err != nil {
			t.Fatalf("RunTurn returned error: %v", err)
		}
		assertToolCall(t, collectEvents(t, events), "Read")
	})

	t.Run("openai-responses", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = io.WriteString(w, strings.Join([]string{
				"event: response.output_item.added",
				`data: {"type":"response.output_item.added","output_index":0,"item":{"type":"function_call","id":"fc_1","call_id":"call_1","name":"Read","arguments":""}}`,
				"",
				"event: response.function_call_arguments.done",
				`data: {"type":"response.function_call_arguments.done","item_id":"fc_1","name":"Read","arguments":"{\"path\":\"/tmp/demo\"}"}`,
				"",
				"event: response.completed",
				`data: {"type":"response.completed","response":{"usage":{"input_tokens":6,"output_tokens":2}}}`,
				"",
			}, "\n"))
		}))
		defer server.Close()

		provider := &openairesponses.Provider{Client: server.Client()}
		events, err := provider.RunTurn(context.Background(), provider.Descriptor("gpt-5", server.URL), "test-key", inference.TurnRequest{
			Prompt:          "read file",
			ToolDefinitions: toolDefs,
		})
		if err != nil {
			t.Fatalf("RunTurn returned error: %v", err)
		}
		assertToolCall(t, collectEvents(t, events), "Read")
	})

	t.Run("openai-compatible", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = io.WriteString(w, strings.Join([]string{
				`data: {"choices":[{"delta":{"tool_calls":[{"index":0,"id":"call_1","type":"function","function":{"name":"Read","arguments":"{\"path\":\"/tmp/demo\"}"}}]}}]}`,
				"",
				`data: {"choices":[{"finish_reason":"tool_calls"}],"usage":{"prompt_tokens":7,"completion_tokens":3}}`,
				"",
				`data: [DONE]`,
				"",
			}, "\n"))
		}))
		defer server.Close()

		provider := &openaicompat.Provider{Client: server.Client()}
		events, err := provider.RunTurn(context.Background(), provider.Descriptor("compat-tools", server.URL), "test-key", inference.TurnRequest{
			Prompt:          "read file",
			ToolDefinitions: toolDefs,
		})
		if err != nil {
			t.Fatalf("RunTurn returned error: %v", err)
		}
		assertToolCall(t, collectEvents(t, events), "Read")
	})
}
