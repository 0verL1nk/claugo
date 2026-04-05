package openairesponses

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

func TestRunTurnBuildsResponsesRequestAndStreamsText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/responses" {
			t.Fatalf("expected path %q, got %q", "/responses", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("expected bearer auth header, got %q", got)
		}

		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if payload["model"] != "gpt-5" {
			t.Fatalf("expected model %q, got %#v", "gpt-5", payload["model"])
		}
		if payload["stream"] != true {
			t.Fatalf("expected stream=true, got %#v", payload["stream"])
		}
		input, ok := payload["input"].([]any)
		if !ok || len(input) != 1 {
			t.Fatalf("expected one input item, got %#v", payload["input"])
		}
		tools, ok := payload["tools"].([]any)
		if !ok || len(tools) != 1 {
			t.Fatalf("expected one tool definition, got %#v", payload["tools"])
		}

		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, strings.Join([]string{
			"event: response.output_item.added",
			`data: {"type":"response.output_item.added","output_index":0,"item":{"type":"message","id":"msg_1","status":"in_progress","role":"assistant","content":[]}}`,
			"",
			"event: response.output_text.delta",
			`data: {"type":"response.output_text.delta","item_id":"msg_1","output_index":0,"content_index":0,"delta":"hello "}`,
			"",
			"event: response.output_text.delta",
			`data: {"type":"response.output_text.delta","item_id":"msg_1","output_index":0,"content_index":0,"delta":"world"}`,
			"",
			"event: response.completed",
			`data: {"type":"response.completed","response":{"id":"resp_1","status":"completed","usage":{"input_tokens":11,"output_tokens":2}}}`,
			"",
		}, "\n"))
	}))
	defer server.Close()

	provider := &Provider{Client: server.Client()}
	events, err := provider.RunTurn(context.Background(), provider.Descriptor("gpt-5", server.URL), "test-key", inference.TurnRequest{
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

func TestRunTurnMapsFunctionCallAndContinuationInput(t *testing.T) {
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

		input, ok := payload["input"].([]any)
		if !ok || len(input) != 3 {
			t.Fatalf("expected continuation input items, got %#v", payload["input"])
		}
		last, ok := input[2].(map[string]any)
		if !ok || last["type"] != "function_call_output" {
			t.Fatalf("expected function_call_output item, got %#v", input[2])
		}
		if last["call_id"] != "call_1" {
			t.Fatalf("expected call_id %q, got %#v", "call_1", last["call_id"])
		}

		_, _ = io.WriteString(w, strings.Join([]string{
			"event: response.output_item.added",
			`data: {"type":"response.output_item.added","output_index":0,"item":{"type":"message","id":"msg_2","status":"in_progress","role":"assistant","content":[]}}`,
			"",
			"event: response.output_text.delta",
			`data: {"type":"response.output_text.delta","item_id":"msg_2","output_index":0,"content_index":0,"delta":"done"}`,
			"",
			"event: response.completed",
			`data: {"type":"response.completed","response":{"id":"resp_2","status":"completed","usage":{"input_tokens":12,"output_tokens":1}}}`,
			"",
		}, "\n"))
	}))
	defer server.Close()

	provider := &Provider{Client: server.Client()}
	desc := provider.Descriptor("gpt-5", server.URL)
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

func TestRunTurnNormalizesResponsesErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"error":{"code":"invalid_api_key","message":"bad key"}}`)
	}))
	defer server.Close()

	provider := &Provider{Client: server.Client()}
	_, err := provider.RunTurn(context.Background(), provider.Descriptor("gpt-5", server.URL), "bad-key", inference.TurnRequest{
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
