package tools

import (
	"context"
	"encoding/json"
	"os"

	"holycode/internal/core"
)

type EditTool struct{}

type editInput struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func (t EditTool) Name() string { return "Edit" }

func (t EditTool) Definition() core.ToolDefinition {
	return core.ToolDefinition{
		Name:        t.Name(),
		Description: "Create or replace file contents in the local workspace.",
		InputSchema: json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"},"content":{"type":"string"}},"required":["path","content"],"additionalProperties":false}`),
	}
}

func (t EditTool) Execute(_ context.Context, input []byte, state *State) (Result, error) {
	var req editInput
	if err := json.Unmarshal(input, &req); err != nil {
		return Result{}, err
	}
	if _, err := os.Stat(req.Path); err == nil {
		if err := requireRead(state, req.Path); err != nil {
			return Result{Output: err.Error(), IsError: true}, nil
		}
	}
	if err := os.WriteFile(req.Path, []byte(req.Content), 0o644); err != nil {
		return Result{}, err
	}
	return Result{Output: "ok"}, nil
}
