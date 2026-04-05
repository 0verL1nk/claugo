package query

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"holycode/internal/api"
	"holycode/internal/core"
	"holycode/internal/inference"
	"holycode/internal/tools"
)

type Result struct {
	Text        string
	Usage       inference.Usage
	ToolResults []inference.ToolResult
}

func Run(ctx context.Context, runtime *api.Runtime, registry *tools.Registry, cfg core.Config) (Result, error) {
	descriptor, err := runtime.ResolveDescriptor(cfg)
	if err != nil {
		return Result{}, &core.RuntimeError{
			Kind:    core.ErrorKindProvider,
			Message: "resolve provider descriptor",
			Err:     err,
		}
	}

	req := inference.TurnRequest{
		Prompt: cfg.Prompt,
		Model:  cfg.Model,
	}
	if descriptor.Capabilities.SupportsToolCalls && descriptor.Capabilities.SupportsConversation {
		req.ToolDefinitions = registry.Definitions()
	}
	var text strings.Builder
	var usage inference.Usage
	state := &tools.State{ReadFiles: map[string]bool{}}

	for {
		events, _, err := runtime.RunTurn(ctx, cfg, req)
		if err != nil {
			var runtimeErr *core.RuntimeError
			if errors.As(err, &runtimeErr) {
				return Result{}, runtimeErr
			}
			return Result{}, &core.RuntimeError{
				Kind:    core.ErrorKindProvider,
				Message: "run provider turn",
				Err:     err,
			}
		}

		currentTurn := inference.AssistantTurn{}
		var pendingToolCalls []inference.ToolCall
		for event := range events {
			switch event.Type {
			case inference.EventTypeTextDelta:
				text.WriteString(event.TextDelta)
				currentTurn.Text += event.TextDelta
				usage.InputTokens += event.Usage.InputTokens
				usage.OutputTokens += event.Usage.OutputTokens
			case inference.EventTypeToolCall:
				if event.ToolCall != nil {
					pendingToolCalls = append(pendingToolCalls, *event.ToolCall)
				}
			case inference.EventTypeProviderError:
				return Result{}, &core.RuntimeError{
					Kind:    core.ErrorKindProvider,
					Message: event.ProviderErrText,
				}
			case inference.EventTypeCompleted:
				usage.InputTokens += event.Usage.InputTokens
				usage.OutputTokens += event.Usage.OutputTokens
			}
		}

		if len(pendingToolCalls) == 0 {
			return Result{
				Text:        text.String(),
				Usage:       usage,
				ToolResults: req.ToolResults,
			}, nil
		}

		currentTurn.ToolCalls = append(currentTurn.ToolCalls, pendingToolCalls...)
		req.AssistantTurns = append(req.AssistantTurns, currentTurn)

		for _, pendingToolCall := range pendingToolCalls {
			tool, err := registry.Lookup(pendingToolCall.Name)
			if err != nil {
				return Result{}, &core.RuntimeError{
					Kind:    core.ErrorKindTool,
					Message: fmt.Sprintf("lookup tool %q", pendingToolCall.Name),
					Err:     err,
				}
			}
			result, err := tool.Execute(ctx, pendingToolCall.Input, state)
			if err != nil {
				return Result{}, &core.RuntimeError{
					Kind:    core.ErrorKindTool,
					Message: fmt.Sprintf("execute tool %q", pendingToolCall.Name),
					Err:     err,
				}
			}

			req.ToolResults = append(req.ToolResults, inference.ToolResult{
				ToolCallID: pendingToolCall.ID,
				Name:       pendingToolCall.Name,
				Output:     result.Output,
				IsError:    result.IsError,
			})
		}
	}
}
