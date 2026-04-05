package openairesponses

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

const defaultBaseURL = "https://api.openai.com/v1"

type Provider struct {
	Client *http.Client
}

func (p *Provider) Name() string {
	return "openai-responses"
}

func (p *Provider) Descriptor(model string, baseURL string) inference.ModelDescriptor {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	caps := inference.Capabilities{}
	if model != "" {
		caps = inference.Capabilities{
			SupportsToolCalls:        true,
			SupportsToolArgStreaming: true,
			SupportsConversation:     true,
		}
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
			Message: "openai-responses api key is required",
		}
	}

	payload, err := buildRequestPayload(descriptor, req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(descriptor.BaseURL, "/")+"/responses", bytes.NewReader(payload))
	if err != nil {
		return nil, &core.RuntimeError{Kind: core.ErrorKindProvider, Message: "create openai-responses request", Err: err}
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
		return nil, &core.RuntimeError{Kind: core.ErrorKindProvider, Message: "send openai-responses request", Err: err}
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
	input, err := buildInputItems(req)
	if err != nil {
		return nil, err
	}

	payload := map[string]any{
		"model":  descriptor.Model,
		"stream": true,
		"input":  input,
	}
	if len(req.ToolDefinitions) > 0 {
		tools, err := buildToolDefinitions(req.ToolDefinitions)
		if err != nil {
			return nil, err
		}
		payload["tools"] = tools
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &core.RuntimeError{Kind: core.ErrorKindProvider, Message: "marshal openai-responses request", Err: err}
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
			"type":        "function",
			"name":        def.Name,
			"description": def.Description,
			"parameters":  parameters,
			"strict":      true,
		})
	}
	return tools, nil
}

func buildInputItems(req inference.TurnRequest) ([]map[string]any, error) {
	items := []map[string]any{
		{
			"type": "message",
			"role": "user",
			"content": []map[string]any{
				{"type": "input_text", "text": req.Prompt},
			},
		},
	}

	resultsByID := map[string]inference.ToolResult{}
	for _, result := range req.ToolResults {
		resultsByID[result.ToolCallID] = result
	}

	for _, turn := range req.AssistantTurns {
		if strings.TrimSpace(turn.Text) != "" {
			items = append(items, map[string]any{
				"type": "message",
				"role": "assistant",
				"content": []map[string]any{
					{"type": "output_text", "text": turn.Text},
				},
			})
		}

		for _, call := range turn.ToolCalls {
			arguments := "{}"
			if len(call.Input) > 0 {
				arguments = string(call.Input)
			}
			items = append(items, map[string]any{
				"type":      "function_call",
				"call_id":   call.ID,
				"name":      call.Name,
				"arguments": arguments,
			})

			result, ok := resultsByID[call.ID]
			if !ok {
				return nil, &core.RuntimeError{
					Kind:    core.ErrorKindProvider,
					Message: fmt.Sprintf("missing tool result for %q", call.ID),
				}
			}
			items = append(items, map[string]any{
				"type":    "function_call_output",
				"call_id": result.ToolCallID,
				"output":  result.Output,
			})
		}
	}

	return items, nil
}

type streamState struct {
	usage         inference.Usage
	functionCalls map[string]*functionCallState
}

type functionCallState struct {
	ItemID    string
	CallID    string
	Name      string
	Arguments strings.Builder
}

func streamEvents(ctx context.Context, body io.Reader, out chan<- inference.Event) {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	state := streamState{functionCalls: map[string]*functionCallState{}}
	var frameType string
	var dataLines []string
	flush := func() bool {
		if len(dataLines) == 0 {
			return true
		}
		if err := handleFrame(frameType, strings.Join(dataLines, "\n"), &state, out); err != nil {
			out <- inference.Event{
				Type:            inference.EventTypeProviderError,
				ProviderErrText: err.Error(),
				ProviderName:    "openai-responses",
			}
			return false
		}
		frameType = ""
		dataLines = nil
		return true
	}

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		line := scanner.Text()
		if line == "" {
			if !flush() {
				return
			}
			continue
		}
		switch {
		case strings.HasPrefix(line, "event:"):
			frameType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		case strings.HasPrefix(line, "data:"):
			dataLines = append(dataLines, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}
	}
	if !flush() {
		return
	}
	if err := scanner.Err(); err != nil {
		out <- inference.Event{
			Type:            inference.EventTypeProviderError,
			ProviderErrText: err.Error(),
			ProviderName:    "openai-responses",
		}
	}
}

func handleFrame(frameType string, raw string, state *streamState, out chan<- inference.Event) error {
	if raw == "" || raw == "[DONE]" {
		return nil
	}

	switch frameType {
	case "response.output_item.added":
		var wrapper struct {
			Item struct {
				Type      string `json:"type"`
				ID        string `json:"id"`
				CallID    string `json:"call_id"`
				Name      string `json:"name"`
				Arguments string `json:"arguments"`
			} `json:"item"`
		}
		if err := json.Unmarshal([]byte(raw), &wrapper); err != nil {
			return err
		}
		if wrapper.Item.Type == "function_call" {
			call := &functionCallState{
				ItemID: wrapper.Item.ID,
				CallID: wrapper.Item.CallID,
				Name:   wrapper.Item.Name,
			}
			call.Arguments.WriteString(wrapper.Item.Arguments)
			state.functionCalls[wrapper.Item.ID] = call
		}
	case "response.output_text.delta":
		var wrapper struct {
			Delta string `json:"delta"`
		}
		if err := json.Unmarshal([]byte(raw), &wrapper); err != nil {
			return err
		}
		out <- inference.Event{
			Type:         inference.EventTypeTextDelta,
			TextDelta:    wrapper.Delta,
			ProviderName: "openai-responses",
		}
	case "response.function_call_arguments.delta":
		var wrapper struct {
			ItemID string `json:"item_id"`
			Delta  string `json:"delta"`
		}
		if err := json.Unmarshal([]byte(raw), &wrapper); err != nil {
			return err
		}
		call := state.functionCalls[wrapper.ItemID]
		if call == nil {
			call = &functionCallState{ItemID: wrapper.ItemID}
			state.functionCalls[wrapper.ItemID] = call
		}
		call.Arguments.WriteString(wrapper.Delta)
	case "response.function_call_arguments.done":
		var wrapper struct {
			ItemID    string `json:"item_id"`
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		}
		if err := json.Unmarshal([]byte(raw), &wrapper); err != nil {
			return err
		}
		call := state.functionCalls[wrapper.ItemID]
		if call == nil {
			call = &functionCallState{ItemID: wrapper.ItemID}
		}
		if wrapper.Name != "" {
			call.Name = wrapper.Name
		}
		args := wrapper.Arguments
		if strings.TrimSpace(args) == "" {
			args = call.Arguments.String()
		}
		if strings.TrimSpace(args) == "" {
			args = "{}"
		}
		out <- inference.Event{
			Type: inference.EventTypeToolCall,
			ToolCall: &inference.ToolCall{
				ID:    call.CallID,
				Name:  call.Name,
				Input: []byte(args),
			},
			ProviderName: "openai-responses",
		}
		delete(state.functionCalls, wrapper.ItemID)
	case "response.completed":
		var wrapper struct {
			Response struct {
				Usage struct {
					InputTokens  int `json:"input_tokens"`
					OutputTokens int `json:"output_tokens"`
				} `json:"usage"`
			} `json:"response"`
		}
		if err := json.Unmarshal([]byte(raw), &wrapper); err != nil {
			return err
		}
		state.usage.InputTokens = wrapper.Response.Usage.InputTokens
		state.usage.OutputTokens = wrapper.Response.Usage.OutputTokens
		out <- inference.Event{
			Type:         inference.EventTypeCompleted,
			StopReason:   inference.StopReasonCompleted,
			Usage:        state.usage,
			ProviderName: "openai-responses",
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
	case "invalid_request_error":
		if payload.Error.Code == "invalid_api_key" {
			kind = core.ErrorKindAuth
		}
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
		Message: fmt.Sprintf("openai-responses API error (%d)", resp.StatusCode),
		Err:     fmt.Errorf("%s", message),
	}
}
