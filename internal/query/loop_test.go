package query

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"holycode/internal/api"
	"holycode/internal/core"
	"holycode/internal/inference"
	fakeprovider "holycode/internal/inference/fake"
	"holycode/internal/tools"
)

type testTool struct {
	name   string
	output string
}

func (t testTool) Name() string { return t.name }

func (t testTool) Definition() core.ToolDefinition {
	return core.ToolDefinition{
		Name:        t.name,
		Description: "test tool",
		InputSchema: json.RawMessage(`{"type":"object"}`),
	}
}

func (t testTool) Execute(_ context.Context, _ []byte, _ *tools.State) (tools.Result, error) {
	return tools.Result{Output: t.output}, nil
}

func TestRunStreamsTextInOrder(t *testing.T) {
	provider := &fakeprovider.Provider{
		RunTurnFn: func(_ context.Context, _ inference.ModelDescriptor, _ string, _ inference.TurnRequest) (<-chan inference.Event, error) {
			ch := make(chan inference.Event, 3)
			ch <- inference.Event{Type: inference.EventTypeTextDelta, TextDelta: "hello "}
			ch <- inference.Event{Type: inference.EventTypeTextDelta, TextDelta: "world"}
			ch <- inference.Event{Type: inference.EventTypeCompleted, StopReason: inference.StopReasonCompleted}
			close(ch)
			return ch, nil
		},
	}

	runtime := api.NewRuntime(inference.NewRegistry(provider))
	result, err := Run(context.Background(), runtime, tools.NewRegistry(), core.Config{
		ProviderName: "fake",
		Prompt:       "say hello",
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.Text != "hello world" {
		t.Fatalf("expected assembled text %q, got %q", "hello world", result.Text)
	}
}

func TestRunContinuesAfterToolCall(t *testing.T) {
	provider := &fakeprovider.Provider{
		RunTurnFn: func(_ context.Context, _ inference.ModelDescriptor, _ string, req inference.TurnRequest) (<-chan inference.Event, error) {
			ch := make(chan inference.Event, 2)
			if len(req.ToolResults) == 0 {
				payload, _ := json.Marshal(map[string]string{"path": "/tmp/ignored"})
				ch <- inference.Event{
					Type: inference.EventTypeToolCall,
					ToolCall: &inference.ToolCall{
						ID:    "tool-1",
						Name:  "Read",
						Input: payload,
					},
				}
				close(ch)
				return ch, nil
			}
			ch <- inference.Event{Type: inference.EventTypeTextDelta, TextDelta: "tool complete"}
			ch <- inference.Event{Type: inference.EventTypeCompleted, StopReason: inference.StopReasonCompleted}
			close(ch)
			return ch, nil
		},
	}

	runtime := api.NewRuntime(inference.NewRegistry(provider))
	result, err := Run(context.Background(), runtime, tools.NewRegistry(testTool{name: "Read", output: "fixture"}), core.Config{
		ProviderName: "fake",
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

func TestRunReturnsProviderErrors(t *testing.T) {
	provider := &fakeprovider.Provider{
		RunTurnFn: func(_ context.Context, _ inference.ModelDescriptor, _ string, _ inference.TurnRequest) (<-chan inference.Event, error) {
			ch := make(chan inference.Event, 1)
			ch <- inference.Event{Type: inference.EventTypeProviderError, ProviderErrText: "boom"}
			close(ch)
			return ch, nil
		},
	}

	runtime := api.NewRuntime(inference.NewRegistry(provider))
	_, err := Run(context.Background(), runtime, tools.NewRegistry(), core.Config{
		ProviderName: "fake",
		Prompt:       "fail",
	})
	if err == nil {
		t.Fatal("expected provider error, got nil")
	}
	var runtimeErr *core.RuntimeError
	if !errors.As(err, &runtimeErr) {
		t.Fatalf("expected RuntimeError, got %T", err)
	}
	if runtimeErr.Kind != core.ErrorKindProvider {
		t.Fatalf("expected provider error kind, got %q", runtimeErr.Kind)
	}
}

func TestRunDoesNotAdvertiseToolsWhenDescriptorLacksToolSupport(t *testing.T) {
	provider := &fakeprovider.Provider{
		DescriptorFn: func(model string, baseURL string) inference.ModelDescriptor {
			return inference.ModelDescriptor{
				ProviderName: "fake",
				Model:        model,
				BaseURL:      baseURL,
				Capabilities: inference.Capabilities{},
			}
		},
		RunTurnFn: func(_ context.Context, _ inference.ModelDescriptor, _ string, req inference.TurnRequest) (<-chan inference.Event, error) {
			if len(req.ToolDefinitions) != 0 {
				t.Fatalf("expected no advertised tools, got %d", len(req.ToolDefinitions))
			}
			ch := make(chan inference.Event, 2)
			ch <- inference.Event{Type: inference.EventTypeTextDelta, TextDelta: "text only"}
			ch <- inference.Event{Type: inference.EventTypeCompleted, StopReason: inference.StopReasonCompleted}
			close(ch)
			return ch, nil
		},
	}

	runtime := api.NewRuntime(inference.NewRegistry(provider))
	result, err := Run(context.Background(), runtime, tools.NewRegistry(testTool{name: "Read", output: "fixture"}), core.Config{
		ProviderName: "fake",
		Model:        "text-only",
		Prompt:       "hello",
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.Text != "text only" {
		t.Fatalf("expected final text %q, got %q", "text only", result.Text)
	}
}

func TestRunReturnsErrorWhenToolContinuationUnsupported(t *testing.T) {
	provider := &fakeprovider.Provider{
		DescriptorFn: func(model string, baseURL string) inference.ModelDescriptor {
			return inference.ModelDescriptor{
				ProviderName: "fake",
				Model:        model,
				BaseURL:      baseURL,
				Capabilities: inference.Capabilities{SupportsToolCalls: true},
			}
		},
		RunTurnFn: func(_ context.Context, _ inference.ModelDescriptor, _ string, req inference.TurnRequest) (<-chan inference.Event, error) {
			ch := make(chan inference.Event, 1)
			if len(req.ToolResults) == 0 {
				payload, _ := json.Marshal(map[string]string{"path": "/tmp/ignored"})
				ch <- inference.Event{
					Type: inference.EventTypeToolCall,
					ToolCall: &inference.ToolCall{
						ID:    "tool-1",
						Name:  "Read",
						Input: payload,
					},
				}
			}
			close(ch)
			return ch, nil
		},
	}

	runtime := api.NewRuntime(inference.NewRegistry(provider))
	_, err := Run(context.Background(), runtime, tools.NewRegistry(testTool{name: "Read", output: "fixture"}), core.Config{
		ProviderName: "fake",
		Model:        "tool-no-continuation",
		Prompt:       "hello",
	})
	if err == nil {
		t.Fatal("expected provider error, got nil")
	}
	var runtimeErr *core.RuntimeError
	if !errors.As(err, &runtimeErr) {
		t.Fatalf("expected RuntimeError, got %T", err)
	}
	if runtimeErr.Kind != core.ErrorKindProvider {
		t.Fatalf("expected provider error kind, got %q", runtimeErr.Kind)
	}
	if !strings.Contains(runtimeErr.Error(), "continuation") {
		t.Fatalf("expected continuation error, got %v", runtimeErr)
	}
}
