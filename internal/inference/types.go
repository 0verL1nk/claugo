package inference

import "holycode/internal/core"

type StopReason string

const (
	StopReasonCompleted StopReason = "completed"
	StopReasonToolCall  StopReason = "tool_call"
	StopReasonError     StopReason = "error"
)

type Usage struct {
	InputTokens  int
	OutputTokens int
}

type ToolCall struct {
	ID    string
	Name  string
	Input []byte
}

type ToolResult struct {
	ToolCallID string
	Name       string
	Output     string
	IsError    bool
}

type AssistantTurn struct {
	Text      string
	ToolCalls []ToolCall
}

type TurnRequest struct {
	Prompt          string
	Model           string
	Messages        []string
	AssistantTurns  []AssistantTurn
	ToolDefinitions []core.ToolDefinition
	ToolResults     []ToolResult
}

type EventType string

const (
	EventTypeTextDelta     EventType = "text_delta"
	EventTypeToolCall      EventType = "tool_call"
	EventTypeCompleted     EventType = "completed"
	EventTypeProviderError EventType = "provider_error"
)

type Event struct {
	Type            EventType
	TextDelta       string
	ToolCall        *ToolCall
	StopReason      StopReason
	Usage           Usage
	ProviderName    string
	ProviderMeta    map[string]string
	ProviderErrText string
}
