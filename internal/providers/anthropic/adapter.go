package anthropic

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

const defaultBaseURL = "https://api.anthropic.com"

type Provider struct {
	Client *http.Client
}

func (p *Provider) Name() string {
	return "anthropic"
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
			Message: "anthropic api key is required",
		}
	}
	payload, err := buildRequestPayload(descriptor, req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(descriptor.BaseURL, "/")+"/v1/messages", bytes.NewReader(payload))
	if err != nil {
		return nil, &core.RuntimeError{Kind: core.ErrorKindProvider, Message: "create anthropic request", Err: err}
	}
	httpReq.Header.Set("content-type", "application/json")
	httpReq.Header.Set("accept", "text/event-stream")
	httpReq.Header.Set("x-api-key", apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	client := p.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, &core.RuntimeError{Kind: core.ErrorKindProvider, Message: "send anthropic request", Err: err}
	}
	if resp.StatusCode >= 300 {
		defer resp.Body.Close()
		return nil, normalizeHTTPError(resp)
	}

	events := make(chan inference.Event, 16)
	go func() {
		defer close(events)
		defer resp.Body.Close()
		streamAnthropicEvents(ctx, resp.Body, events)
	}()
	return events, nil
}

func buildRequestPayload(descriptor inference.ModelDescriptor, req inference.TurnRequest) ([]byte, error) {
	messages, err := buildMessages(req)
	if err != nil {
		return nil, err
	}
	payload := map[string]any{
		"model":      descriptor.Model,
		"max_tokens": 1024,
		"stream":     true,
		"messages":   messages,
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
		return nil, &core.RuntimeError{Kind: core.ErrorKindProvider, Message: "marshal anthropic request", Err: err}
	}
	return body, nil
}

func buildToolDefinitions(defs []core.ToolDefinition) ([]map[string]any, error) {
	tools := make([]map[string]any, 0, len(defs))
	for _, def := range defs {
		var inputSchema any
		if len(def.InputSchema) == 0 {
			inputSchema = map[string]any{"type": "object", "properties": map[string]any{}}
		} else if err := json.Unmarshal(def.InputSchema, &inputSchema); err != nil {
			return nil, &core.RuntimeError{Kind: core.ErrorKindConfig, Message: fmt.Sprintf("invalid input schema for tool %q", def.Name), Err: err}
		}
		tools = append(tools, map[string]any{
			"name":         def.Name,
			"description":  def.Description,
			"input_schema": inputSchema,
		})
	}
	return tools, nil
}

func buildMessages(req inference.TurnRequest) ([]map[string]any, error) {
	messages := []map[string]any{
		{
			"role": "user",
			"content": []map[string]any{
				{"type": "text", "text": req.Prompt},
			},
		},
	}
	resultsByID := map[string]inference.ToolResult{}
	for _, result := range req.ToolResults {
		resultsByID[result.ToolCallID] = result
	}

	for _, turn := range req.AssistantTurns {
		assistantContent := make([]map[string]any, 0, len(turn.ToolCalls)+1)
		if turn.Text != "" {
			assistantContent = append(assistantContent, map[string]any{"type": "text", "text": turn.Text})
		}
		for _, call := range turn.ToolCalls {
			input := map[string]any{}
			if len(call.Input) > 0 {
				if err := json.Unmarshal(call.Input, &input); err != nil {
					return nil, &core.RuntimeError{Kind: core.ErrorKindConfig, Message: fmt.Sprintf("decode tool call input for %q", call.Name), Err: err}
				}
			}
			assistantContent = append(assistantContent, map[string]any{
				"type":  "tool_use",
				"id":    call.ID,
				"name":  call.Name,
				"input": input,
			})
		}
		if len(assistantContent) > 0 {
			messages = append(messages, map[string]any{
				"role":    "assistant",
				"content": assistantContent,
			})
		}

		userContent := make([]map[string]any, 0, len(turn.ToolCalls))
		for _, call := range turn.ToolCalls {
			result, ok := resultsByID[call.ID]
			if !ok {
				return nil, &core.RuntimeError{
					Kind:    core.ErrorKindProvider,
					Message: fmt.Sprintf("missing tool result for %q", call.ID),
				}
			}
			userContent = append(userContent, map[string]any{
				"type":        "tool_result",
				"tool_use_id": result.ToolCallID,
				"content":     result.Output,
				"is_error":    result.IsError,
			})
		}
		if len(userContent) > 0 {
			messages = append(messages, map[string]any{
				"role":    "user",
				"content": userContent,
			})
		}
	}

	return messages, nil
}

type streamState struct {
	usage  inference.Usage
	blocks map[int]*contentBlockState
}

type contentBlockState struct {
	Type  string
	ID    string
	Name  string
	Input strings.Builder
}

func streamAnthropicEvents(ctx context.Context, body io.Reader, out chan<- inference.Event) {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	state := streamState{blocks: map[int]*contentBlockState{}}
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
				ProviderName:    "anthropic",
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
			ProviderName:    "anthropic",
		}
	}
}

func handleFrame(frameType string, raw string, state *streamState, out chan<- inference.Event) error {
	var payload map[string]json.RawMessage
	if raw == "" || raw == "[DONE]" {
		return nil
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return err
	}

	switch frameType {
	case "message_start":
		var wrapper struct {
			Message struct {
				Usage struct {
					InputTokens  int `json:"input_tokens"`
					OutputTokens int `json:"output_tokens"`
				} `json:"usage"`
			} `json:"message"`
		}
		if err := json.Unmarshal([]byte(raw), &wrapper); err != nil {
			return err
		}
		state.usage.InputTokens = wrapper.Message.Usage.InputTokens
		state.usage.OutputTokens = wrapper.Message.Usage.OutputTokens
	case "content_block_start":
		var wrapper struct {
			Index        int `json:"index"`
			ContentBlock struct {
				Type  string          `json:"type"`
				ID    string          `json:"id"`
				Name  string          `json:"name"`
				Input json.RawMessage `json:"input"`
			} `json:"content_block"`
		}
		if err := json.Unmarshal([]byte(raw), &wrapper); err != nil {
			return err
		}
		block := &contentBlockState{
			Type: wrapper.ContentBlock.Type,
			ID:   wrapper.ContentBlock.ID,
			Name: wrapper.ContentBlock.Name,
		}
		if len(wrapper.ContentBlock.Input) > 0 && string(wrapper.ContentBlock.Input) != "null" && string(wrapper.ContentBlock.Input) != "{}" {
			block.Input.Write(wrapper.ContentBlock.Input)
		}
		state.blocks[wrapper.Index] = block
	case "content_block_delta":
		var wrapper struct {
			Index int `json:"index"`
			Delta struct {
				Type        string `json:"type"`
				Text        string `json:"text"`
				PartialJSON string `json:"partial_json"`
			} `json:"delta"`
		}
		if err := json.Unmarshal([]byte(raw), &wrapper); err != nil {
			return err
		}
		switch wrapper.Delta.Type {
		case "text_delta":
			out <- inference.Event{Type: inference.EventTypeTextDelta, TextDelta: wrapper.Delta.Text, ProviderName: "anthropic"}
		case "input_json_delta":
			if block := state.blocks[wrapper.Index]; block != nil {
				block.Input.WriteString(wrapper.Delta.PartialJSON)
			}
		}
	case "content_block_stop":
		var wrapper struct {
			Index int `json:"index"`
		}
		if err := json.Unmarshal([]byte(raw), &wrapper); err != nil {
			return err
		}
		block := state.blocks[wrapper.Index]
		if block == nil {
			return nil
		}
		if block.Type == "tool_use" {
			input := block.Input.String()
			if strings.TrimSpace(input) == "" {
				input = "{}"
			}
			out <- inference.Event{
				Type: inference.EventTypeToolCall,
				ToolCall: &inference.ToolCall{
					ID:    block.ID,
					Name:  block.Name,
					Input: []byte(input),
				},
				ProviderName: "anthropic",
			}
		}
		delete(state.blocks, wrapper.Index)
	case "message_delta":
		var wrapper struct {
			Delta struct {
				StopReason string `json:"stop_reason"`
			} `json:"delta"`
			Usage struct {
				OutputTokens int `json:"output_tokens"`
			} `json:"usage"`
		}
		if err := json.Unmarshal([]byte(raw), &wrapper); err != nil {
			return err
		}
		state.usage.OutputTokens = wrapper.Usage.OutputTokens
	case "message_stop":
		out <- inference.Event{
			Type:         inference.EventTypeCompleted,
			StopReason:   inference.StopReasonCompleted,
			Usage:        state.usage,
			ProviderName: "anthropic",
		}
	}
	return nil
}

func normalizeHTTPError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	var payload struct {
		Error struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"error"`
	}
	_ = json.Unmarshal(body, &payload)

	kind := core.ErrorKindProvider
	switch payload.Error.Type {
	case "authentication_error":
		kind = core.ErrorKindAuth
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
		Message: fmt.Sprintf("anthropic API error (%d)", resp.StatusCode),
		Err:     fmt.Errorf("%s", message),
	}
}
