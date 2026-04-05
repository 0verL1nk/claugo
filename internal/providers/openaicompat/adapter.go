package openaicompat

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"holycode/internal/core"
	"holycode/internal/inference"
)

type Provider struct {
	Client *http.Client
}

func (p *Provider) Name() string {
	return "openai-compatible"
}

func (p *Provider) Descriptor(model string, baseURL string) inference.ModelDescriptor {
	caps := inference.Capabilities{}
	switch model {
	case "compat-tools":
		caps = inference.Capabilities{
			SupportsToolCalls:        true,
			SupportsToolArgStreaming: true,
			SupportsConversation:     true,
		}
	case "":
		// Leave zero-value capabilities for invalid or unresolved models.
	default:
		// Text-only minimum contract.
	}
	return inference.ModelDescriptor{
		ProviderName: p.Name(),
		Model:        model,
		BaseURL:      baseURL,
		Capabilities: caps,
	}
}

func (p *Provider) RunTurn(ctx context.Context, descriptor inference.ModelDescriptor, apiKey string, req inference.TurnRequest) (<-chan inference.Event, error) {
	if apiKey == "" {
		return nil, &core.RuntimeError{
			Kind:    core.ErrorKindAuth,
			Message: "openai-compatible api key is required",
		}
	}
	if strings.TrimSpace(descriptor.BaseURL) == "" {
		return nil, &core.RuntimeError{
			Kind:    core.ErrorKindConfig,
			Message: "openai-compatible base URL is required",
		}
	}

	payload, err := buildRequestPayload(descriptor, req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(descriptor.BaseURL, "/")+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return nil, &core.RuntimeError{Kind: core.ErrorKindProvider, Message: "create openai-compatible request", Err: err}
	}
	httpReq.Header.Set("content-type", "application/json")
	httpReq.Header.Set("accept", "text/event-stream")
	httpReq.Header.Set("authorization", "Bearer "+apiKey)

	client := p.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, &core.RuntimeError{Kind: core.ErrorKindProvider, Message: "send openai-compatible request", Err: err}
	}
	if resp.StatusCode >= 300 {
		defer resp.Body.Close()
		return nil, normalizeHTTPError(resp)
	}

	events := make(chan inference.Event, 16)
	go func() {
		defer close(events)
		defer resp.Body.Close()
		streamEvents(ctx, resp.Body, events)
	}()
	return events, nil
}

func buildRequestPayload(descriptor inference.ModelDescriptor, req inference.TurnRequest) ([]byte, error) {
	messages, err := buildMessages(req)
	if err != nil {
		return nil, err
	}
	payload := map[string]any{
		"model":    descriptor.Model,
		"stream":   true,
		"messages": messages,
		"stream_options": map[string]any{
			"include_usage": true,
		},
	}
	if descriptor.Capabilities.SupportsToolCalls && len(req.ToolDefinitions) > 0 {
		tools, err := buildToolDefinitions(req.ToolDefinitions)
		if err != nil {
			return nil, err
		}
		payload["tools"] = tools
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &core.RuntimeError{Kind: core.ErrorKindProvider, Message: "marshal openai-compatible request", Err: err}
	}
	return body, nil
}

func buildToolDefinitions(defs []core.ToolDefinition) ([]map[string]any, error) {
	tools := make([]map[string]any, 0, len(defs))
	for _, def := range defs {
		var parameters any
		if len(def.InputSchema) == 0 {
			parameters = map[string]any{"type": "object", "properties": map[string]any{}}
		} else if err := json.Unmarshal(def.InputSchema, &parameters); err != nil {
			return nil, &core.RuntimeError{Kind: core.ErrorKindConfig, Message: fmt.Sprintf("invalid input schema for tool %q", def.Name), Err: err}
		}
		tools = append(tools, map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        def.Name,
				"description": def.Description,
				"parameters":  parameters,
			},
		})
	}
	return tools, nil
}

func buildMessages(req inference.TurnRequest) ([]map[string]any, error) {
	messages := []map[string]any{
		{
			"role":    "user",
			"content": req.Prompt,
		},
	}

	resultsByID := map[string]inference.ToolResult{}
	for _, result := range req.ToolResults {
		resultsByID[result.ToolCallID] = result
	}

	for _, turn := range req.AssistantTurns {
		assistant := map[string]any{"role": "assistant"}
		if strings.TrimSpace(turn.Text) != "" {
			assistant["content"] = turn.Text
		}
		if len(turn.ToolCalls) > 0 {
			toolCalls := make([]map[string]any, 0, len(turn.ToolCalls))
			for _, call := range turn.ToolCalls {
				arguments := "{}"
				if len(call.Input) > 0 {
					arguments = string(call.Input)
				}
				toolCalls = append(toolCalls, map[string]any{
					"id":   call.ID,
					"type": "function",
					"function": map[string]any{
						"name":      call.Name,
						"arguments": arguments,
					},
				})
			}
			assistant["tool_calls"] = toolCalls
		}
		messages = append(messages, assistant)

		for _, call := range turn.ToolCalls {
			result, ok := resultsByID[call.ID]
			if !ok {
				return nil, &core.RuntimeError{
					Kind:    core.ErrorKindProvider,
					Message: fmt.Sprintf("missing tool result for %q", call.ID),
				}
			}
			messages = append(messages, map[string]any{
				"role":         "tool",
				"tool_call_id": result.ToolCallID,
				"content":      result.Output,
			})
		}
	}

	return messages, nil
}

type streamState struct {
	usage     inference.Usage
	toolCalls map[int]*toolCallState
}

type toolCallState struct {
	ID        string
	Name      string
	Arguments strings.Builder
}

func streamEvents(ctx context.Context, body io.Reader, out chan<- inference.Event) {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	state := streamState{toolCalls: map[int]*toolCallState{}}
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" || !strings.HasPrefix(line, "data:") {
			continue
		}
		raw := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if raw == "[DONE]" {
			return
		}
		if err := handleChunk(raw, &state, out); err != nil {
			out <- inference.Event{
				Type:            inference.EventTypeProviderError,
				ProviderErrText: err.Error(),
				ProviderName:    "openai-compatible",
			}
			return
		}
	}
	if err := scanner.Err(); err != nil {
		out <- inference.Event{
			Type:            inference.EventTypeProviderError,
			ProviderErrText: err.Error(),
			ProviderName:    "openai-compatible",
		}
	}
}

func handleChunk(raw string, state *streamState, out chan<- inference.Event) error {
	var payload struct {
		Choices []struct {
			Delta struct {
				Content   string `json:"content"`
				ToolCalls []struct {
					Index    int    `json:"index"`
					ID       string `json:"id"`
					Type     string `json:"type"`
					Function struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls"`
			} `json:"delta"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return err
	}

	if payload.Usage.PromptTokens > 0 {
		state.usage.InputTokens = payload.Usage.PromptTokens
	}
	if payload.Usage.CompletionTokens > 0 {
		state.usage.OutputTokens = payload.Usage.CompletionTokens
	}

	for _, choice := range payload.Choices {
		if choice.Delta.Content != "" {
			out <- inference.Event{
				Type:         inference.EventTypeTextDelta,
				TextDelta:    choice.Delta.Content,
				ProviderName: "openai-compatible",
			}
		}

		for _, toolCall := range choice.Delta.ToolCalls {
			stateCall := state.toolCalls[toolCall.Index]
			if stateCall == nil {
				stateCall = &toolCallState{}
				state.toolCalls[toolCall.Index] = stateCall
			}
			if toolCall.ID != "" {
				stateCall.ID = toolCall.ID
			}
			if toolCall.Function.Name != "" {
				stateCall.Name = toolCall.Function.Name
			}
			if toolCall.Function.Arguments != "" {
				stateCall.Arguments.WriteString(toolCall.Function.Arguments)
			}
		}

		switch choice.FinishReason {
		case "tool_calls":
			indexes := make([]int, 0, len(state.toolCalls))
			for index := range state.toolCalls {
				indexes = append(indexes, index)
			}
			for _, index := range indexes {
				call := state.toolCalls[index]
				if call == nil {
					continue
				}
				args := call.Arguments.String()
				if strings.TrimSpace(args) == "" {
					args = "{}"
				}
				out <- inference.Event{
					Type: inference.EventTypeToolCall,
					ToolCall: &inference.ToolCall{
						ID:    call.ID,
						Name:  call.Name,
						Input: []byte(args),
					},
					ProviderName: "openai-compatible",
				}
				delete(state.toolCalls, index)
			}
			out <- inference.Event{
				Type:         inference.EventTypeCompleted,
				StopReason:   inference.StopReasonToolCall,
				Usage:        state.usage,
				ProviderName: "openai-compatible",
			}
		case "stop":
			out <- inference.Event{
				Type:         inference.EventTypeCompleted,
				StopReason:   inference.StopReasonCompleted,
				Usage:        state.usage,
				ProviderName: "openai-compatible",
			}
		}
	}

	return nil
}

func normalizeHTTPError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	var payload struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
			Type    string `json:"type"`
		} `json:"error"`
	}
	_ = json.Unmarshal(body, &payload)

	kind := core.ErrorKindProvider
	switch payload.Error.Code {
	case "invalid_api_key":
		kind = core.ErrorKindAuth
	case "rate_limit_exceeded":
		kind = core.ErrorKindRateLimit
	}
	switch payload.Error.Type {
	case "rate_limit_error":
		kind = core.ErrorKindRateLimit
	}

	message := strings.TrimSpace(payload.Error.Message)
	if message == "" {
		message = strings.TrimSpace(string(body))
	}
	if message == "" {
		message = resp.Status
	}

	return &core.RuntimeError{
		Kind:    kind,
		Message: fmt.Sprintf("openai-compatible API error (%d)", resp.StatusCode),
		Err:     fmt.Errorf("%s", message),
	}
}
